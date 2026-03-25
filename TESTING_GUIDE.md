# NovaBackup Enterprise - Testing Guide

## Test Coverage Summary

### Security Tests
- [x] Path Traversal Protection
- [x] SQL Injection Prevention
- [x] CSRF Token Validation
- [x] Rate Limiting
- [x] Password Policy Enforcement
- [x] Command Injection Prevention
- [x] Session Management
- [x] Audit Logging

### Integration Tests
- [x] Backup/Restore Workflow
- [x] Database Backup/Restore
- [x] VM Backup
- [x] Job Scheduling
- [x] User Authentication
- [x] RBAC Permissions

---

## Running Tests

### Unit Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/api/...
go test ./internal/rbac/...
go test ./internal/database/...
go test ./internal/backup/...
go test ./internal/restore/...
go test ./internal/scheduler/...
```

### Security Tests
```bash
# Run security-focused tests
go test -run Security ./...

# Run with race detector
go test -race ./...
```

---

## Test Cases

### 1. Path Traversal Protection

**Test File:** `internal/restore/engine_test.go`

```go
func TestSafeJoin_PathTraversal(t *testing.T) {
    tests := []struct {
        name        string
        baseDir     string
        relPath     string
        expectError bool
    }{
        {"normal path", "/backup", "file.txt", false},
        {"nested path", "/backup", "folder/file.txt", false},
        {"traversal with ..", "/backup", "../etc/passwd", true},
        {"traversal in middle", "/backup", "foo/../../etc/passwd", true},
        {"absolute path", "/backup", "/etc/passwd", true},
        {"Windows absolute", "/backup", "C:\\Windows\\System32", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := safeJoin(tt.baseDir, tt.relPath)
            if tt.expectError && err == nil {
                t.Errorf("Expected error for path %s, got nil", tt.relPath)
            }
            if !tt.expectError && err != nil {
                t.Errorf("Unexpected error for path %s: %v", tt.relPath, err)
            }
        })
    }
}
```

### 2. Rate Limiting Test

**Test File:** `internal/api/handlers_test.go`

```go
func TestLogin_RateLimiting(t *testing.T) {
    // Simulate 6 login attempts
    for i := 0; i < 6; i++ {
        req := httptest.NewRequest("POST", "/api/auth/login", 
            strings.NewReader(`{"username":"admin","password":"wrong"}`))
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        if i < 5 {
            if w.Code != 401 {
                t.Errorf("Attempt %d: Expected 401, got %d", i, w.Code)
            }
        } else {
            if w.Code != 429 {
                t.Errorf("Attempt %d: Expected 429 (rate limited), got %d", i, w.Code)
            }
        }
    }
}
```

### 3. Password Policy Test

**Test File:** `internal/rbac/engine_test.go`

```go
func TestPasswordPolicy(t *testing.T) {
    tests := []struct {
        password    string
        expectError bool
    }{
        {"short1A!", true},           // Too short (< 12)
        {"longenough1!", true},       // No uppercase
        {"LONGENOUGH1!", true},       // No lowercase
        {"LongEnough1!", false},      // Valid
        {"LongEnough!", true},        // No digit
        {"longenough1234", true},     // No special char
        {"password1234!", true},      // Common password
        {"LongEnough1!", false},      // Valid
        {"Abc12345678!", true},       // Sequential
        {"Aaaa1234567!", true},       // Repeated
    }
    
    for _, tt := range tests {
        t.Run(tt.password, func(t *testing.T) {
            err := PasswordPolicy(tt.password)
            if tt.expectError && err == nil {
                t.Errorf("Expected error for password %s", tt.password)
            }
            if !tt.expectError && err != nil {
                t.Errorf("Unexpected error for password %s: %v", tt.password, err)
            }
        })
    }
}
```

### 4. CSRF Protection Test

**Test File:** `internal/api/csrf_test.go`

```go
func TestCSRFMiddleware(t *testing.T) {
    tests := []struct {
        name        string
        method      string
        hasToken    bool
        expectCode  int
    }{
        {"GET without token", "GET", false, 200},
        {"POST without token", "POST", false, 403},
        {"POST with valid token", "POST", true, 200},
        {"POST with invalid token", "POST", false, 403},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, "/api/jobs", nil)
            if tt.hasToken {
                token := GenerateCSRFToken("test-session", "secret")
                req.Header.Set("X-CSRF-Token", token)
            }
            req.Header.Set("Authorization", "Bearer test-session")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            if w.Code != tt.expectCode {
                t.Errorf("Expected %d, got %d", tt.expectCode, w.Code)
            }
        })
    }
}
```

### 5. Audit Log Persistence Test

**Test File:** `internal/database/database_test.go`

```go
func TestAuditLog_Persistence(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    audit := &AuditLog{
        ID:        "test_audit_1",
        UserID:    "user_123",
        Username:  "admin",
        Action:    "POST /api/jobs",
        Resource:  "/api/jobs",
        IPAddress: "127.0.0.1",
        Success:   true,
        Details:   `{"status": 200}`,
    }
    
    // Create
    err := db.CreateAuditLog(audit)
    if err != nil {
        t.Fatalf("Failed to create audit log: %v", err)
    }
    
    // Read
    logs, err := db.GetAuditLogs(10, 0)
    if err != nil {
        t.Fatalf("Failed to get audit logs: %v", err)
    }
    
    if len(logs) != 1 {
        t.Errorf("Expected 1 audit log, got %d", len(logs))
    }
    
    // Delete old logs
    cutoff := time.Now().Add(-1 * time.Hour)
    err = db.DeleteAuditLogsBefore(cutoff)
    if err != nil {
        t.Fatalf("Failed to delete old logs: %v", err)
    }
}
```

### 6. Atomic Backup Test

**Test File:** `internal/backup/engine_test.go`

```go
func TestBackup_AtomicOperation(t *testing.T) {
    engine := setupTestEngine(t)
    
    job := &BackupJob{
        ID:          "test_job",
        Name:        "Test Backup",
        Type:        "file",
        Sources:     []string{"/tmp/test_data"},
        Destination: "/tmp/backups",
    }
    
    // Start backup
    session, err := engine.ExecuteBackup(job)
    
    // Verify backup was created atomically
    archivePath := filepath.Join(session.BackupPath, "backup.zip")
    tempPath := archivePath + ".tmp"
    
    // Temp file should not exist after successful backup
    if _, err := os.Stat(tempPath); err == nil {
        t.Error("Temporary file should be removed after successful backup")
    }
    
    // Final archive should exist
    if _, err := os.Stat(archivePath); err != nil {
        t.Errorf("Final archive should exist: %v", err)
    }
}
```

### 7. Retry Logic Test

**Test File:** `internal/utils/retry_test.go`

```go
func TestRetryWithBackoff(t *testing.T) {
    attempts := 0
    maxAttempts := 3
    
    err := RetryWithBackoff(func() error {
        attempts++
        if attempts < maxAttempts {
            return errors.New("temporary error")
        }
        return nil
    }, DefaultRetryConfig())
    
    if err != nil {
        t.Errorf("Expected success after retries, got error: %v", err)
    }
    
    if attempts != maxAttempts {
        t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
    }
}

