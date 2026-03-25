import pytest

try:
    from fastapi.testclient import TestClient
    from novabackup.api import get_app

    HAS_API = True
except Exception:
    HAS_API = False


@pytest.mark.skipif(not HAS_API, reason="FastAPI not available; skipping API tests.")
def test_api_endpoints_vms_and_normalize():
    app = get_app()
    client = TestClient(app)
    resp = client.get("/vms")
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)

    resp2 = client.get("/normalize/KVM")
    assert resp2.status_code == 200
    assert isinstance(resp2.json(), dict)
    assert "normalized" in resp2.json()
