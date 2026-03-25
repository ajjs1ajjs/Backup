from typing import List, Dict, Any


class CloudProvider:
    """Abstract cloud provider interface for VM backups/restores."""

    def list_vms(self) -> List[Dict[str, Any]]:
        raise NotImplementedError

    def backup_to_cloud(
        self,
        vm_id: str,
        cloud_provider: str,
        region: str,
        dest: str,
        backup_type: str,
        snapshot_name: str | None = None,
    ) -> Dict[str, Any]:
        raise NotImplementedError

    def restore_from_cloud(self, backup_id: str, dest: str) -> Dict[str, Any]:
        raise NotImplementedError
