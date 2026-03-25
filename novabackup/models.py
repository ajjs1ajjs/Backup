from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel


class VMModel(BaseModel):
    id: str
    name: str
    type: str
    status: str


class NormalizeResponse(BaseModel):
    normalized: str


class SnapshotModel(BaseModel):
    snapshot_id: str
    provider: str
    vm_id: str
    dest: str
    type: str
    name: Optional[str] = None
    status: str


class BackupModel(BaseModel):
    backup_id: str
    vm_id: str
    dest: str
    type: str
    snapshot: SnapshotModel
    provider: str
    status: str
    created_at: str


class BackupsResponse(BaseModel):
    backups: List[BackupModel]


class BackupCreateRequest(BaseModel):
    vm_id: str
    dest: str
    type: str = "full"
    name: Optional[str] = None


class RestoreRequest(BaseModel):
    dest: str


class RestoreResponse(BaseModel):
    backup_id: str
    status: str
    restored_at: Optional[str] = None
