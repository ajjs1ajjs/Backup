import os
import pytest

from novabackup.db_sa import SA_DBManager


def test_sa_db_multidialect():
    dialect_urls = {
        "postgres": os.environ.get("NOVABACKUP_DATABASE_URL_POSTGRES"),
        "mssql": os.environ.get("NOVABACKUP_DATABASE_URL_MSSQL"),
        "oracle": os.environ.get("NOVABACKUP_DATABASE_URL_ORACLE"),
    }
    if not any(v for v in dialect_urls.values()):
        pytest.skip(
            "No database URLs provided for Postgres, MSSQL, or Oracle. Set environment variables to run dialect tests."
        )

    for dialect, url in dialect_urls.items():
        if not url:
            continue
        bm = SA_DBManager(url)
        job = bm.create_backup(
            vm_id=f"vm-{dialect}",
            dest=f"./backups-{dialect}",
            backup_type="full",
            snapshot_name=f"{dialect}-snap",
        )
        assert isinstance(job, dict) and job.get("backup_id") is not None
        backups = bm.list_backups()
        assert isinstance(backups, list)
        res = bm.restore_backup(job["backup_id"], dest=f"./restore-{dialect}")
        assert res.get("backup_id") == job["backup_id"]
