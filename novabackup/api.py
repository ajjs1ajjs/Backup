import os
from typing import List
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import APIKeyHeader
from novabackup.core import list_vms, normalize_vm_type
from novabackup.models import (
    VMModel,
    NormalizeResponse,
    BackupModel,
    BackupsResponse,
    BackupCreateRequest,
    RestoreRequest,
    RestoreResponse,
)
from novabackup.backup import BackupManager

# Optional API key authentication (can be disabled by not setting NOVABACKUP_API_KEY)
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


async def get_api_key(api_key: str = Depends(api_key_header)):
    if api_key is None:
        return None
    expected = os.environ.get("NOVABACKUP_API_KEY")
    if expected and api_key != expected:
        raise HTTPException(status_code=403, detail="Invalid API Key")
    return api_key


def get_app():
    app = FastAPI(title="Novabackup API")

    @app.get("/vms", response_model=List[VMModel])
    async def vms(_api: str = Depends(get_api_key)):
        vms = list_vms()
        return [VMModel(**vm) for vm in vms]

    @app.get("/normalize/{vm_type}", response_model=NormalizeResponse)
    async def normalize(vm_type: str, _api: str = Depends(get_api_key)):
        normalized = normalize_vm_type(vm_type)
        return NormalizeResponse(normalized=normalized)

    @app.post("/backups", response_model=BackupModel)
    async def backups_create(
        req: BackupCreateRequest, _api: str = Depends(get_api_key)
    ):
        manager = BackupManager()
        job = manager.create_backup(
            vm_id=req.vm_id, dest=req.dest, backup_type=req.type, snapshot_name=req.name
        )
        return BackupModel(**job)

    @app.get("/backups", response_model=BackupsResponse)
    async def backups_list(_api: str = Depends(get_api_key)):
        manager = BackupManager()
        backups = manager.list_backups()
        return BackupsResponse(backups=[BackupModel(**b) for b in backups])

    @app.post("/backups/{backup_id}/restore", response_model=RestoreResponse)
    async def backups_restore(
        backup_id: str, req: RestoreRequest, _api: str = Depends(get_api_key)
    ):
        manager = BackupManager()
        res = manager.restore_backup(backup_id, req.dest)
        return RestoreResponse(
            backup_id=backup_id,
            status=res.get("status", "unknown"),
            restored_at=res.get("restored_at"),
        )

    return app
