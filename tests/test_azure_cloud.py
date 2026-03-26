from unittest.mock import MagicMock

from novabackup.azure_real import AzureCloudProvider


def test_azure_cloud_provider_list_and_backup():
    # Create provider with a dummy subscription id and no credential (will cause client to be None initially)
    provider = AzureCloudProvider(subscription_id="xxx")
    # Override the client with a mock
    mock_client = MagicMock()
    provider.client = mock_client

    # Setup mock VM list
    mock_vm = MagicMock()
    mock_vm.id = "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/test-vm"
    mock_vm.name = "test-vm"
    mock_client.virtual_machines.list_all.return_value = [mock_vm]

    # Test list_vms
    vms = provider.list_vms()
    assert isinstance(vms, list)
    assert len(vms) == 1
    assert vms[0]["id"] == mock_vm.id
    assert vms[0]["name"] == mock_vm.name
    assert vms[0]["type"] == "Azure"
    assert vms[0]["status"] == "running"

    # Test backup_to_cloud (it doesn't actually call Azure, just returns a dict)
    resp = provider.backup_to_cloud(
        vm_id="test-vm",
        cloud_provider="Azure",
        region="eastus",
        dest="https://account.blob.core.windows.net/container",
        backup_type="full",
        snapshot_name="test-snap",
    )
    assert isinstance(resp, dict)
    assert resp["provider"] == "Azure"
    assert resp["status"] == "created"
    assert "backup_id" in resp
    assert "snapshot_id" in resp
    assert "created_at" in resp
    assert resp["dest"] == "https://account.blob.core.windows.net/container"

    # Test restore_from_cloud
    restore_resp = provider.restore_from_cloud(
        backup_id="azure-backup-123",
        dest="https://account.blob.core.windows.net/container/restore",
    )
    assert isinstance(restore_resp, dict)
    assert restore_resp["provider"] == "Azure"
    assert restore_resp["status"] == "restored"
    assert restore_resp["backup_id"] == "azure-backup-123"
    assert (
        restore_resp["dest"]
        == "https://account.blob.core.windows.net/container/restore"
    )
    assert "restored_at" in restore_resp
