from fastapi.testclient import TestClient
from novabackup.api import get_app


def test_static_dashboard_served():
    app = get_app()
    client = TestClient(app)
    resp = client.get("/static/dashboard.html")
    assert resp.status_code == 200
    assert "Novabackup Dashboard" in resp.text
