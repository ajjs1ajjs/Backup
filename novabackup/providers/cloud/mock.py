from typing import List, Dict, Any

from .base import CloudProvider


class MockCloudProvider(CloudProvider):
    def list_vms(self) -> List[Dict[str, Any]]:
        return [
            {
                "id": "cloud_vm1",
                "name": "cloud-dev-1",
                "type": "Cloud",
                "status": "running",
            },
            {
                "id": "cloud_vm2",
                "name": "cloud-dev-2",
                "type": "Cloud",
                "status": "stopped",
            },
        ]

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: str,
        dest: str,
        backup_type: str,
        snapshot_name: str | None = None,
    ) -> Dict[str, Any]:
        # Mock response for cloud backup
        import uuid, datetime

        backup_id = f"cloud-{uuid.uuid4()}"
        snapshot_id = f"cloud-snap-{uuid.uuid4()}"
        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "snapshot_id": snapshot_id,
            "provider": cloud_provider,
            "status": "created",
            "created_at": now,
            "dest": dest,
        }

    def restore_from_cloud(self, backup_id: str, dest: str) -> Dict[str, Any]:
        import datetime

        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "dest": dest,
            "status": "restored",
            "restored_at": now,
        }
