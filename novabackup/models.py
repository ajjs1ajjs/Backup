from datetime import datetime
from typing import List, Optional, Annotated
import re

from pydantic import BaseModel, Field, field_validator, model_validator


class VMModel(BaseModel):
    id: str = Field(..., min_length=1, max_length=256, description="VM identifier")
    name: str = Field(..., min_length=1, max_length=256, description="VM name")
    type: str = Field(..., min_length=1, max_length=64, description="VM type")
    status: str = Field(..., pattern="^(running|stopped|paused|unknown)$", description="VM status")


class NormalizeResponse(BaseModel):
    normalized: str = Field(..., min_length=1, max_length=64)


class SnapshotModel(BaseModel):
    snapshot_id: str = Field(..., min_length=1, max_length=256, description="Snapshot identifier")
    provider: str = Field(..., min_length=1, max_length=64, description="Provider name")
    vm_id: str = Field(..., min_length=1, max_length=256, description="VM identifier")
    dest: str = Field(..., min_length=1, max_length=1024, description="Destination path")
    type: str = Field(..., pattern="^(full|incremental|differential)$", description="Backup type")
    name: Optional[str] = Field(None, max_length=256, description="Snapshot name")
    status: str = Field(..., pattern="^(created|pending|completed|failed|restored)$", description="Snapshot status")


class BackupModel(BaseModel):
    backup_id: str = Field(..., min_length=1, max_length=256, description="Backup identifier")
    vm_id: str = Field(..., min_length=1, max_length=256, description="VM identifier")
    dest: str = Field(..., min_length=1, max_length=1024, description="Destination path")
    type: str = Field(..., pattern="^(full|incremental|differential)$", description="Backup type")
    snapshot: SnapshotModel
    provider: str = Field(..., min_length=1, max_length=64, description="Provider name")
    status: str = Field(..., pattern="^(created|pending|completed|failed|restored)$", description="Backup status")
    created_at: str = Field(..., description="ISO 8601 timestamp")

    @field_validator('created_at')
    @classmethod
    def validate_created_at(cls, v: str) -> str:
        """Validate that created_at is a valid ISO 8601 timestamp."""
        try:
            datetime.fromisoformat(v.replace('Z', '+00:00'))
        except ValueError:
            raise ValueError('created_at must be a valid ISO 8601 timestamp')
        return v


class BackupsResponse(BaseModel):
    backups: List[BackupModel]


class BackupCreateRequest(BaseModel):
    vm_id: str = Field(..., min_length=1, max_length=256, description="VM identifier to backup")
    dest: str = Field(..., min_length=1, max_length=1024, description="Destination path for backup")
    type: str = Field(default="full", pattern="^(full|incremental|differential)$", description="Backup type")
    name: Optional[str] = Field(None, max_length=256, description="Backup name")
    destination_type: str = Field(default="local", pattern="^(local|cloud)$", description="Destination type")
    cloud_provider: Optional[str] = Field(None, max_length=64, description="Cloud provider name")
    cloud_region: Optional[str] = Field(None, max_length=128, description="Cloud region")
    cloud_dest: Optional[str] = Field(None, max_length=1024, description="Cloud destination path")

    @model_validator(mode='after')
    def validate_cloud_backup(self) -> 'BackupCreateRequest':
        """Validate cloud backup specific fields."""
        if self.destination_type == "cloud":
            if not self.cloud_provider:
                raise ValueError("cloud_provider is required for cloud backups")
            if not self.cloud_dest:
                raise ValueError("cloud_dest is required for cloud backups")
        return self

    @field_validator('vm_id')
    @classmethod
    def validate_vm_id(cls, v: str) -> str:
        """Validate VM ID format."""
        if not v.strip():
            raise ValueError("vm_id cannot be empty or whitespace")
        return v.strip()

    @field_validator('dest')
    @classmethod
    def validate_dest(cls, v: str) -> str:
        """Validate destination path."""
        if not v.strip():
            raise ValueError("dest cannot be empty or whitespace")
        # Check for valid path or URL patterns
        v = v.strip()
        if not (v.startswith('/') or v.startswith('./') or v.startswith('../') or 
                v.startswith('s3://') or v.startswith('gs://') or v.startswith('azure://')):
            raise ValueError("dest must be a valid path or cloud URL")
        return v


class RestoreRequest(BaseModel):
    dest: str = Field(..., min_length=1, max_length=1024, description="Destination path for restore")

    @field_validator('dest')
    @classmethod
    def validate_dest(cls, v: str) -> str:
        """Validate destination path."""
        if not v.strip():
            raise ValueError("dest cannot be empty or whitespace")
        return v.strip()


class RestoreResponse(BaseModel):
    backup_id: str = Field(..., min_length=1, max_length=256, description="Backup identifier")
    status: str = Field(..., pattern="^(restored|failed|pending)$", description="Restore status")
    restored_at: Optional[str] = Field(None, description="ISO 8601 timestamp of restore")

    @field_validator('restored_at')
    @classmethod
    def validate_restored_at(cls, v: Optional[str]) -> Optional[str]:
        """Validate that restored_at is a valid ISO 8601 timestamp."""
        if v is None:
            return v
        try:
            datetime.fromisoformat(v.replace('Z', '+00:00'))
        except ValueError:
            raise ValueError('restored_at must be a valid ISO 8601 timestamp')
        return v
