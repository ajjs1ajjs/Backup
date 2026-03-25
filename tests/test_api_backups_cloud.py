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


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_cloud_backup_flow():
    client, headers = _client_with_token()
    req = {
        "vm_id": "vm1",
        "dest": "./backups",
        "type": "full",
        "name": "cloud-test",
        "destination_type": "cloud",
        "cloud_provider": "AWS",
        "cloud_region": "us-east-1",
        "cloud_dest": "s3://my-bucket/backup",
    }
    resp = client.post("/backups", json=req, headers=headers)
    assert resp.status_code in (200, 201)
    data = resp.json()
    backup_id = data.get("backup_id")
    assert backup_id is not None

    resp = client.get("/backups", headers=headers)
    assert resp.status_code == 200
    _ = resp.json()

    resp = client.post(
        f"/backups/{backup_id}/restore",
        json={"dest": "s3://my-bucket/restore"},
        headers=headers,
    )
    assert resp.status_code == 200
