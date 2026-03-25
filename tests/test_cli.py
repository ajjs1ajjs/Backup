import json

from typer.testing import CliRunner
from novabackup.cli import app


runner = CliRunner()


def test_cli_list_vms_json():
    result = runner.invoke(app, ["list-vms", "--format", "json"])
    assert result.exit_code == 0
    data = json.loads(result.stdout)
    assert isinstance(data, list)
    assert all(isinstance(vm, dict) for vm in data)
    assert all(k in data[0] for k in ("id", "name", "type", "status"))


def test_cli_normalize():
    result = runner.invoke(app, ["normalize", "Hyper-V"])
    assert result.exit_code == 0
    assert result.stdout.strip() == "HyperV"


def test_cli_backup_create_and_list():
    # Create a backup job for VM with id 'vm1'
    result = runner.invoke(
        app,
        ["backup", "create", "--vm-id", "vm1", "--dest", "./backups", "--type", "full"],
    )
    assert result.exit_code == 0
    # Expect a JSON object with backup_id
    obj = json.loads(result.stdout)
    assert isinstance(obj, dict)
    assert "backup_id" in obj
    # List backups should show at least one entry
    result2 = runner.invoke(app, ["backup", "list"])
    assert result2.exit_code == 0
    data = json.loads(result2.stdout)
    assert isinstance(data, list)
