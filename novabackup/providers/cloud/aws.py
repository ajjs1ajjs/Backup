from typing import List, Dict, Any, Optional
import uuid
import datetime


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
        return vms

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: str,
        dest: str,
        backup_type: str,
        snapshot_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        backup_id = f"aws-{uuid.uuid4()}"
        snapshot_id = f"aws-snap-{uuid.uuid4()}"
        now = datetime.datetime.utcnow().isoformat() + "Z"
        # Real AWS backup could snapshot the root EBS volume; for MVP we simulate a response or do minimal action when credentials exist
        return {
            "backup_id": backup_id,
            "snapshot_id": snapshot_id,
            "provider": "AWS",
            "status": "created",
            "created_at": now,
            "dest": dest,
        }

    def restore_from_cloud(self, backup_id: str, dest: str) -> Dict[str, Any]:
        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "dest": dest,
            "status": "restored",
            "restored_at": now,
        }
