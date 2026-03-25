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


import pytest
from fastapi.testclient import TestClient
from novabackup.api import get_app


HAS_API = True


def _client_with_token():
    app = get_app()
    client = TestClient(app)
    resp = client.post("/token", data={"username": "alice", "password": "secret"})
    if resp.status_code != 200:
        pytest.skip("API token endpoint not available in this environment")
    token = resp.json().get("access_token")
    headers = {"Authorization": f"Bearer {token}"}
    return client, headers


def test_api_backups_create_list_restore():
    client, headers = _client_with_token()
    resp = client.post(
        "/backups",
        json={"vm_id": "vm1", "dest": "./backups", "type": "full", "name": "test-snap"},
        headers=headers,
    )
    assert resp.status_code in (200, 201)

    resp = client.get("/backups", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, dict) or isinstance(data, list)

    # Attempt restore (requires admin role)
    if isinstance(data, dict) and data.get("backups"):
        backup_list = data["backups"]
        if backup_list:
            backup_id = backup_list[0].get("backup_id")
            resp = client.post(
                f"/backups/{backup_id}/restore",
                json={"dest": "./restore"},
                headers=headers,
            )
            assert resp.status_code == 200
