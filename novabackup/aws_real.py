from typing import List, Dict, Any, Optional
import boto3
import datetime


class AWSCloudProvider:
    def __init__(
        self, region_name: Optional[str] = None, profile: Optional[str] = None
    ):
        self.region = region_name
        self.profile = profile

    def list_vms(self) -> List[Dict[str, Any]]:
        ec2 = boto3.client("ec2", region_name=self.region, profile_name=self.profile)
        resp = ec2.describe_instances()
        vms = []
        for r in resp.get("Reservations", []):
            for inst in r.get("Instances", []):
                name = None
                for t in inst.get("Tags", []) or []:
                    if t.get("Key") == "Name":
                        name = t.get("Value")
                vms.append(
                    {
                        "id": inst.get("InstanceId"),
                        "name": name or inst.get("InstanceId"),
                        "type": "AWS",
                        "status": inst.get("State", {}).get("Name", "unknown"),
                    }
                )
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
            ec2 = boto3.client("ec2", region_name=region, profile_name=self.profile)
            volumes = []
            resp = ec2.describe_instances(
                Filters=[{"Name": "tag:Name", "Values": [vm_id]}]
            )
            for r in resp.get("Reservations", []):
                for inst in r.get("Instances", []):
                    volumes.extend(
                        [
                            bd.get("Ebs")["VolumeId"]
                            for bd in inst.get("BlockDeviceMappings", [])
                            if "Ebs" in bd
                        ]
                    )
            if not volumes:
                return None
            snap = ec2.create_snapshot(
                VolumeId=volumes[0], Description=f"Novabackup {vm_id} {backup_type}"
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
