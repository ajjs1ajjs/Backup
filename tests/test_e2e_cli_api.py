import json

from typer.testing import CliRunner
from fastapi.testclient import TestClient
from novabackup.cli import app
from novabackup.api import get_app


def test_e2e_cli_api_backup_flow():
    # CLI part: create a backup
    cli = CliRunner()
    result = cli.invoke(
        app,
        [
            "backup",
            "create",
            "--vm-id",
            "vm1",
            "--dest",
            "./backups",
            "--type",
            "full",
            "--name",
            "e2e-test",
        ],
    )
    assert result.exit_code == 0
    created = json.loads(result.stdout)
    backup_id = created.get("backup_id")
    assert backup_id is not None

    # API part: list backups and confirm backup exists
    api_app = get_app()
    client = TestClient(api_app)
    resp = client.get("/backups")
    assert resp.status_code == 200
    data = resp.json()
    backups = data.get("backups") or data
    assert isinstance(backups, list)
    assert any(b.get("backup_id") == backup_id for b in backups)

    # API part: restore backup via API
    resp = client.post(f"/backups/{backup_id}/restore", json={"dest": "./restore"})
    assert resp.status_code == 200
    restore_data = resp.json()
    assert restore_data.get("backup_id") == backup_id
    assert restore_data.get("status") in {"restored", "created"}

    # Final check: list backups again to ensure state updated
    resp = client.get("/backups")
    assert resp.status_code == 200
    data = resp.json()
    backups = data.get("backups") or data
    assert isinstance(backups, list)
    assert any(b.get("backup_id") == backup_id for b in backups)