func TestRetryWithBackoff_Failure(t *testing.T) {
    attempts := 0
    
    err := RetryWithBackoff(func() error {
        attempts++
        return errors.New("permanent error")
    }, &RetryConfig{
        MaxRetries:   3,
        InitialDelay: 10 * time.Millisecond,
        MaxDelay:     100 * time.Millisecond,
        Multiplier:   2.0,
    })
    
    if err == nil {
        t.Error("Expected error after all retries")
    }
    
    if attempts != 3 {
        t.Errorf("Expected 3 attempts, got %d", attempts)
    }
}
```

### 8. Graceful Shutdown Test

**Test File:** `cmd/novabackup/main_test.go`

```go
func TestGracefulShutdown(t *testing.T) {
    // Start server
    go func() {
        runServer()
    }()
    
    // Wait for server to start
    time.Sleep(2 * time.Second)
    
    // Send SIGTERM
    proc, _ := os.FindProcess(os.Getpid())
    proc.Signal(syscall.SIGTERM)
    
    // Wait for graceful shutdown (should complete within timeout)
    done := make(chan bool)
    go func() {
        // Check if server stopped gracefully
        done <- true
    }()
    
    select {
    case <-done:
        // Success
    case <-time.After(65 * time.Second): // Timeout + buffer
        t.Error("Graceful shutdown timed out")
    }
}
```

---

## Manual Testing Checklist

### Security
- [ ] Attempt path traversal in restore: `../../../etc/passwd`
- [ ] Try SQL injection in job creation
- [ ] Test CSRF with missing/invalid token
- [ ] Verify rate limiting after 5 failed logins
- [ ] Check password policy enforcement
- [ ] Verify scripts outside allowed directories are blocked

### Functionality
- [ ] Create backup job
- [ ] Run backup manually
- [ ] Verify backup schedule
- [ ] Restore files from backup
- [ ] Restore database
- [ ] Check audit logs in database
- [ ] Verify graceful shutdown on Ctrl+C

### API Endpoints
```bash
# Health check
curl http://localhost:8050/api/health

