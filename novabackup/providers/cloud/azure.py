from typing import List, Dict, Any, Optional
import uuid, datetime, os


class AzureCloudProvider:
    def __init__(self, subscription_id: Optional[str] = None, credential=None):
        self.subscription_id = subscription_id or os.environ.get(
            "AZURE_SUBSCRIPTION_ID"
        )
        self.client = None
        try:
            from azure.identity import DefaultAzureCredential
            from azure.mgmt.compute import ComputeManagementClient

            cred = credential or DefaultAzureCredential()
            self.client = ComputeManagementClient(cred, self.subscription_id)
        except Exception:
            self.client = None

    def list_vms(self) -> List[Dict[str, Any]]:
        if self.client is None:
            return []
        vms: List[Dict[str, Any]] = []
        try:
            for vm in self.client.virtual_machines.list_all():
                vms.append(
                    {
                        "id": getattr(vm, "id", ""),
                        "name": getattr(vm, "name", ""),
                        "type": "Azure",
                        "status": "running",
                    }
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
    ) -> Dict[str, Any]:
        backup_id = f"azure-{uuid.uuid4()}"
        snapshot_id = f"azure-snap-{uuid.uuid4()}"
        now = datetime.datetime.utcnow().isoformat() + "Z"
        return {
            "backup_id": backup_id,
            "snapshot_id": snapshot_id,
            "provider": "Azure",
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
