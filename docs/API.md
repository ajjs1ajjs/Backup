# NovaBackup Enterprise API Documentation

## Base URL
```
http://localhost:8050/api
```

## Authentication

### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your_password"
}

Response:
{
  "success": true,
  "token": "eyJhbGc...",
  "user": {
    "id": "user_123",
    "username": "admin",
    "role": "admin"
  }
}
```

### Logout
```http
POST /api/auth/logout
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Change Password
```http
POST /api/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "old_pass",
  "new_password": "NewSecureP@ss123!"
}
```

**Password Requirements:**
- Minimum 12 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one digit (0-9)
- At least one special character (!@#$%^&*)
- Cannot contain common passwords (password, admin123, etc.)
- Cannot contain sequential characters (abc, 123)
- Cannot contain repeated characters (aaa, 111)

---

## Security Headers

All state-changing requests (POST, PUT, DELETE, PATCH) require:

1. **Authorization Header**
```
Authorization: Bearer <your_session_token>
```

2. **CSRF Token** (for state-changing operations)
```
X-CSRF-Token: <csrf_token>
```

To get CSRF token:
1. Make a GET request to any endpoint
2. Extract token from `X-CSRF-Token` response header
3. Use this token for subsequent POST/PUT/DELETE requests

---

## Backup Jobs

### List Jobs
```http
GET /api/jobs
Authorization: Bearer <token>
```

### Create Job
```http
POST /api/jobs
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "name": "Daily Backup",
  "type": "file",
  "sources": [
    "C:\\Data",
    "D:\\Documents"
  ],
  "destination": "E:\\Backups",
  "compression": true,
  "compression_level": 5,
  "encryption": true,
  "encryption_key": "your_secret_key",
  "incremental": true,
  "full_backup_every": 7,
  "schedule": "daily",
  "schedule_time": "02:00",
  "retention_days": 30,
  "enabled": true
}
```

**Validation Rules:**
- `name`: Required, 1-100 characters
- `type`: Must be one of: `file`, `database`, `vm`, `cloud`
- `sources`: Required for file type, must be absolute paths
- `destination`: Required, must be absolute path
- `retention_days`: 1-3650
- `compression_level`: 0-9
- `schedule`: Must be valid schedule type or cron expression

### Update Job
```http
PUT /api/jobs/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "name": "Updated Backup",
  "enabled": false
}
```

### Delete Job
```http
DELETE /api/jobs/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Run Job
```http
POST /api/jobs/:id/run
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Stop Job
```http
POST /api/jobs/:id/stop
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

---

## Backup Sessions

### List Sessions
```http
GET /api/backup/sessions
Authorization: Bearer <token>
```

### Get Session
```http
GET /api/backup/sessions/:id
Authorization: Bearer <token>
```

### Browse Session Files
```http
GET /api/backup/sessions/:id/files
Authorization: Bearer <token>
```

### Verify Backup
```http
POST /api/backup/verify
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "session_id": "session_123"
}
```

### Get Verification History
```http
GET /api/backup/verifications
Authorization: Bearer <token>
```

### Get CBT Statistics
```http
GET /api/backup/cbt-stats
Authorization: Bearer <token>
```

---

## Restore

### List Restore Points
```http
GET /api/restore/points
Authorization: Bearer <token>
```

### Restore Files
```http
POST /api/restore/files
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "session_id": "session_123",
  "destination": "C:\\Restored",
  "files": [
    "folder/file1.txt",
    "folder/file2.doc"
  ],
  "overwrite": true
}
```

**Validation Rules:**
- `destination`: Must be absolute path
- `files`: Optional, if empty restores all files

### Restore Database
```http
POST /api/restore/database
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "session_id": "session_123",
  "db_type": "mysql",
  "conn_str": "server=localhost;user=root;password=pass;",
  "target_database": "restored_db"
}
```

### Instant Restore
```http
POST /api/restore/instant
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "session_id": "session_123",
  "vm_name": "restored_vm"
}
```

---

## Storage Repositories

### List Repositories
```http
GET /api/storage/repos
Authorization: Bearer <token>
```

### Create Repository
```http
POST /api/storage/repos
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "name": "NAS Storage",
  "type": "smb",
  "path": "\\\\nas\\backups",
  "credentials": {
    "username": "backup_user",
    "password": "secure_password"
  }
}
```

### Update Repository
```http
PUT /api/storage/repos/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Delete Repository
```http
DELETE /api/storage/repos/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

---

## Settings

### Get Settings
```http
GET /api/settings
Authorization: Bearer <token>
```

### Update Settings
```http
PUT /api/settings
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "server": {
    "ip": "0.0.0.0",
    "port": 8050,
    "https": false,
    "https_port": 8443
  },
  "notifications": {
    "email": {
      "smtp_server": "smtp.example.com",
      "smtp_port": 587,
      "from": "backup@example.com",
      "to": ["admin@example.com"]
    }
  },
  "retention": {
    "type": "days",
    "value": 30
  }
}
```

### Update Server Settings
```http
PUT /api/settings/server
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Update Directory Settings
```http
PUT /api/settings/directories
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Update Retention Settings
```http
PUT /api/settings/retention
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Update Notification Settings
```http
PUT /api/settings/notifications
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

---

## Users

### List Users
```http
GET /api/users
Authorization: Bearer <token>
```

### Get User
```http
GET /api/users/:id
Authorization: Bearer <token>
```

### Create User
```http
POST /api/users
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "NewSecureP@ss123!",
  "email": "user@example.com",
  "full_name": "New User",
  "role": "backup_user"
}
```

**Available Roles:**
- `admin` - Full access
- `backup_admin` - Manage backups and jobs
- `backup_user` - Run backups, view logs
- `readonly` - View only

### Update User
```http
PUT /api/users/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Delete User
```http
DELETE /api/users/:id
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Enable User
```http
POST /api/users/:id/enable
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Disable User
```http
POST /api/users/:id/disable
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

---

## Reports & Statistics

### Get Statistics
```http
GET /api/reports/statistics
Authorization: Bearer <token>
```

### Get Daily Report
```http
GET /api/reports/daily
Authorization: Bearer <token>
```

---

## Audit Logs

### Get Audit Logs
```http
GET /api/audit/logs
Authorization: Bearer <token>
Query Parameters:
  - limit: Number of records (default: 100)
  - offset: Pagination offset
  - user_id: Filter by user
  - action: Filter by action type
  - success: Filter by success status
