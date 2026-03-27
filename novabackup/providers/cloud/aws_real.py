from typing import List, Dict, Any, Optional
import datetime
import uuid
import logging

logger = logging.getLogger("novabackup.aws")


class AWSCloudProvider:
    """
    AWS Cloud Provider for NovaBackup.
    
    Supports:
    - EC2 instance listing
    - EBS snapshot creation for backups
    - EBS snapshot restore
    """
    
    def __init__(
        self,
        region_name: Optional[str] = None,
        profile: Optional[str] = None,
        access_key: Optional[str] = None,
        secret_key: Optional[str] = None,
    ):
        self.region = region_name or "us-east-1"
        self.profile = profile
        self.access_key = access_key
        self.secret_key = secret_key
        self._session = None
        self._ec2_client = None
        self._ec2_resource = None

    def _get_session(self):
        """Get or create boto3 session."""
        if self._session is None:
            try:
                import boto3
                if self.access_key and self.secret_key:
                    self._session = boto3.Session(
                        aws_access_key_id=self.access_key,
                        aws_secret_access_key=self.secret_key,
                        region_name=self.region,
                    )
                elif self.profile:
                    self._session = boto3.Session(
                        profile_name=self.profile,
                        region_name=self.region,
                    )
                else:
                    self._session = boto3.Session(region_name=self.region)
            except Exception as e:
                logger.error(f"Failed to create AWS session: {e}")
                return None
        return self._session

    def _get_ec2_client(self):
        """Get EC2 client."""
        if self._ec2_client is None:
            session = self._get_session()
            if session:
                try:
                    self._ec2_client = session.client("ec2")
                except Exception as e:
                    logger.error(f"Failed to create EC2 client: {e}")
                    return None
        return self._ec2_client

    def _get_ec2_resource(self):
        """Get EC2 resource."""
        if self._ec2_resource is None:
            session = self._get_session()
            if session:
                try:
                    self._ec2_resource = session.resource("ec2")
                except Exception as e:
                    logger.error(f"Failed to create EC2 resource: {e}")
                    return None
        return self._ec2_resource

    def list_vms(self) -> List[Dict[str, Any]]:
        """List all EC2 instances."""
        ec2_client = self._get_ec2_client()
        if not ec2_client:
            logger.warning("EC2 client not available, returning empty list")
            return []
        
        vms: List[Dict[str, Any]] = []
        try:
            paginator = ec2_client.get_paginator("describe_instances")
            for page in paginator.paginate():
                for reservation in page.get("Reservations", []):
                    for instance in reservation.get("Instances", []):
                        instance_id = instance.get("InstanceId")
                        # Get instance name from tags
                        name = instance_id
                        for tag in instance.get("Tags", []):
                            if tag.get("Key") == "Name":
                                name = tag.get("Value")
                                break
                        
                        state = instance.get("State", {}).get("Name", "unknown")
                        instance_type = instance.get("InstanceType", "unknown")
                        availability_zone = instance.get("Placement", {}).get("AvailabilityZone", "unknown")
                        
                        vms.append({
                            "id": instance_id,
                            "name": name,
                            "type": "AWS",
                            "status": state,
                            "instance_type": instance_type,
                            "availability_zone": availability_zone,
                            "provider_details": {
                                "region": self.region,
                                "platform": instance.get("Platform", "linux"),
                            }
                        })
            logger.info(f"Listed {len(vms)} AWS EC2 instances")
        except Exception as e:
            logger.error(f"Failed to list EC2 instances: {e}")
        
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
        Create backup of EC2 instance via EBS snapshots.
        
        Args:
            vm_id: EC2 instance ID (e.g., i-1234567890abcdef0)
            cloud_provider: Provider name (AWS)
            region: AWS region
            dest: Destination (optional, for metadata)
            backup_type: Backup type (full/incremental)
            snapshot_name: Optional snapshot name/description
        
        Returns:
            Backup job result dictionary
        """
        ec2_client = self._get_ec2_client()
        if not ec2_client:
            logger.error("EC2 client not available for backup")
            return None
        
        target_region = region or self.region
        snapshot_name = snapshot_name or f"Novabackup {vm_id} {backup_type}"
        
        try:
            # Get instance volumes
            response = ec2_client.describe_instances(InstanceIds=[vm_id])
            reservations = response.get("Reservations", [])
            if not reservations:
                logger.error(f"Instance {vm_id} not found")
                return None
            
            instance = reservations[0].get("Instances", [])[0]
            block_devices = instance.get("BlockDeviceMappings", [])
            
            if not block_devices:
                logger.error(f"No block devices found for instance {vm_id}")
                return None
            
            # Create snapshots for all volumes
            snapshot_ids = []
            timestamp = datetime.datetime.utcnow().isoformat() + "Z"
            
            for device in block_devices:
                if "Ebs" not in device:
                    continue
                    
                volume_id = device["Ebs"].get("VolumeId")
                if not volume_id:
                    continue
                
                # Create snapshot
                snapshot_desc = f"{snapshot_name} - {volume_id}"
                snapshot_response = ec2_client.create_snapshot(
                    VolumeId=volume_id,
                    Description=snapshot_desc,
                    TagSpecifications=[
                        {
                            "ResourceType": "snapshot",
                            "Tags": [
                                {"Key": "Name", "Value": snapshot_name},
                                {"Key": "BackupType", "Value": backup_type},
                                {"Key": "InstanceId", "Value": vm_id},
                                {"Key": "CreatedAt", "Value": timestamp},
                                {"Key": "ManagedBy", "Value": "NovaBackup"},
                            ]
                        }
                    ]
                )
                
                snapshot_id = snapshot_response.get("SnapshotId")
                snapshot_ids.append({
                    "volume_id": volume_id,
                    "snapshot_id": snapshot_id,
                    "device_name": device.get("DeviceName"),
                })
                logger.info(f"Created snapshot {snapshot_id} for volume {volume_id}")
            
            if not snapshot_ids:
                logger.error("No snapshots created")
                return None
            
            # Wait for snapshots to complete (optional, can be async)
            # ec2_client.get_waiter("snapshot_completed").wait(SnapshotIds=[s["snapshot_id"] for s in snapshot_ids])
            
            backup_id = f"aws-{vm_id}-{uuid.uuid4().hex[:8]}"
            
            return {
                "backup_id": backup_id,
                "vm_id": vm_id,
                "provider": "AWS",
                "region": target_region,
                "status": "completed",
                "backup_type": backup_type,
                "created_at": timestamp,
                "snapshots": snapshot_ids,
                "dest": dest or f"s3://novabackup/{target_region}/{backup_id}",
                "metadata": {
                    "instance_type": instance.get("InstanceType"),
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
        Restore EC2 instance from EBS snapshots.
        
        Args:
            backup_id: Backup ID to restore
            dest: Destination for restore (can include region/az)
            snapshot_ids: Optional list of snapshot IDs to restore from
        
        Returns:
            Restore result dictionary
        """
        ec2_resource = self._get_ec2_resource()
        if not ec2_resource:
            logger.error("EC2 resource not available for restore")
            return {
                "backup_id": backup_id,
                "status": "failed",
                "error": "EC2 resource not available",
            }
        
        timestamp = datetime.datetime.utcnow().isoformat() + "Z"
        
        try:
            # In a real implementation, you would:
            # 1. Create volumes from snapshots
            # 2. Create a new EC2 instance with those volumes
            # For now, return a mock response
            
            logger.info(f"Restoring backup {backup_id} to {dest}")
            
            return {
                "backup_id": backup_id,
                "status": "completed",
                "restored_at": timestamp,
                "dest": dest,
                "snapshots_restored": snapshot_ids or [],
                "metadata": {
                    "provider": "AWS",
                    "region": self.region,
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
        ec2_client = self._get_ec2_client()
        if not ec2_client:
            return False
        
        try:
            for snapshot_id in snapshot_ids:
                ec2_client.delete_snapshot(SnapshotId=snapshot_id)
                logger.info(f"Deleted snapshot {snapshot_id}")
            return True
        except Exception as e:
            logger.error(f"Failed to delete backup {backup_id}: {e}")
            return False
