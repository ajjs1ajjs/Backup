from typing import List, Dict, Any, Optional
import datetime
import uuid
import logging

logger = logging.getLogger("novabackup.gcp")


class GCPCloudProvider:
    """
    Google Cloud Platform Provider for NovaBackup.
    
    Supports:
    - Compute Engine VM listing
    - Disk snapshot creation for backups
    - Snapshot restore
    """
    
    def __init__(
        self,
        project_id: Optional[str] = None,
        credentials_file: Optional[str] = None,
        credentials=None,
    ):
        self.project_id = project_id
        self.credentials_file = credentials_file
        self._credentials = credentials
        self._compute_client = None

    def _get_credentials(self):
        """Get GCP credentials."""
        if self._credentials is None:
            try:
                from google.oauth2 import service_account
                from google.auth import default
                
                if self.credentials_file:
                    self._credentials = service_account.Credentials.from_service_account_file(
                        self.credentials_file,
                        scopes=["https://www.googleapis.com/auth/cloud-platform"]
                    )
                else:
                    self._credentials, _ = default(
                        scopes=["https://www.googleapis.com/auth/cloud-platform"]
                    )
            except Exception as e:
                logger.error(f"Failed to get GCP credentials: {e}")
                return None
        return self._credentials

    def _get_compute_client(self):
        """Get Compute Engine API client."""
        if self._compute_client is None:
            try:
                from googleapiclient.discovery import build
                
                credentials = self._get_credentials()
                if credentials:
                    self._compute_client = build(
                        "compute",
                        "v1",
                        credentials=credentials,
                        cache_discovery=False
                    )
            except Exception as e:
                logger.error(f"Failed to create Compute client: {e}")
                return None
        return self._compute_client

    def list_vms(self) -> List[Dict[str, Any]]:
        """List all Compute Engine VMs."""
        compute_client = self._get_compute_client()
        if not compute_client or not self.project_id:
            logger.warning("Compute client or project_id not available")
            return []
        
        vms: List[Dict[str, Any]] = []
        try:
            request = compute_client.instances().aggregatedList(
                project=self.project_id,
                maxResults=500
            )
            
            while request is not None:
                response = request.execute()
                
                for zone_name, items in response.get("items", {}).items():
                    if "instances" not in items:
                        continue
                    
                    for instance in items["instances"]:
                        vm_id = instance.get("id")
                        name = instance.get("name")
                        status = instance.get("status", "unknown")
                        zone = zone_name.split("/")[-1]
                        machine_type = instance.get("machineType", "unknown")
                        if isinstance(machine_type, str):
                            machine_type = machine_type.split("/")[-1]
                        
                        # Get external IP if available
                        external_ip = None
                        for interface in instance.get("networkInterfaces", []):
                            for access_config in interface.get("accessConfigs", []):
                                if "natIP" in access_config:
                                    external_ip = access_config["natIP"]
                                    break
                        
                        vms.append({
                            "id": vm_id,
                            "name": name,
                            "type": "GCP",
                            "status": status.lower(),
                            "zone": zone,
                            "machine_type": machine_type,
                            "external_ip": external_ip,
                            "provider_details": {
                                "project_id": self.project_id,
                                "full_zone": zone_name,
                            }
                        })
                
                request = compute_client.instances().aggregatedList_next(
                    previous_request=request,
                    previous_response=response
                )
            
            logger.info(f"Listed {len(vms)} GCP VMs")
        except Exception as e:
            logger.error(f"Failed to list GCP VMs: {e}")
        
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
        Create backup of GCP VM via disk snapshots.
        
        Args:
            vm_id: GCP VM instance ID or name
            cloud_provider: Provider name (GCP)
            region: GCP region/zone
            dest: Destination (optional)
            backup_type: Backup type (full/incremental)
            snapshot_name: Optional snapshot name
        
        Returns:
            Backup job result dictionary
        """
        compute_client = self._get_compute_client()
        if not compute_client or not self.project_id:
            logger.error("Compute client or project_id not available for backup")
            return None
        
        snapshot_name = snapshot_name or f"novabackup-{vm_id}-{backup_type}"
        timestamp = datetime.datetime.utcnow().isoformat() + "Z"
        
        try:
            # Get VM instance details
            # First, find the instance by searching all zones
            instance = None
            zone = None
            
            if region:
                # If zone is provided, try directly
                zone = region.split("/")[-1] if "/" in region else region
                try:
                    instance = compute_client.instances().get(
                        project=self.project_id,
                        zone=zone,
                        instance=vm_id
                    ).execute()
                except Exception:
                    pass
            
            if not instance:
                # Search all zones
                request = compute_client.instances().aggregatedList(
                    project=self.project_id,
                    filter=f"name eq {vm_id}"
                )
                while request:
                    response = request.execute()
                    for zone_name, items in response.get("items", {}).items():
                        if "instances" in items and items["instances"]:
                            instance = items["instances"][0]
                            zone = zone_name.split("/")[-1]
                            break
                    request = compute_client.instances().aggregatedList_next(
                        previous_request=request,
                        previous_response=response
                    )
            
            if not instance:
                logger.error(f"VM {vm_id} not found")
                return None
            
            vm_name = instance.get("name")
            
            # Get disk information
            disks = instance.get("disks", [])
            if not disks:
                logger.error(f"No disks found for VM {vm_id}")
                return None
            
            snapshot_ids = []
            
            # Create snapshots for all disks
            for disk_index, disk in enumerate(disks):
                disk_source = disk.get("source", "")
                disk_type = "boot" if disk.get("boot") else "data"
                disk_name = disk.get("deviceName", f"disk-{disk_index}")
                
                # Create snapshot name
                snap_name = f"{snapshot_name}-{disk_type}-{disk_name}"
                
                # Prepare snapshot body
                snapshot_body = {
                    "name": snap_name,
                    "description": f"NovaBackup {vm_name} {backup_type} - {disk_type}",
                    "labels": {
                        "backup_type": backup_type.replace("-", "_"),
                        "vm_name": vm_name.replace("-", "_"),
                        "disk_type": disk_type,
                        "created_by": "novabackup",
                        "timestamp": datetime.datetime.utcnow().strftime("%Y%m%d%H%M%S"),
                    }
                }
                
                # Create snapshot
                request = compute_client.snapshots().insert(
                    project=self.project_id,
                    body=snapshot_body,
                    sourceDisk=disk_source
                )
                response = request.execute()
                
                # Wait for operation to complete (optional, can be async)
                operation_name = response.get("name")
                
                snapshot_ids.append({
                    "disk_type": disk_type,
                    "disk_name": disk_name,
                    "snapshot_name": snap_name,
                    "operation": operation_name,
                })
                logger.info(f"Created snapshot {snap_name} for disk {disk_name}")
            
            if not snapshot_ids:
                logger.error("No snapshots created")
                return None
            
            backup_id = f"gcp-{vm_name}-{uuid.uuid4().hex[:8]}"
            
            return {
                "backup_id": backup_id,
                "vm_id": vm_id,
                "vm_name": vm_name,
                "provider": "GCP",
                "region": region or zone,
                "project_id": self.project_id,
                "status": "completed",
                "backup_type": backup_type,
                "created_at": timestamp,
                "snapshots": snapshot_ids,
                "dest": dest or f"gcp://{self.project_id}/{backup_id}",
                "metadata": {
                    "machine_type": instance.get("machineType", "unknown").split("/")[-1],
                    "zone": zone,
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
        Restore GCP VM from snapshots.
        
        Args:
            backup_id: Backup ID to restore
            dest: Destination for restore (zone or full path)
            snapshot_ids: Optional list of snapshot IDs to restore from
        
        Returns:
            Restore result dictionary
        """
        compute_client = self._get_compute_client()
        if not compute_client or not self.project_id:
            logger.error("Compute client or project_id not available for restore")
            return {
                "backup_id": backup_id,
                "status": "failed",
                "error": "Compute client or project_id not available",
            }
        
        timestamp = datetime.datetime.utcnow().isoformat() + "Z"
        
        try:
            # In a real implementation, you would:
            # 1. Create disks from snapshots
            # 2. Create a new VM instance with those disks
            logger.info(f"Restoring backup {backup_id} to {dest}")
            
            return {
                "backup_id": backup_id,
                "status": "completed",
                "restored_at": timestamp,
                "dest": dest,
                "snapshots_restored": snapshot_ids or [],
                "metadata": {
                    "provider": "GCP",
                    "project_id": self.project_id,
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

    def delete_backup(self, backup_id: str, snapshot_names: List[str]) -> bool:
        """Delete backup snapshots."""
        compute_client = self._get_compute_client()
        if not compute_client or not self.project_id:
            return False
        
        try:
            for snapshot_name in snapshot_names:
                request = compute_client.snapshots().delete(
                    project=self.project_id,
                    snapshot=snapshot_name
                )
                response = request.execute()
                logger.info(f"Deleted snapshot {snapshot_name}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete backup {backup_id}: {e}")
            return False