```

**Response:**
```json
{
  "logs": [
    {
      "id": "audit_1234567890",
      "timestamp": "2026-03-25T10:30:00Z",
      "user_id": "user_123",
      "username": "admin",
      "action": "POST /api/jobs",
      "resource": "/api/jobs",
      "ip_address": "192.168.1.100",
      "success": true,
      "details": {
        "status": 201,
        "job_id": "job_456"
      }
    }
  ],
  "total": 150
}
```

---

## Database Management

### List Databases
```http
POST /api/database/list
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "server": "localhost",
  "db_type": "mysql",
  "credentials": {
    "username": "root",
    "password": "password"
  }
}
```

### Backup Database
```http
POST /api/database/backup
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "db_type": "mysql",
  "server": "localhost",
  "port": 3306,
  "database": "mydb",
  "credentials": {
    "username": "backup_user",
    "password": "secure_password"
  }
}
```

---

## VM Management

### List VMs
```http
POST /api/vm/list
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
```

### Backup VM
```http
POST /api/vm/backup
Authorization: Bearer <token>
X-CSRF-Token: <csrf_token>
Content-Type: application/json

{
  "vm_name": "Production-VM",
  "hyperv_host": "hyperv.example.com"
}
```

---

## Health Check

### Get Health Status
```http
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "7.0.0",
  "time": "2026-03-25T10:30:00Z"
}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message in Ukrainian",
  "code": "ERROR_CODE"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request format |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions or invalid CSRF token |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |

### Rate Limiting

Login endpoint is rate-limited:
- **5 attempts** per **15 minutes** per IP+username combination
- Lockout duration: **30 minutes**

**Response when rate limited:**
```json
{
  "error": "Забагато невдалих спроб входу. Спробуйте пізніше.",
  "retry_after": 900
}
```

---

## Security Best Practices

1. **Always use HTTPS** in production
2. **Change default passwords** immediately after installation
3. **Set strong master key** via `NOVABACKUP_MASTER_KEY` environment variable
4. **Regularly rotate** user passwords and API tokens
5. **Monitor audit logs** for suspicious activity
6. **Keep backups encrypted** with strong keys
7. **Restrict network access** to API endpoints
8. **Regular security updates**

---

## API Versioning

Current API version: **v1** (implicit)

Future versions will be prefixed: `/api/v2/...`

---

## Support

For API support, contact:
- Email: support@novabackup.local
- Documentation: https://docs.novabackup.local/api
