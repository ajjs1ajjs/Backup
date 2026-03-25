import json
import os
import sqlite3
from typing import Any, Dict, Optional


def _parse_db_path(db_url: Optional[str]) -> str:
    if not db_url:
        return os.path.join(os.getcwd(), "novabackup.db")
    u = db_url
    if u.startswith("sqlite:///"):
        return u[len("sqlite:///") :]
    if u.startswith("sqlite:///"):
        return u[len("sqlite:///") :]
    if u.startswith("sqlite://"):
        return u[len("sqlite://") :]
    return u


def _ensure_schema(conn: sqlite3.Connection) -> None:
    cur = conn.cursor()
    cur.execute(
        """
        CREATE TABLE IF NOT EXISTS backups (
            backup_id TEXT PRIMARY KEY,
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
            restore_id TEXT PRIMARY KEY,
            backup_id TEXT,
            vm_id TEXT,
            dest TEXT,
            status TEXT,
            started_at TEXT,
            finished_at TEXT
        )
        """
    )
    conn.commit()


def migrate_json_to_db(
    json_path: Optional[str] = None, database_url: Optional[str] = None
) -> Dict[str, int]:
    """Migrate backups.json data into a SQLite DB.

    The function expects backups.json to contain a dict mapping backup_id -> backup_job,
    where each job includes 'vm_id', 'dest', 'type', 'provider', 'status', 'created_at',
    and an optional 'snapshot' dict with snapshotId and related fields.
    Returns a summary dict with migrated counts.
    """
    json_path = json_path or os.path.join(os.getcwd(), "backups.json")
    if not os.path.exists(json_path):
        return {"migrated_backups": 0, "migrated_restores": 0}

    db_path = _parse_db_path(database_url)
    need_db = not db_path.endswith(".json")
    if need_db:
        conn = sqlite3.connect(db_path)
        _ensure_schema(conn)
    else:
        return {"migrated_backups": 0, "migrated_restores": 0}

    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    migrated_backups = 0
    cur = conn.cursor()
    if isinstance(data, dict):
        for backup_id, job in data.items():
            if not isinstance(job, dict):
                continue
            vm_id = job.get("vm_id") or ""
            dest = job.get("dest") or ""
            btype = job.get("type") or ""
            provider = job.get("provider") or ""
            status = job.get("status") or ""
            created_at = job.get("created_at") or ""
            snapshot = job.get("snapshot") or {}
            snapshot_id = snapshot.get("snapshot_id") or ""
            snapshot_provider = (
                snapshot.get("provider") or snapshot.get("SnapshotProvider") or ""
            )
            snapshot_type = snapshot.get("type") or ""
            snapshot_name = snapshot.get("name") or ""

            cur.execute(
                """
                INSERT OR IGNORE INTO backups(backup_id, vm_id, dest, type, provider, status, created_at,
                    snapshot_id, snapshot_provider, snapshot_type, snapshot_name)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    backup_id,
                    vm_id,
                    dest,
                    btype,
                    provider,
                    status,
                    created_at,
                    snapshot_id,
                    snapshot_provider,
                    snapshot_type,
                    snapshot_name,
                ),
            )
            migrated_backups += 1
        conn.commit()

    conn.close()
    return {"migrated_backups": migrated_backups, "migrated_restores": 0}
