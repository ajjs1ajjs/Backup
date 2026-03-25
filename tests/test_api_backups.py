import json
from fastapi.testclient import TestClient
from novabackup.api import get_app


def test_api_backups_crud():
    app = get_app()
    client = TestClient(app)

    # Create a backup
    resp = client.post(
        "/backups",
        json={"vm_id": "vm1", "dest": "./backups", "type": "full", "name": "test-snap"},
    )
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, dict)
    assert "backup_id" in data
    backup_id = data["backup_id"]

    # List backups
    resp = client.get("/backups")
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, dict) and "backups" in data
    assert isinstance(data["backups"], list)

    # Restore backup
    resp = client.post(f"/backups/{backup_id}/restore", json={"dest": "./restore"})
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, dict)
    assert data.get("backup_id") == backup_id
    assert "restored_at" in data or "restored_at" in data
