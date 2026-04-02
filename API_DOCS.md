# Backup System API Documentation

## Base URL
```
http://localhost:5000/api
```

## Authentication
All API requests require JWT token in Authorization header:
```
Authorization: Bearer <token>
```

## Endpoints

### Jobs

#### GET /api/jobs
Get all jobs
```
Response: {
  jobs: [...],
  total: 100,
  page: 1,
  pageSize: 20
}
```

#### POST /api/jobs
Create new job
```
Body: {
  "name": "Daily Backup",
  "jobType": "full_backup",
  "sourceId": "vm-1",
  "destinationId": "repo-1",
  "schedule": "0 2 * * *",
  "enabled": true
}
```

#### POST /api/jobs/{jobId}/run
Run job immediately

#### POST /api/jobs/{jobId}/stop
Stop running job

### Backups

#### GET /api/backups
Get all backups
```
Query params: jobId, repositoryId, page, pageSize
```

#### GET /api/backups/{backupId}
Get backup details

#### DELETE /api/backups/{backupId}
Delete backup

#### POST /api/backups/{backupId}/verify
Verify backup integrity

### Restore

#### POST /api/restore
Start restore
```
Body: {
  "backupId": "backup-123",
  "restoreType": "full_vm",
  "targetHost": "hyperv-host",
  "destinationPath": "C:\\VMs"
}
```

#### GET /api/restore/{restoreId}
Get restore progress

### Repositories

#### GET /api/repositories
Get all repositories

#### POST /api/repositories
Create repository

#### POST /api/repositories/{repositoryId}/test
Test repository connection

### Agents

#### GET /api/agents
Get all agents

#### DELETE /api/agents/{agentId}
Remove agent

### Reports

#### GET /api/reports/summary
Get summary statistics

#### GET /api/reports/activity
Get activity log

#### GET /api/reports/storage
Get storage report

### Settings

#### GET /api/settings
Get all settings

#### PUT /api/settings/{key}
Update setting

## WebSocket Events

### /ws/backup
Backup progress events
```json
{
  "type": "progress",
  "jobId": "job-123",
  "percent": 50,
  "speedMbps": 125.5,
  "bytesProcessed": 50000000000,
  "totalBytes": 100000000000
}
```

### /ws/agent
Agent status events
```json
{
  "type": "status",
  "agentId": 1,
  "status": "backing_up",
  "currentJob": "job-123"
}
```

## Error Responses

```json
{
  "error": "NotFound",
  "message": "Resource not found",
  "code": 404
}
```

## Rate Limits
- 100 requests per minute per IP
- 10 requests per second for heavy operations
