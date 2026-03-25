import datetime
import os
import sqlite3
import uuid
import platform
from typing import Any, Dict, List


class DBManager:
    """Lightweight SQLite-based backend for backups and restores using sqlite3.

    Supports URL patterns like sqlite:///path/to/db.sqlite
    If SQLAlchemy is not installed, this module provides a self-contained, minimal DB.
    """

    def __init__(self, database_url: str = None):
        # Resolve database path
        url = database_url or os.environ.get(
            "NOVABACKUP_DATABASE_URL", "sqlite:///novabackup.db"
        )
        if url.startswith("sqlite:///"):
            path = url[len("sqlite:///") :]
        else:
            # Fallback: treat as a path
            path = url
        self._path = path or "./novabackup.db"
        self._conn = sqlite3.connect(self._path, check_same_thread=False)
        self._conn.execute("PRAGMA foreign_keys = ON;")
        self._ensure_schema()

    def _ensure_schema(self) -> None:
        cur = self._conn.cursor()
        cur.execute(
            """
            CREATE TABLE IF NOT EXISTS backups (
              id INTEGER PRIMARY KEY AUTOINCREMENT,
              backup_id TEXT UNIQUE,
              vm_id TEXT,
              dest TEXT,
              type TEXT,
              provider TEXT,
              status TEXT,
              created_at TEXT,
              snapshot_id TEXT,
              snapshot_provider TEXT,
              snapshot_type TEXT,
              snapshot_name TEXT
            )
            """
        )
        cur.execute(
            """
            CREATE TABLE IF NOT EXISTS restores (
              id INTEGER PRIMARY KEY AUTOINCREMENT,
              restore_id TEXT UNIQUE,
              backup_id TEXT,
              vm_id TEXT,
              dest TEXT,
              status TEXT,
              started_at TEXT,
              finished_at TEXT
            )
            """
        )
        self._conn.commit()

    def create_backup(
        self,
        vm_id: str,
        dest: str,
        backup_type: str = "full",
        snapshot_name: str = None,
    ) -> Dict[str, Any]:
        backup_id = str(uuid.uuid4())
        snapshot_id = str(uuid.uuid4())
        now = datetime.datetime.utcnow().isoformat() + "Z"
        provider = "HyperV" if platform.system().lower() == "windows" else "Libvirt"
        cur = self._conn.cursor()
        cur.execute(
            """
            INSERT INTO backups(backup_id, vm_id, dest, type, provider, status, created_at,
                                snapshot_id, snapshot_provider, snapshot_type, snapshot_name)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                backup_id,
                vm_id,
                dest,
                backup_type,
                provider,
                "created",
                now,
                snapshot_id,
                provider,
                backup_type,
                snapshot_name,
            ),
        )
        self._conn.commit()
        snapshot = {
            "snapshot_id": snapshot_id,
            "provider": provider,
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "name": snapshot_name,
            "status": "created",
        }
        return {
            "backup_id": backup_id,
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "snapshot": snapshot,
            "provider": provider,
            "status": "created",
            "created_at": now,
        }

    def list_backups(self) -> List[Dict[str, Any]]:
        cur = self._conn.cursor()
        cur.execute(
            """SELECT backup_id, vm_id, dest, type, provider, status, created_at, snapshot_id, snapshot_provider, snapshot_type, snapshot_name FROM backups"""
        )
        rows = cur.fetchall()
        backups = []
        for r in rows:
            (
                backup_id,
                vm_id,
                dest,
                btype,
                provider,
                status,
                created_at,
                snapshot_id,
                sp,
                st,
                sn,
            ) = r
            snapshot = {
                "snapshot_id": snapshot_id,
                "provider": sp,
                "vm_id": vm_id,
                "dest": dest,
                "type": st,
                "name": sn,
                "status": status,
            }
            backups.append(
                {
                    "backup_id": backup_id,
                    "vm_id": vm_id,
                    "dest": dest,
                    "type": btype,
                    "snapshot": snapshot,
                    "provider": provider,
                    "status": status,
                    "created_at": created_at,
                }
            )
        return backups

    def restore_backup(self, backup_id: str, dest: str) -> Dict[str, Any]:
        cur = self._conn.cursor()
        cur.execute(
            "SELECT backup_id, vm_id, dest, type, provider, status, created_at FROM backups WHERE backup_id = ?",
            (backup_id,),
        )
        row = cur.fetchone()
        if not row:
            raise KeyError(f"Backup {backup_id} not found")
        (b_id, vm_id, old_dest, btype, provider, status, created_at) = row
        now = datetime.datetime.utcnow().isoformat() + "Z"
        restore_id = str(uuid.uuid4())
        cur.execute(
            """
            INSERT INTO restores(restore_id, backup_id, vm_id, dest, status, started_at, finished_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            (restore_id, backup_id, vm_id, dest, "restored", now, now),
        )
        cur.execute(
            "UPDATE backups SET status = ? WHERE backup_id = ?", ("restored", backup_id)
        )
        self._conn.commit()
        return {
            "backup_id": backup_id,
            "vm_id": vm_id,
            "dest": dest,
            "status": "restored",
            "restored_at": now,
        }
