import os
from typing import List, Optional
from datetime import timedelta
import asyncio
import json

from fastapi import FastAPI, Depends, HTTPException, status, Response, WebSocket, WebSocketDisconnect
from fastapi.security import OAuth2PasswordRequestForm, OAuth2PasswordBearer
from fastapi.middleware.cors import CORSMiddleware
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
    require_scope,
    revoke_token,
    get_audit_logs,
    audit_log,
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
from novabackup.notifications import (
    get_notification_manager,
    setup_notifications_from_env,
    notify,
)
from novabackup.scheduler import (
    get_scheduler,
    create_scheduled_backup,
    ScheduledJob,
)


class Token(BaseModel):
    access_token: Optional[str] = None
    token_type: str = "bearer"
    refresh_token: Optional[str] = None


def get_app():
    app = FastAPI(title="Novabackup API")

    # Setup notifications from environment
    setup_notifications_from_env()

    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # In production, specify exact origins
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # WebSocket connection manager for real-time updates
    class ConnectionManager:
        def __init__(self):
            self.active_connections: List[WebSocket] = []

        async def connect(self, websocket: WebSocket):
            await websocket.accept()
            self.active_connections.append(websocket)

        def disconnect(self, websocket: WebSocket):
            if websocket in self.active_connections:
                self.active_connections.remove(websocket)

        async def broadcast(self, message: dict):
            message_text = json.dumps(message)
            for connection in self.active_connections:
                try:
                    await connection.send_text(message_text)
                except:
                    pass

    manager = ConnectionManager()

    @app.websocket("/ws")
    async def websocket_endpoint(websocket: WebSocket):
        await manager.connect(websocket)
        try:
            while True:
                # Keep connection alive
                data = await websocket.receive_text()
                # Optionally handle messages from client
        except WebSocketDisconnect:
            manager.disconnect(websocket)

    @app.get("/ws/status")
    async def get_ws_status():
        """Get WebSocket connection status."""
        return {"connected_clients": len(manager.active_connections)}

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

    @app.get("/audit/logs")
    @track_requests(method="GET", endpoint="/audit/logs")
    async def get_logs(
        limit: int = 100,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Get recent audit logs (admin only)."""
        audit_log("audit_logs_accessed", current_user.get("username", "unknown"))
        return {"logs": get_audit_logs(limit)}

    @app.post("/auth/logout")
    @track_requests(method="POST", endpoint="/auth/logout")
    async def logout(
        token: str = Depends(oauth2_scheme),
        current_user: dict = Depends(get_current_user),
    ):
        """Logout user by revoking their token."""
        revoke_token(token)
        audit_log("logout", current_user.get("username", "unknown"))
        return {"message": "Successfully logged out"}

    @app.get("/auth/me")
    @track_requests(method="GET", endpoint="/auth/me")
    async def get_me(current_user: dict = Depends(get_current_user)):
        """Get current user information."""
        return {
            "username": current_user.get("username"),
            "roles": current_user.get("roles", []),
            "scopes": current_user.get("scopes", []),
        }

    # ==================== Notifications ====================
    
    @app.post("/notifications/test")
    @track_requests(method="POST", endpoint="/notifications/test")
    async def test_notification(
        message: str = "Test notification from NovaBackup",
        level: str = "info",
        channels: Optional[List[str]] = None,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Send test notification."""
        await notify(message, level, channels)
        return {"message": "Notification sent", "level": level}
    
    @app.get("/notifications/history")
    @track_requests(method="GET", endpoint="/notifications/history")
    async def get_notification_history(
        limit: int = 100,
        current_user: dict = Depends(get_current_user),
    ):
        """Get notification history."""
        manager = get_notification_manager()
        return {"notifications": manager.get_history(limit)}
    
    @app.get("/notifications/channels")
    @track_requests(method="GET", endpoint="/notifications/channels")
    async def get_notification_channels(
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Get registered notification channels."""
        manager = get_notification_manager()
        return {"channels": list(manager.channels.keys())}
    
    # ==================== Scheduler ====================
    
    @app.get("/scheduler/jobs")
    @track_requests(method="GET", endpoint="/scheduler/jobs")
    async def list_scheduled_jobs(
        enabled_only: bool = False,
        current_user: dict = Depends(get_current_user),
    ):
        """List all scheduled backup jobs."""
        scheduler = get_scheduler()
        jobs = scheduler.list_jobs(enabled_only)
        return {"jobs": [job.to_dict() for job in jobs]}
    
    @app.get("/scheduler/jobs/{job_id}")
    @track_requests(method="GET", endpoint="/scheduler/jobs/{job_id}")
    async def get_scheduled_job(
        job_id: str,
        current_user: dict = Depends(get_current_user),
    ):
        """Get a specific scheduled job."""
        scheduler = get_scheduler()
        job = scheduler.get_job(job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        return job.to_dict()
    
    @app.post("/scheduler/jobs")
    @track_requests(method="POST", endpoint="/scheduler/jobs")
    async def create_scheduled_job(
        job_data: dict,
        current_user: dict = Depends(require_role(["admin", "user"])),
    ):
        """Create a new scheduled backup job."""
        job = create_scheduled_backup(
            name=job_data.get("name", "Scheduled Backup"),
            vm_id=job_data.get("vm_id"),
            schedule_type=job_data.get("schedule_type", "interval"),
            schedule_value=str(job_data.get("schedule_value", "3600")),
            backup_type=job_data.get("backup_type", "full"),
            destination_type=job_data.get("destination_type", "local"),
            cloud_provider=job_data.get("cloud_provider"),
            cloud_region=job_data.get("cloud_region"),
            cloud_dest=job_data.get("cloud_dest"),
        )
        return {"job": job.to_dict(), "message": "Scheduled job created"}
    
    @app.delete("/scheduler/jobs/{job_id}")
    @track_requests(method="DELETE", endpoint="/scheduler/jobs/{job_id}")
    async def delete_scheduled_job(
        job_id: str,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Delete a scheduled job."""
        scheduler = get_scheduler()
        if not scheduler.remove_job(job_id):
            raise HTTPException(status_code=404, detail="Job not found")
        return {"message": "Job deleted"}
    
    @app.post("/scheduler/jobs/{job_id}/enable")
    @track_requests(method="POST", endpoint="/scheduler/jobs/{job_id}/enable")
    async def enable_scheduled_job(
        job_id: str,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Enable a scheduled job."""
        scheduler = get_scheduler()
        if not scheduler.enable_job(job_id):
            raise HTTPException(status_code=404, detail="Job not found")
        return {"message": "Job enabled"}
    
    @app.post("/scheduler/jobs/{job_id}/disable")
    @track_requests(method="POST", endpoint="/scheduler/jobs/{job_id}/disable")
    async def disable_scheduled_job(
        job_id: str,
        current_user: dict = Depends(require_role(["admin"])),
    ):
        """Disable a scheduled job."""
        scheduler = get_scheduler()
        if not scheduler.disable_job(job_id):
            raise HTTPException(status_code=404, detail="Job not found")
        return {"message": "Job disabled"}
    
    @app.get("/scheduler/status")
    @track_requests(method="GET", endpoint="/scheduler/status")
    async def get_scheduler_status(
        current_user: dict = Depends(get_current_user),
    ):
        """Get scheduler status."""
        scheduler = get_scheduler()
        return {
            "running": scheduler.running,
            "total_jobs": len(scheduler.jobs),
            "enabled_jobs": len([j for j in scheduler.jobs.values() if j.enabled]),
        }

    return app
