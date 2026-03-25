import os
from typing import List, Optional
from datetime import timedelta

from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm, OAuth2PasswordBearer
from pydantic import BaseModel

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
from novabackup.security import (
    authenticate_user,
    create_access_token,
    create_refresh_token,
    refresh_access_token,
    get_current_user,
    require_role,
)


class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"
    refresh_token: Optional[str] = None


def get_app():
    app = FastAPI(title="Novabackup API")

    @app.post("/token")
    async def login(form_data: OAuth2PasswordRequestForm = Depends()):
        user = authenticate_user(form_data.username, form_data.password)
        if not user:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Incorrect username or password",
            )
        access_token = create_access_token(
            {"sub": user["username"], "roles": user["roles"]},
            expires_delta=timedelta(minutes=60),
        )
        refresh_token = create_refresh_token(
            {"sub": user["username"], "roles": user["roles"]}
        )
        return {
            "access_token": access_token,
            "token_type": "bearer",
            "refresh_token": refresh_token,
        }

    @app.post("/token/refresh")
    async def token_refresh(req: Token):
        return refresh_access_token(req.refresh_token)

    @app.get("/vms", response_model=List[VMModel])
    async def vms(current_user: dict = Depends(get_current_user)):
        vms = list_vms()
        return [VMModel(**vm) for vm in vms]

    @app.get("/normalize/{vm_type}", response_model=NormalizeResponse)
    async def normalize(vm_type: str, current_user: dict = Depends(get_current_user)):
        normalized = normalize_vm_type(vm_type)
        return NormalizeResponse(normalized=normalized)

    @app.post("/backups", response_model=BackupModel)
    async def backups_create(
        req: BackupCreateRequest,
        current_user: dict = Depends(require_role(["admin", "user"])),
    ):
        manager = BackupManager()
        if getattr(req, "destination_type", "local").lower() == "cloud":
            job = manager.create_backup(
                vm_id=req.vm_id,
                dest=req.cloud_dest or req.dest,
                backup_type=req.type,
                snapshot_name=req.name,
                destination_type=req.destination_type,
                cloud_provider=req.cloud_provider,
                cloud_region=req.cloud_region,
                cloud_dest=req.cloud_dest,
            )
        else:
            job = manager.create_backup(
                vm_id=req.vm_id,
                dest=req.dest,
                backup_type=req.type,
                snapshot_name=req.name,
            )
        return BackupModel(**job)

    @app.get("/backups", response_model=BackupsResponse)
    async def backups_list(current_user: dict = Depends(get_current_user)):
        manager = BackupManager()
        backups = manager.list_backups()
        return BackupsResponse(backups=[BackupModel(**b) for b in backups])

    @app.post("/backups/{backup_id}/restore", response_model=RestoreResponse)
    async def backups_restore(
        backup_id: str,
        req: RestoreRequest,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        manager = BackupManager()
        res = manager.restore_backup(backup_id, req.dest)
        return RestoreResponse(
            backup_id=backup_id,
            status=res.get("status", "unknown"),
            restored_at=res.get("restored_at"),
        )

    return app
