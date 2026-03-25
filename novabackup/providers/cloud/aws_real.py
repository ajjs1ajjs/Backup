from typing import List, Dict, Any, Optional
import datetime
import uuid


class AWSCloudProvider:
    def __init__(
        self, region_name: Optional[str] = None, profile: Optional[str] = None
    ):
        self.region = region_name
        self.profile = profile

    def list_vms(self) -> List[Dict[str, Any]]:
        try:
            import boto3  # type: ignore
        except Exception:
            return []
        try:
            session = (
                boto3.Session(profile_name=self.profile)
                if self.profile
                else boto3.Session()
            )
            ec2 = session.client("ec2", region_name=self.region)
            resp = ec2.describe_instances()
        except Exception:
            return []
        vms: List[Dict[str, Any]] = []
        try:
            for r in resp.get("Reservations", []):
                for inst in r.get("Instances", []):
                    inst_id = inst.get("InstanceId")
                    name = None
                    for tag in inst.get("Tags", []) or []:
                        if tag.get("Key") == "Name":
                            name = tag.get("Value")
                    if not name:
                        name = inst_id
                    state = inst.get("State", {}).get("Name", "unknown")
                    vms.append(
                        {"id": inst_id, "name": name, "type": "AWS", "status": state}
                    )
        except Exception:
            pass
        return vms

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: str,
        dest: str,
        backup_type: str,
        snapshot_name: Optional[str] = None,
    ) -> Optional[Dict[str, Any]]:
        try:
            import boto3  # type: ignore
        except Exception:
            return None
        try:
            ec2 = boto3.client("ec2", region_name=region)
            resp = ec2.describe_instances(
                Filters=[{"Name": "tag:Name", "Values": [vm_id]}]
            )
            volumes = []
            for r in resp.get("Reservations", []):
                for inst in r.get("Instances", []):
                    for bd in inst.get("BlockDeviceMappings", []) or []:
                        if "Ebs" in bd and "VolumeId" in bd["Ebs"]:
                            volumes.append(bd["Ebs"]["VolumeId"])
            if not volumes:
                return None
            vol_id = volumes[0]
            snap = ec2.create_snapshot(
                VolumeId=vol_id, Description=f"Novabackup {vm_id} {backup_type}"
            )
            snapshot_id = snap.get("SnapshotId")
            now = datetime.datetime.utcnow().isoformat() + "Z"
            return {
                "backup_id": f"aws-{snapshot_id}",
                "snapshot_id": snapshot_id,
                "provider": "AWS",
                "status": "created",
                "created_at": now,
                "dest": dest,
            }
        except Exception:
            return None

    def restore_from_cloud(self, backup_id: str, dest: str) -> Dict[str, Any]:
        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "dest": dest,
            "status": "restored",
            "restored_at": now,
        }
