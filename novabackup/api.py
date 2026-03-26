import os
from typing import List, Optional
from datetime import timedelta

from fastapi import FastAPI, Depends, HTTPException, status, Response
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
from novabackup.observability import (
    REQUEST_COUNT,
    REQUEST_LATENCY,
    BACKUPS_CREATED,
    BACKUPS_RESTORED,
    track_requests,
    metrics_response,
    audit_log,
    logger,
)


class Token(BaseModel):
    access_token: Optional[str] = None
    token_type: str = "bearer"
    refresh_token: Optional[str] = None


def get_app():
    app = FastAPI(title="Novabackup API")

    @app.post("/token")
    @track_requests(method="POST", endpoint="/token")
    async def login(form_data: OAuth2PasswordRequestForm = Depends()):
        user = authenticate_user(form_data.username, form_data.password)
        if not user:
            audit_log(
                action="login_failed",
                user=form_data.username,
                details="invalid credentials",
            )
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
        audit_log(action="login_success", user=user["username"])
        return {
            "access_token": access_token,
            "token_type": "bearer",
            "refresh_token": refresh_token,
        }

    @app.post("/token/refresh")
    @track_requests(method="POST", endpoint="/token/refresh")
    async def token_refresh(req: Token):
        # We don't have the user from the token here, but we could decode it.
        # For simplicity, we'll just log the refresh action without user.
        audit_log(action="token_refresh", user="unknown")
        return refresh_access_token(req.refresh_token)

    @app.get("/vms", response_model=List[VMModel])
    @track_requests(method="GET", endpoint="/vms")
    async def vms(current_user: dict = Depends(get_current_user)):
        audit_log(action="list_vms", user=current_user.get("username", "unknown"))
        vms = list_vms()
        return [VMModel(**vm) for vm in vms]

    @app.get("/normalize/{vm_type}", response_model=NormalizeResponse)
    @track_requests(method="GET", endpoint="/normalize/{vm_type}")
    async def normalize(vm_type: str, current_user: dict = Depends(get_current_user)):
        audit_log(
            action="normalize_vm_type",
            user=current_user.get("username", "unknown"),
            details=f"vm_type={vm_type}",
        )
        normalized = normalize_vm_type(vm_type)
        return NormalizeResponse(normalized=normalized)

    @app.post("/backups", response_model=BackupModel)
    @track_requests(method="POST", endpoint="/backups")
    async def backups_create(
        req: BackupCreateRequest,
        current_user: dict = Depends(require_role(["admin", "user"])),
    ):
        audit_log(
            action="create_backup_start",
            user=current_user.get("username", "unknown"),
            details=f"req={req.dict()}",
        )
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
            audit_log(
                action="create_backup_success",
                user=current_user.get("username", "unknown"),
                details=f"backup_id={job.get('backup_id')}",
            )
            BACKUPS_CREATED.labels(
                destination_type="cloud", provider=req.cloud_provider or "unknown"
            ).inc()
        else:
            job = manager.create_backup(
                vm_id=req.vm_id,
                dest=req.dest,
                backup_type=req.type,
                snapshot_name=req.name,
            )
            audit_log(
                action="create_backup_success",
                user=current_user.get("username", "unknown"),
                details=f"backup_id={job.get('backup_id')}",
            )
            BACKUPS_CREATED.labels(destination_type="local", provider="local").inc()
        return BackupModel(**job)

    @app.get("/backups", response_model=BackupsResponse)
    @track_requests(method="GET", endpoint="/backups")
    async def backups_list(current_user: dict = Depends(get_current_user)):
        audit_log(action="list_backups", user=current_user.get("username", "unknown"))
        manager = BackupManager()
        backups = manager.list_backups()
        return BackupsResponse(backups=[BackupModel(**b) for b in backups])

    @app.post("/backups/{backup_id}/restore", response_model=RestoreResponse)
    @track_requests(method="POST", endpoint="/backups/{backup_id}/restore")
    async def backups_restore(
        backup_id: str,
        req: RestoreRequest,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        audit_log(
            action="restore_backup_start",
            user=current_user.get("username", "unknown"),
            details=f"backup_id={backup_id}, dest={req.dest}",
        )
        manager = BackupManager()
        res = manager.restore_backup(backup_id, req.dest)
        audit_log(
            action="restore_backup_success",
            user=current_user.get("username", "unknown"),
            details=f"backup_id={backup_id}, status={res.get('status')}",
        )
        BACKUPS_RESTORED.labels(
            destination_type="local"
            if not req.dest.startswith(("s3://", "gs://", "azure://"))
            else "cloud"
        ).inc()
        return RestoreResponse(
            backup_id=backup_id,
            status=res.get("status", "unknown"),
            restored_at=res.get("restored_at"),
        )

    @app.get("/metrics")
    async def metrics():
        """Endpoint to expose Prometheus metrics."""
        return metrics_response()

    return app
