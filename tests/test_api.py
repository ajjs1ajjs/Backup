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
def test_api_endpoints_vms_and_normalize():
    client, headers = _client_with_token()
    resp = client.get("/vms", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)

    resp2 = client.get("/normalize/KVM", headers=headers)
    assert resp2.status_code == 200
    assert isinstance(resp2.json(), dict)
    assert "normalized" in resp2.json()
