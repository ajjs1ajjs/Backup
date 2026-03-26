import json
import os
import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional
from novabackup.cloudops import CloudOrchestrator


# Snapshotter interfaces (minimal MVP)
class AbstractSnapshotter:
    def create_snapshot(
        self,
        vm_id: str,
        dest: str,
        snapshot_name: Optional[str] = None,
        backup_type: str = "full",
    ) -> Dict[str, str]:
        snapshot_id = str(uuid.uuid4())
        return {
            "snapshot_id": snapshot_id,
            "provider": self.__class__.__name__.replace("Snapshotter", ""),
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "name": snapshot_name,
            "status": "created",
        }


class HyperVSnapshotter(AbstractSnapshotter):
    pass


class LibvirtSnapshotter(AbstractSnapshotter):
    pass


class BackupManager:
    def __init__(
        self, storage_path: Optional[str] = None, database_url: Optional[str] = None
    ):
        # Attempt to use a DB backend if available; fallback to JSON store otherwise
        self.use_db = False
        self._db = None
        try:
            from .db import DBManager  # type: ignore

            _db_available = True
        except Exception:
            _db_available = False
            DBManager = None  # type: ignore
        if _db_available:
            url = database_url or os.environ.get("NOVABACKUP_DATABASE_URL")
            if url:
                try:
                    self._db = DBManager(database_url=url)
                    self.use_db = True
                except Exception:
                    self.use_db = False
        # Also try SQLAlchemy-based DB backend when NOVABACKUP_DATABASE_URL is provided
        if not self.use_db:
            try:
                from .db_sa import SA_DBManager  # type: ignore

                url = database_url or os.environ.get("NOVABACKUP_DATABASE_URL")
                if url:
                    self._db = SA_DBManager(url)
                    self.use_db = True
            except Exception:
                pass
        if not self.use_db:
            # Storage is a simple JSON file to persist backups across CLI invocations
            self.storage_path = storage_path or self._default_storage_path()
            self.jobs: Dict[str, Dict[str, Any]] = self._load_backups()
        # Try to instantiate a provider snapshotter lazily for non-DB path
        self._snapshotter = self._select_snapshotter()
        self._cloud_orchestrator = None

    def _get_cloud_orchestrator(self) -> Optional[CloudOrchestrator]:
        # Lazy construction of cloud orchestrator with available providers
        if self._cloud_orchestrator is not None:
            return self._cloud_orchestrator
        providers = []
        try:
            from novabackup.providers.cloud.aws import AWSCloudProvider  # type: ignore

            providers.append(AWSCloudProvider())  # type: ignore
        except Exception:
            pass
        try:
            from novabackup.azure_real import AzureCloudProvider  # type: ignore

            providers.append(AzureCloudProvider())  # type: ignore
        except Exception:
            pass
        try:
            from novabackup.gcp_real import GCPCloudProvider  # type: ignore

            providers.append(GCPCloudProvider())  # type: ignore
        except Exception:
            pass
        try:
            from novabackup.providers.cloud.mock import MockCloudProvider  # type: ignore

            providers.append(MockCloudProvider())  # type: ignore
        except Exception:
            pass
        if not providers:
            return None
        self._cloud_orchestrator = CloudOrchestrator(providers=providers)  # type: ignore
        return self._cloud_orchestrator

    def _default_storage_path(self) -> str:
        project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        return os.path.join(project_root, "backups.json")

    def _load_backups(self) -> Dict[str, Dict[str, Any]]:
        if os.path.exists(self.storage_path):
            try:
                with open(self.storage_path, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    if isinstance(data, dict):
                        return data
            except Exception:
                pass
        return {}

    def _save_backups(self) -> None:
        try:
            with open(self.storage_path, "w", encoding="utf-8") as f:
                json.dump(self.jobs, f, indent=2)
        except Exception:
            pass

    def _select_snapshotter(self) -> Optional[AbstractSnapshotter]:
        # Import and instantiate available providers, prefer Hyper-V, then Libvirt
        hyperv = None
        libvirt = None
        try:
            from novabackup.core import HyperVProvider  # type: ignore
        except Exception:
            HyperVProvider = None  # type: ignore
        try:
            from novabackup.core import LibvirtProvider  # type: ignore
        except Exception:
            LibvirtProvider = None  # type: ignore
        if "HyperVProvider" in locals() and HyperVProvider is not None:
            hyperv = True
        if "LibvirtProvider" in locals() and LibvirtProvider is not None:
            libvirt = True
        if hyperv:
            return HyperVSnapshotter()
        if libvirt:
            return LibvirtSnapshotter()
        return None

    def create_backup(
        self,
        vm_id: str,
        dest: str,
        backup_type: str = "full",
        snapshot_name: Optional[str] = None,
        destination_type: str = "local",
        cloud_provider: Optional[str] = None,
        cloud_region: Optional[str] = None,
        cloud_dest: Optional[str] = None,
    ) -> Dict[str, Any]:
        # Cloud backup path
        if destination_type and destination_type.lower() == "cloud":
            orchestrator = self._get_cloud_orchestrator()
            if orchestrator is not None:
                # Determine which provider to use
                provider_param = cloud_provider or "MOCK"
                res = orchestrator.backup_to_cloud(
                    vm_id=vm_id,
                    provider=provider_param,
                    region=cloud_region,
                    dest=cloud_dest or dest,
                    backup_type=backup_type,
                    snapshot_name=snapshot_name,
                )
                if res is not None:
                    backup_id = res.get("backup_id") or f"cloud-{uuid.uuid4()}"
                    snapshot_entry = {
                        "snapshot_id": res.get("snapshot_id") or "",
                        "provider": res.get("provider") or (cloud_provider or ""),
                        "vm_id": vm_id,
                        "dest": cloud_dest or dest,
                        "type": backup_type,
                        "name": snapshot_name,
                        "status": res.get("status") or "created",
                    }
                    job = {
                        "backup_id": backup_id,
                        "vm_id": vm_id,
                        "dest": cloud_dest or dest,
                        "type": backup_type,
                        "snapshot": snapshot_entry,
                        "provider": res.get("provider") or (cloud_provider or ""),
                        "status": res.get("status") or "created",
                        "created_at": res.get("created_at")
                        or datetime.utcnow().isoformat() + "Z",
                    }
                    self.jobs[backup_id] = job
                    self._save_backups()
                    return job
        if self.use_db and self._db is not None:
            return self._db.create_backup(vm_id, dest, backup_type, snapshot_name)
        if self._snapshotter:
            snapshot = self._snapshotter.create_snapshot(
                vm_id, dest, snapshot_name, backup_type
            )
            provider = snapshot.get("provider", "Unknown")
        else:
            # Fallback mock snapshot
            snapshot = {
                "snapshot_id": str(uuid.uuid4()),
                "provider": "Mock",
                "vm_id": vm_id,
                "dest": dest,
                "type": backup_type,
                "name": snapshot_name,
                "status": "created",
            }
            provider = "Mock"

        backup_id = str(uuid.uuid4())
        job = {
            "backup_id": backup_id,
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "snapshot": snapshot,
            "provider": provider,
            "status": "created",
            "created_at": datetime.utcnow().isoformat() + "Z",
        }
        self.jobs[backup_id] = job
        self._save_backups()
        return job

    def list_backups(self) -> List[Dict[str, Any]]:
        if self.use_db and self._db is not None:
            return self._db.list_backups()
        return list(self.jobs.values())

    def restore_backup(self, backup_id: str, dest: str) -> Dict[str, Any]:
        if self.use_db and self._db is not None:
            return self._db.restore_backup(backup_id, dest)
        if backup_id not in self.jobs:
            raise KeyError(f"Backup {backup_id} not found")
        job = self.jobs[backup_id]
        # Simulate restore operation
        result = {
            "backup_id": backup_id,
            "vm_id": job["vm_id"],
            "dest": dest,
            "status": "restored",
            "restored_at": datetime.utcnow().isoformat() + "Z",
        }
        job["status"] = "restored"
        self._save_backups()
        return result
