import json
import os
import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional


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
        if not self.use_db:
            # Storage is a simple JSON file to persist backups across CLI invocations
            self.storage_path = storage_path or self._default_storage_path()
            self.jobs: Dict[str, Dict[str, Any]] = self._load_backups()
        # Try to instantiate a provider snapshotter lazily for non-DB path
        self._snapshotter = self._select_snapshotter()

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
    ) -> Dict[str, Any]:
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
