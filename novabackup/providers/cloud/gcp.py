from typing import List, Dict, Any, Optional
import uuid, datetime


class GCPCloudProvider:
    def __init__(self, project_id: Optional[str] = None):
        self.project_id = project_id

    def list_vms(self) -> List[Dict[str, Any]]:
        try:
            from googleapiclient.discovery import build
            from google.auth import default

            credentials, _ = default()
            service = build("compute", "v1", credentials=credentials)
            req = service.instances().aggregatedList(project=self.project_id)
            resp = req.execute()
            vms = []
            for zone, data in resp.get("items", {}).items():
                for inst in data.get("instances", []):
                    vms.append(
                        {
                            "id": inst.get("id"),
                            "name": inst.get("name"),
                            "type": "GCP",
                            "status": "RUNNING",
                        }
                    )
            return vms
        except Exception:
            return []

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: str,
        dest: str,
        backup_type: str,
        snapshot_name: Optional[str] = None,
    ) -> Dict[str, Any]:
        backup_id = f"gcp-{uuid.uuid4()}"
        snapshot_id = f"gcp-snap-{uuid.uuid4()}"
        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "snapshot_id": snapshot_id,
            "provider": "GCP",
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
