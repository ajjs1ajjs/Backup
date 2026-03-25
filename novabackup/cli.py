import json
from typing import Optional
import typer
import json
from .core import list_vms, normalize_vm_type
from .backup import BackupManager

app = typer.Typer(help="Novabackup Python MVP CLI.")


@app.command("list-vms")
def list_vms_cmd(
    format: Optional[str] = typer.Option("json", help="Output format: json or plain"),
):
    """List available VMs (mocked for MVP)."""
    vms = list_vms()
    if format == "json":
        typer.echo(json.dumps(vms, indent=2))
    else:
        for vm in vms:
            typer.echo(f"{vm['id']}\t{vm['name']}\t{vm['type']}\t{vm['status']}")


@app.command("normalize")
def normalize_cmd(vm_type: str = typer.Argument(..., help="VM type to normalize")):
    """Normalize a VM type to a canonical form."""
    norm = normalize_vm_type(vm_type)
    typer.echo(norm)


# Backup group (full CLI for creating/restoring VM backups)
backup_app = typer.Typer(help="Backup operations for VMs.")


@backup_app.command("create")
def backup_create(
    vm_id: str = typer.Option(..., "--vm-id", help="ID of the VM to back up"),
    dest: str = typer.Option(
        "./backups", "--dest", help="Destination folder for backups"
    ),
    backup_type: str = typer.Option(
        "full", "--type", help="Backup type: full or incremental"
    ),
    name: Optional[str] = typer.Option(None, "--name", help="Optional backup name"),
):
    """Create a backup for a VM."""
    manager = BackupManager()
    job = manager.create_backup(
        vm_id=vm_id, dest=dest, backup_type=backup_type, snapshot_name=name
    )
    typer.echo(json.dumps(job, indent=2))


@backup_app.command("list")
def backup_list():
    """List existing backups."""
    manager = BackupManager()
    jobs = manager.list_backups()
    typer.echo(json.dumps(jobs, indent=2))


@backup_app.command("restore")
def backup_restore(
    backup_id: str = typer.Option(..., "--backup-id", help="Backup ID to restore"),
    dest: str = typer.Option("./restore", "--dest", help="Destination for restored VM"),
):
    """Restore a backup to a destination."""
    manager = BackupManager()
    try:
        result = manager.restore_backup(backup_id, dest)
        typer.echo(json.dumps(result, indent=2))
    except Exception as e:
        typer.echo(str(e))


# Attach the backup commands under the main CLI namespace as 'backup'
app.add_typer(backup_app, name="backup")
