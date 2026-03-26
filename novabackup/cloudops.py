from typing import List, Optional, Any


class CloudOrchestrator:
    def __init__(self, providers: Optional[List[Any]] = None):
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
        # Find provider by class name (case-insensitive match)
        provider_lower = provider.lower() if provider else ""
        for p in self.providers:
            if provider_lower == p.__class__.__name__.lower():
                try:
                    if hasattr(p, "backup_to_cloud"):
                        return p.backup_to_cloud(
                            vm_id, provider, region, dest, backup_type, snapshot_name
                        )
                except Exception:
                    # If the provider fails, we return None for this provider.
                    return None
        return None

    def restore_from_cloud(
        self, provider: str, backup_id: str, dest: str
    ) -> Optional[dict]:
        provider_lower = provider.lower() if provider else ""
        for p in self.providers:
            if provider_lower == p.__class__.__name__.lower():
                try:
                    if hasattr(p, "restore_from_cloud"):
                        return p.restore_from_cloud(backup_id, dest)
                except Exception:
                    return None
        return None
