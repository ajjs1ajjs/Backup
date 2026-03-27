import logging
from typing import List, Optional, Any, Dict

logger = logging.getLogger(__name__)


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
            except Exception as e:
                logger.error(f"Error listing VMs from provider {p.__class__.__name__}: {e}")
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
        last_error: Optional[Exception] = None
        
        for p in self.providers:
            if provider_lower == p.__class__.__name__.lower():
                try:
                    if hasattr(p, "backup_to_cloud"):
                        return p.backup_to_cloud(
                            vm_id, provider, region, dest, backup_type, snapshot_name
                        )
                except Exception as e:
                    last_error = e
                    logger.error(
                        f"Backup to cloud failed for provider {p.__class__.__name__}: {e}",
                        exc_info=True
                    )
                    # Continue to next provider instead of returning None immediately
                    continue
        
        if last_error:
            logger.error(f"All cloud backup attempts failed. Last error: {last_error}")
        return None

    def restore_from_cloud(
        self, provider: str, backup_id: str, dest: str
    ) -> Optional[dict]:
        provider_lower = provider.lower() if provider else ""
        last_error: Optional[Exception] = None
        
        for p in self.providers:
            if provider_lower == p.__class__.__name__.lower():
                try:
                    if hasattr(p, "restore_from_cloud"):
                        return p.restore_from_cloud(backup_id, dest)
                except Exception as e:
                    last_error = e
                    logger.error(
                        f"Restore from cloud failed for provider {p.__class__.__name__}: {e}",
                        exc_info=True
                    )
                    continue
        
        if last_error:
            logger.error(f"All cloud restore attempts failed. Last error: {last_error}")
        return None
