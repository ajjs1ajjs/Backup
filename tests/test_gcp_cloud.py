import os
from unittest.mock import MagicMock, patch

from novabackup.gcp_real import GCPCloudProvider


@patch("novabackup.gcp_real.build")
def test_gcp_cloud_provider_list_and_backup(mock_build):
    # Setup mock service
    mock_service = MagicMock()
    mock_build.return_value = mock_service

    # Setup aggregatedList response
    mock_aggregated_list = MagicMock()
    mock_service.instances().aggregatedList.return_value = mock_aggregated_list

    # Setup the execute return value
    mock_aggregated_list.execute.return_value = {
        "items": {
            "zone1": {
                "instances": [{"id": "123", "name": "test-gcp-vm", "status": "RUNNING"}]
            }
        }
    }

    # Create provider with a dummy project id
    provider = GCPCloudProvider(project_id="test-project")
    # Note: the real provider would have set up the service, but we mocked it.

    # Test list_vms
    vms = provider.list_vms()
    assert isinstance(vms, list)
    assert len(vms) == 1
    assert vms[0]["id"] == "123"
    assert vms[0]["name"] == "test-gcp-vm"
    assert vms[0]["type"] == "GCP"
    assert vms[0]["status"] == "RUNNING"

    # Test backup_to_cloud (it doesn't actually call GCP, just returns a dict)
    resp = provider.backup_to_cloud(
        vm_id="test-gcp-vm",
        cloud_provider="GCP",
        region="us-central1",
        dest="gs://my-bucket/backups",
        backup_type="full",
        snapshot_name="test-snap",
    )
    assert isinstance(resp, dict)
    assert resp["provider"] == "GCP"
    assert resp["status"] == "created"
    assert "backup_id" in resp
    assert "snapshot_id" in resp
    assert "created_at" in resp
    assert resp["dest"] == "gs://my-bucket/backups"

    # Test restore_from_cloud
    restore_resp = provider.restore_from_cloud(
        backup_id="gcp-backup-123", dest="gs://my-bucket/backups/restore"
    )
    assert isinstance(restore_resp, dict)
    assert restore_resp["provider"] == "GCP"
    assert restore_resp["status"] == "restored"
    assert restore_resp["backup_id"] == "gcp-backup-123"
    assert restore_resp["dest"] == "gs://my-bucket/backups/restore"
    assert "restored_at" in restore_resp
