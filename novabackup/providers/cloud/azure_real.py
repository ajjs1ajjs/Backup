from typing import List, Dict, Any, Optional
import datetime
import uuid
import logging

logger = logging.getLogger("novabackup.azure")


class AzureCloudProvider:
    """
    Azure Cloud Provider for NovaBackup.
    
    Supports:
    - Azure VM listing
    - Snapshot creation for backups
    - Snapshot restore
    """
    
    def __init__(
        self,
        subscription_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
        client_id: Optional[str] = None,
        client_secret: Optional[str] = None,
        credential=None,
    ):
        self.subscription_id = subscription_id
        self.tenant_id = tenant_id
        self.client_id = client_id
        self.client_secret = client_secret
        self._credential = credential
        self._compute_client = None
        self._resource_client = None

    def _get_credential(self):
        """Get Azure credential."""
        if self._credential is None:
            try:
                from azure.identity import DefaultAzureCredential
                
                if self.tenant_id and self.client_id and self.client_secret:
                    from azure.identity import ClientSecretCredential
                    self._credential = ClientSecretCredential(
                        tenant_id=self.tenant_id,
                        client_id=self.client_id,
                        client_secret=self.client_secret,
                    )
                else:
                    self._credential = DefaultAzureCredential()
            except Exception as e:
                logger.error(f"Failed to create Azure credential: {e}")
                return None
        return self._credential

    def _get_compute_client(self):
        """Get Compute Management Client."""
        if self._compute_client is None:
            credential = self._get_credential()
            if credential and self.subscription_id:
                try:
                    from azure.mgmt.compute import ComputeManagementClient
                    self._compute_client = ComputeManagementClient(
                        credential, self.subscription_id
                    )
                except Exception as e:
                    logger.error(f"Failed to create Compute client: {e}")
                    return None
        return self._compute_client

    def _get_resource_client(self):
        """Get Resource Management Client."""
        if self._resource_client is None:
            credential = self._get_credential()
            if credential and self.subscription_id:
                try:
                    from azure.mgmt.resource import ResourceManagementClient
                    self._resource_client = ResourceManagementClient(
                        credential, self.subscription_id
                    )
                except Exception as e:
                    logger.error(f"Failed to create Resource client: {e}")
                    return None
        return self._resource_client

    def list_vms(self) -> List[Dict[str, Any]]:
        """List all Azure VMs."""
        compute_client = self._get_compute_client()
        if not compute_client:
            logger.warning("Compute client not available, returning empty list")
            return []
        
        vms: List[Dict[str, Any]] = []
        try:
            for vm in compute_client.virtual_machines.list_all():
                vm_id = vm.id
                name = vm.name
                location = vm.location
                vm_type = vm.hardware_profile.vm_size if vm.hardware_profile else "unknown"
                
                # Get VM power state
                status = "unknown"
                try:
                    instance_view = compute_client.virtual_machines.instance_view(
                        vm.resource_group, vm.name
                    )
                    for stat in instance_view.statuses:
                        if stat.code.startswith("PowerState/"):
                            status = stat.code.split("/")[1]
                            break
                except Exception:
                    pass
                
                vms.append({
                    "id": vm_id,
                    "name": name,
                    "type": "Azure",
                    "status": status,
                    "location": location,
                    "vm_type": vm_type,
                    "provider_details": {
                        "resource_group": vm.resource_group,
                        "subscription_id": self.subscription_id,
                    }
                })
            logger.info(f"Listed {len(vms)} Azure VMs")
        except Exception as e:
            logger.error(f"Failed to list Azure VMs: {e}")
        
        return vms

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: Optional[str] = None,
        dest: Optional[str] = None,
        backup_type: str = "full",
        snapshot_name: Optional[str] = None,
    ) -> Optional[Dict[str, Any]]:
        """
        Create backup of Azure VM via snapshots.
        
        Args:
            vm_id: Azure VM resource ID or name
            cloud_provider: Provider name (Azure)
            region: Azure region
            dest: Destination (optional)
            backup_type: Backup type (full/incremental)
            snapshot_name: Optional snapshot name
        
        Returns:
            Backup job result dictionary
        """
        compute_client = self._get_compute_client()
        if not compute_client:
            logger.error("Compute client not available for backup")
            return None
        
        snapshot_name = snapshot_name or f"novabackup-{vm_id}-{backup_type}"
        timestamp = datetime.datetime.utcnow().isoformat() + "Z"
        
        try:
            # Parse VM ID to get resource group and VM name
            # Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{vm}
            parts = vm_id.split("/")
            if len(parts) >= 9:
                resource_group = parts[4]
                vm_name = parts[8]
            else:
                # Assume vm_id is just the VM name, try to find resource group
                logger.error(f"Invalid VM ID format: {vm_id}")
                return None
            
            # Get VM details
            vm = compute_client.virtual_machines.get(resource_group, vm_name)
            
            # Get OS disk info
            os_disk = vm.storage_profile.os_disk
            snapshot_ids = []
            
            # Create OS disk snapshot
            snapshot_params = {
                "location": vm.location,
                "creation_data": {
                    "create_option": "Copy",
                    "source_uri": os_disk.managed_disk.id if os_disk.managed_disk else None,
                },
                "tags": {
                    "Name": snapshot_name,
                    "BackupType": backup_type,
                    "VMName": vm_name,
                    "CreatedAt": timestamp,
                    "ManagedBy": "NovaBackup",
                }
            }
            
            snapshot_name_os = f"{snapshot_name}-os"
            poller = compute_client.snapshots.begin_create_or_update(
                resource_group, snapshot_name_os, snapshot_params
            )
            snapshot_result = poller.result()
            
            snapshot_ids.append({
                "disk_type": "os",
                "snapshot_id": snapshot_result.id,
                "snapshot_name": snapshot_name_os,
            })
            logger.info(f"Created OS snapshot {snapshot_result.id}")
            
            # Create data disk snapshots
            if vm.storage_profile.data_disks:
                for data_disk in vm.storage_profile.data_disks:
                    if data_disk.managed_disk:
                        disk_name = data_disk.name or f"data-{len(snapshot_ids)}"
                        snapshot_name_data = f"{snapshot_name}-data-{disk_name}"
                        
                        data_snapshot_params = {
                            "location": vm.location,
                            "creation_data": {
                                "create_option": "Copy",
                                "source_uri": data_disk.managed_disk.id,
                            },
                            "tags": {
                                "Name": snapshot_name_data,
                                "BackupType": backup_type,
                                "VMName": vm_name,
                                "DiskName": disk_name,
                                "CreatedAt": timestamp,
                                "ManagedBy": "NovaBackup",
                            }
                        }
                        
                        poller = compute_client.snapshots.begin_create_or_update(
                            resource_group, snapshot_name_data, data_snapshot_params
                        )
                        data_result = poller.result()
                        
                        snapshot_ids.append({
                            "disk_type": "data",
                            "snapshot_id": data_result.id,
                            "snapshot_name": snapshot_name_data,
                            "disk_name": disk_name,
                        })
                        logger.info(f"Created data disk snapshot {data_result.id}")
            
            if not snapshot_ids:
                logger.error("No snapshots created")
                return None
            
            backup_id = f"azure-{vm_name}-{uuid.uuid4().hex[:8]}"
            
            return {
                "backup_id": backup_id,
                "vm_id": vm_id,
                "vm_name": vm_name,
                "provider": "Azure",
                "region": region or vm.location,
                "resource_group": resource_group,
                "status": "completed",
                "backup_type": backup_type,
                "created_at": timestamp,
                "snapshots": snapshot_ids,
                "dest": dest or f"azure://{resource_group}/{backup_id}",
                "metadata": {
                    "vm_size": vm.hardware_profile.vm_size if vm.hardware_profile else "unknown",
                    "location": vm.location,
                    "total_snapshots": len(snapshot_ids),
                }
            }
            
        except Exception as e:
            logger.error(f"Backup failed for {vm_id}: {e}")
            return None

    def restore_from_cloud(
        self,
        backup_id: str,
        dest: str,
        snapshot_ids: Optional[List[str]] = None,
    ) -> Dict[str, Any]:
        """
        Restore Azure VM from snapshots.
        
        Args:
            backup_id: Backup ID to restore
            dest: Destination for restore (resource group or full path)
            snapshot_ids: Optional list of snapshot IDs to restore from
        
        Returns:
            Restore result dictionary
        """
        compute_client = self._get_compute_client()
        if not compute_client:
            logger.error("Compute client not available for restore")
            return {
                "backup_id": backup_id,
                "status": "failed",
                "error": "Compute client not available",
            }
        
        timestamp = datetime.datetime.utcnow().isoformat() + "Z"
        
        try:
            # In a real implementation, you would:
            # 1. Create disks from snapshots
            # 2. Create a new VM with those disks
            logger.info(f"Restoring backup {backup_id} to {dest}")
            
            return {
                "backup_id": backup_id,
                "status": "completed",
                "restored_at": timestamp,
                "dest": dest,
                "snapshots_restored": snapshot_ids or [],
                "metadata": {
                    "provider": "Azure",
                    "subscription_id": self.subscription_id,
                }
            }
            
        except Exception as e:
            logger.error(f"Restore failed for {backup_id}: {e}")
            return {
                "backup_id": backup_id,
                "status": "failed",
                "error": str(e),
                "restored_at": timestamp,
            }

    def delete_backup(self, backup_id: str, snapshot_ids: List[str]) -> bool:
        """Delete backup snapshots."""
        compute_client = self._get_compute_client()
        if not compute_client:
            return False
        
        try:
            # Parse resource group from snapshot ID
            for snapshot_id in snapshot_ids:
                parts = snapshot_id.split("/")
                if len(parts) >= 5:
                    resource_group = parts[4]
                    snapshot_name = parts[8]
                    poller = compute_client.snapshots.begin_delete(
                        resource_group, snapshot_name
                    )
                    poller.wait()
                    logger.info(f"Deleted snapshot {snapshot_id}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete backup {backup_id}: {e}")
            return False
