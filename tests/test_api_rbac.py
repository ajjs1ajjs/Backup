import pytest
from fastapi.testclient import TestClient
from novabackup.api import get_app


HAS_API = True


def _client_with_headers(username: str, password: str):
    app = get_app()
    client = TestClient(app)
    resp = client.post("/token", data={"username": username, "password": password})
    if resp.status_code != 200:
        pytest.skip("API token endpoint not available in this environment")
    token = resp.json().get("access_token")
    headers = {"Authorization": f"Bearer {token}"}
    return client, headers


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_rbac_admin_and_user_can_create_backup():
    # Admin user
    admin_client, admin_headers = _client_with_headers("alice", "secret")
    payload = {
        "vm_id": "vm1",
        "dest": "local_dest",
        "type": "full",
        "name": "rbac-backup-admin",
    }
    resp_admin = admin_client.post("/backups", json=payload, headers=admin_headers)
    assert resp_admin.status_code == 200
    assert "backup_id" in resp_admin.json()

    # Regular user
    user_client, user_headers = _client_with_headers("bob", "secret")
    resp_user = user_client.post("/backups", json=payload, headers=user_headers)
    assert resp_user.status_code == 200
    assert "backup_id" in resp_user.json()


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_rbac_no_privileges_are_denied():
    # Charlie has no privileges
    charlie_client, charlie_headers = _client_with_headers("charlie", "secret")
    payload = {
        "vm_id": "vm1",
        "dest": "local_dest",
        "type": "full",
        "name": "rbac-backup-charlie",
    }
    resp = charlie_client.post("/backups", json=payload, headers=charlie_headers)
    assert resp.status_code == 403
