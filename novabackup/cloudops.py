from typing import List, Optional, Any

"""Cloud orchestration layer for backups/restores across AWS, Azure, and GCP.
This is a lightweight orchestrator that coordinates among cloud provider backends.
Currently, the cloud providers are implemented in novabackup/providers/cloud/*.py.
"""


class CloudOrchestrator:
    def __init__(self, providers: Optional[List[Any]] = None):
        # Accept a list of provider instances; if None, rely on environment to instantiate
        self.providers = providers or []

    def list_vms(self) -> List[dict]:
        vms: List[dict] = []
        for p in self.providers:
            try:
                items = p.list_vms()
                if isinstance(items, list):
                    vms.extend(items)
            except Exception:
                continue
        return vms

    def backup_to_cloud(
        self,
        vm_id: str,
        provider: str,
        region: Optional[str],
        dest: str,
        backup_type: str,
        snapshot_name: Optional[str] = None,
    ) -> Optional[dict]:
        for p in self.providers:
            try:
                name = getattr(p, "__class__", type(p)).__name__.lower()
                if (provider or name).lower() == "aws" and hasattr(
                    p, "backup_to_cloud"
                ):
                    return p.backup_to_cloud(
                        vm_id, provider, region, dest, backup_type, snapshot_name
                    )
                if (provider or name).lower() == "azure" and hasattr(
                    p, "backup_to_cloud"
                ):
                    return p.backup_to_cloud(
                        vm_id, provider, region, dest, backup_type, snapshot_name
                    )
                if (provider or name).lower() == "gcp" and hasattr(
                    p, "backup_to_cloud"
                ):
                    return p.backup_to_cloud(
                        vm_id, provider, region, dest, backup_type, snapshot_name
                    )
            except Exception:
                continue
        return None

    def restore_from_cloud(
        self, provider: str, backup_id: str, dest: str
    ) -> Optional[dict]:
        for p in self.providers:
            try:
                name = getattr(p, "__class__", type(p)).__name__.lower()
                if (provider or name).lower() in ["aws", "azure", "gcp"] and hasattr(
                    p, "restore_from_cloud"
                ):
                    return p.restore_from_cloud(backup_id, dest)
            except Exception:
                continue
        return None