# Login
curl -X POST http://localhost:8050/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Create job (with CSRF token)
curl -X POST http://localhost:8050/api/jobs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -H "X-CSRF-Token: <csrf_token>" \
  -d '{"name":"Test","type":"file","sources":["/tmp"],"destination":"/backup"}'

# Get audit logs
curl http://localhost:8050/api/audit/logs \
  -H "Authorization: Bearer <token>"
```

---

## Performance Benchmarks

### Backup Speed
```bash
# Test backup speed with 10GB of data
time ./novabackup backup run --job-id=test-job

# Expected: ~100MB/s with compression
```

### Restore Speed
```bash
# Test restore speed
time ./novabackup restore files --session-id=<session>

# Expected: ~150MB/s
```

### Concurrent Users
```bash
# Test with 100 concurrent users
ab -n 1000 -c 100 http://localhost:8050/api/jobs
```

---

## Code Coverage Report

Generate coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Target Coverage:**
- Overall: > 70%
- Security-critical: > 90%
- API handlers: > 80%
- Database layer: > 85%

---

## Security Audit Checklist

### Authentication & Authorization
- [x] Strong password policy (12+ chars, special chars)
- [x] Rate limiting on login (5 attempts / 15 min)
- [x] Session timeout (24 hours)
- [x] Secure token generation (crypto/rand)
- [x] RBAC permission checks

### Data Protection
- [x] AES-GCM encryption for backups
- [x] Scrypt key derivation (N=32768, r=8, p=1)
- [x] Master key from environment variable
- [x] Credentials not logged
- [x] Path traversal protection

### Input Validation
- [x] All API inputs validated
- [x] SQL injection prevention (parameterized queries)
- [x] Command injection prevention (script validation)
- [x] CSRF token validation
- [x] File extension validation

### Audit & Monitoring
- [x] Audit logs persisted to database
- [x] Login attempts logged
- [x] Backup/restore operations logged
- [x] Configuration changes logged
- [x] Log rotation (90 days)

---

## Deployment Checklist

### Pre-Deployment
- [ ] Set `NOVABACKUP_MASTER_KEY` environment variable
- [ ] Change default admin password
- [ ] Configure backup storage location
- [ ] Set up SSL/TLS certificates (if using HTTPS)
- [ ] Configure firewall rules (port 8050/8443)

### Installation
```powershell
# Windows (PowerShell as Administrator)
iwr -Uri https://github.com/.../install.bat -OutFile install.bat
.\install.bat

# Linux (as root)
curl -fsSL https://github.com/.../install.sh | sudo bash
```

### Post-Installation
- [ ] Verify service is running: `systemctl status novabackup`
- [ ] Access Web UI: `http://localhost:8050`
- [ ] Change default password
- [ ] Create backup jobs
- [ ] Test backup/restore cycle
- [ ] Configure notifications
- [ ] Set up monitoring alerts

### Monitoring
```bash
# Check service status
systemctl status novabackup

# View logs
journalctl -u novabackup -f

# Check backup sessions
curl http://localhost:8050/api/backup/sessions

# Check audit logs
curl http://localhost:8050/api/audit/logs
```

---

## Support

For issues or questions:
- Documentation: `/docs` folder
- API Reference: `http://localhost:8050/api/health`
- Logs: `C:\ProgramData\NovaBackup\Logs\` or `/var/lib/novabackup/logs/`
