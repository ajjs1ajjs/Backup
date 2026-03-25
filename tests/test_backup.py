import json
from novabackup.backup import BackupManager


def test_backup_manager_create_and_list_and_restore(tmp_path):
    bm = BackupManager(storage_path=str(tmp_path / "backups.json"))
    # Create a backup
    job = bm.create_backup(
        vm_id="vm1", dest="./backups", backup_type="full", snapshot_name="test-snap"
    )
    assert "backup_id" in job
    assert job["vm_id"] == "vm1"

    # List backups should include the created one
    all_jobs = bm.list_backups()
    assert isinstance(all_jobs, list)
    assert any(j["backup_id"] == job["backup_id"] for j in all_jobs)

    # Restore backup
    res = bm.restore_backup(job["backup_id"], dest="./restore")
    assert res["backup_id"] == job["backup_id"]
    assert res["status"] == "restored"
