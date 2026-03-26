import pytest
from fastapi.testclient import TestClient
from novabackup.api import get_app


HAS_API = True


def _client_with_token(username: str, password: str):
    app = get_app()
    client = TestClient(app)
    resp = client.post("/token", data={"username": username, "password": password})
    if resp.status_code != 200:
        pytest.skip("API token endpoint not available in this environment")
    token = resp.json().get("access_token")
    headers = {"Authorization": f"Bearer {token}"}
    return client, headers


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_token_login_and_refresh_flow():
    app = get_app()
    client = TestClient(app)

    # Login as admin (alice)
    resp = client.post("/token", data={"username": "alice", "password": "secret"})
    assert resp.status_code == 200
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data

    # Refresh tokens
    refresh_resp = client.post(
        "/token/refresh",
        json={"access_token": None, "refresh_token": data["refresh_token"]},
    )
    assert refresh_resp.status_code == 200
    refresh_data = refresh_resp.json()
    assert "access_token" in refresh_data
    assert "refresh_token" in refresh_data


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_token_login_invalid_credentials():
    app = get_app()
    client = TestClient(app)
    resp = client.post("/token", data={"username": "alice", "password": "wrong"})
    assert resp.status_code == 401


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_protected_endpoint_requires_auth():
    app = get_app()
    client = TestClient(app)
    resp = client.get("/vms")
    assert resp.status_code == 401
