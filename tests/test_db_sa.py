import os
import json
import sqlite3
from novabackup.backup import BackupManager


def test_db_sa_backend_via_sqlite_url(tmp_path):
    db_path = tmp_path / "test_novabackup.db"
    url = f"sqlite:///{db_path}"
    os.environ["NOVABACKUP_DATABASE_URL"] = url
    try:
        bm = BackupManager(database_url=url)
        job = bm.create_backup(
            vm_id="vm1", dest="./backups", backup_type="full", snapshot_name="test"
        )
        assert "backup_id" in job
        assert bm.list_backups()
        res = bm.restore_backup(job["backup_id"], dest="./restore")
        assert res["backup_id"] == job["backup_id"]
    finally:
        del os.environ["NOVABACKUP_DATABASE_URL"]
