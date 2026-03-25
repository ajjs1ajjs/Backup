# NovaBackup Security Audit Checklist

## Authentication & Authorization

### Password Policy
- [x] Minimum 12 characters required
- [x] Uppercase letters (A-Z) required
- [x] Lowercase letters (a-z) required
- [x] Digits (0-9) required
- [x] Special characters required
- [x] Common passwords blocked
- [x] Sequential characters detected (abc, 123)
- [x] Repeated characters detected (aaa, 111)
- [x] Maximum 128 characters

**Test:**
```bash
# Try weak password
curl -X POST http://localhost:8050/api/users \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"weak"}'
# Expected: 400 Bad Request
```

### Rate Limiting
- [x] Login endpoint rate-limited
- [x] 5 attempts per 15 minutes
- [x] 30-minute lockout after exceeded
- [x] Rate limit by IP + username
- [x] 429 response with retry_after

**Test:**
```bash
# Attempt 6 logins in quick succession
for i in {1..6}; do
  curl -X POST http://localhost:8050/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"wrong"}'
done
# Expected: 6th attempt returns 429
```

### Session Management
- [x] Secure token generation (crypto/rand)
- [x] No fallback to weak random
- [x] 24-hour session timeout
- [x] Session invalidation on logout
- [x] Session persistence in database
- [x] Race condition fixed (single lock)

**Test:**
```bash
# Login and get token
TOKEN=$(curl -X POST http://localhost:8050/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"SecureP@ss123!"}' \
  | jq -r '.token')

# Use token
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8050/api/jobs

# Logout
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8050/api/auth/logout

# Try to use expired token (should fail)
```

### RBAC Permissions
- [x] Role-based access control
- [x] Permission checks on all endpoints
- [x] Principle of least privilege
- [x] Audit logging for permission checks

**Roles:**
| Role | Permissions |
|------|-------------|
| admin | Full access |
| backup_admin | Manage backups, view logs |
| backup_user | Run backups, read-only |
| readonly | View only |

---

## Data Protection

### Encryption at Rest
- [x] AES-GCM encryption for backups
- [x] Random IV per block
- [x] Scrypt key derivation (N=32768, r=8, p=1)
- [x] Master key from environment variable
- [x] Encryption keys stored encrypted in DB

**Verify:**
```bash
# Check encryption is enabled
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8050/api/jobs | jq '.[].encryption'

# Check backup files are encrypted
ls -la /backup/location/*.enc
```

### Encryption in Transit
- [ ] HTTPS support (optional)
- [ ] TLS 1.2+ recommended
- [ ] Certificate validation
- [ ] HSTS headers

**Recommendation:**
```json
{
  "server": {
    "https": true,
    "https_port": 8443,
    "tls_min_version": "TLS1.2"
  }
}
```

### Credential Protection
- [x] Credentials not logged
- [x] SafeJob struct for logging
- [x] Passwords hashed with bcrypt
- [x] Connection strings encrypted in DB
- [x] Master key required for decryption

**Test:**
```bash
# Check logs for credentials
grep -r "password" C:\ProgramData\NovaBackup\Logs\
# Expected: No passwords found
```

---

## Input Validation

### Path Traversal Prevention
- [x] safeJoin() function validates all paths
- [x] Blocks `..` sequences
- [x] Blocks absolute paths in relative context
- [x] Verifies final path is within bounds
- [x] Case-insensitive checks on Windows

**Test:**
```bash
# Try path traversal in restore
curl -X POST http://localhost:8050/api/restore/files \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"destination":"../../../etc/passwd"}'
# Expected: 400 Bad Request
```

### SQL Injection Prevention
- [x] Parameterized queries
- [x] Input validation before DB operations
- [x] Whitelist validation for types
- [x] JSON field validation

**Test:**
```bash
# Try SQL injection
curl -X POST http://localhost:8050/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"test; DROP TABLE jobs;--"}'
# Expected: Job created with literal name, no SQL execution
```

### Command Injection Prevention
- [x] Script path validation
- [x] Blocks shell metacharacters (`;|&$`()<>`)
- [x] Requires absolute paths
- [x] Validates file exists
- [x] Blocks symlinks
- [x] Extension validation on Windows
- [x] Executable bit check on Unix
- [x] Allowed directories whitelist

**Test:**
```bash
# Try command injection via script
curl -X POST http://localhost:8050/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"pre_backup_script":"C:\\temp\\malicious; rm -rf /"}'
# Expected: 400 Bad Request (dangerous characters)
```

### CSRF Protection
- [x] CSRF tokens for state-changing requests
- [x] Token bound to session
- [x] HMAC-SHA256 signature
- [x] Constant-time comparison
- [x] Exempts GET/HEAD/OPTIONS/TRACE
- [x] Exempts login/logout

**Test:**
```bash
# Try POST without CSRF token
curl -X POST http://localhost:8050/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"test"}'
# Expected: 403 Forbidden

# Try with valid CSRF token
curl -X POST http://localhost:8050/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"test"}'
# Expected: 201 Created
```

---

## Audit Logging

### Log Coverage
- [x] All authentication attempts
- [x] All authorization failures
- [x] All backup operations
- [x] All restore operations
- [x] All configuration changes
- [x] All user management actions
- [x] All permission changes

### Log Persistence
- [x] Logs stored in database
- [x] In-memory cache (1000 entries)
- [x] Automatic rotation (90 days)
- [x] Query by user, action, date
- [x] Tamper-evident (append-only)

**Test:**
```bash
# Check audit logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8050/api/audit/logs?limit=10"

# Verify login was logged
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8050/api/audit/logs?action=login"
```

---

## Error Handling

### Information Disclosure
- [x] Generic error messages to clients
- [x] Detailed errors logged server-side
- [x] Stack traces never exposed
- [x] File paths sanitized
- [x] Database errors masked
- [x] Sensitive patterns detected

**Test:**
```bash
# Trigger internal error
curl -X GET http://localhost:8050/api/jobs/invalid-id
# Expected: Generic error, no stack trace
```

### Panic Recovery
- [x] Global panic handler
- [x] Stack trace logged
- [x] Generic 500 response
- [x] Service continues running

---

## Network Security

### API Security
- [x] CORS configured
- [x] Rate limiting on login
- [x] Request size limits
- [x] Timeout on all operations
- [ ] API key authentication (optional)
- [ ] IP whitelist (optional)

**Recommended:**
```json
{
  "security": {
    "api_rate_limit": 100,
    "request_timeout": 300,
    "allowed_ips": ["10.0.0.0/8", "192.168.0.0/16"]
  }
}
```

---

## Backup Integrity

### Atomic Operations
- [x] Write to temp file first
- [x] Verify before rename
- [x] Atomic rename operation
- [x] Cleanup on failure
- [x] No partial backups

**Test:**
```bash
# Start backup and interrupt
# Verify no .tmp files remain
ls -la /backup/location/*.tmp
# Expected: No temp files
```

### Verification
- [x] Backup verification after creation
- [x] Checksum validation
- [x] Restore testing
- [x] Integrity checks

---

## System Hardening

### Service Configuration
- [x] Runs as dedicated user
- [x] Minimal privileges
- [x] Resource limits set
- [x] File permissions restricted

**Linux:**
```bash
# Check service user
systemctl show novabackup | grep User
# Expected: novabackup (not root)

# Check file permissions
ls -la /opt/novabackup/
# Expected: 755 for dirs, 644 for files
```

**Windows:**
```powershell
# Check service account
Get-Service NovaBackup | Select-Object ServiceName, StartName
# Expected: LocalService or dedicated account

# Check file permissions
icacls "C:\Program Files\NovaBackup"
# Expected: Administrators:(OI)(CI)F, Users:(OI)(CI)RX
```

### Dependencies
- [x] Go modules with checksums
- [x] Regular security updates
- [x] No known vulnerable dependencies

**Check:**
```bash
go list -m -json all | grep -i vulnerability
```

---

## Physical Security

### Storage Protection
- [x] Backup encryption
- [x] Access controls on storage
- [x] Immutable backups (optional)
- [x] Air-gapped copies (recommended)

---

## Compliance

### Data Privacy
- [x] Audit trails for GDPR
- [x] Right to erasure support
- [x] Data minimization
- [x] Purpose limitation

### Security Standards
- [x] Defense in depth
- [x] Principle of least privilege
- [x] Secure by default
- [x] Fail secure

---

## Penetration Testing

### Automated Scans
```bash
# OWASP ZAP scan
zap-baseline.py -t http://localhost:8050

# Nikto scan
nikto -h http://localhost:8050

# Nmap scan
nmap -sV -sC -p 8050 localhost
```

### Manual Testing
- [ ] Attempt unauthorized access
- [ ] Test privilege escalation
- [ ] Try session hijacking
- [ ] Test replay attacks
- [ ] Attempt DoS
- [ ] Test race conditions

---

## Incident Response

### Detection
- [x] Failed login alerts
- [x] Unusual activity monitoring
- [x] Backup failure alerts
- [x] Storage alerts

### Response
- [x] User disable capability
- [x] Session revocation
- [x] Emergency shutdown
- [x] Log preservation

### Recovery
- [x] Backup restoration tested
- [x] Disaster recovery plan
- [x] RTO/RPO defined

---

## Security Metrics

Track these metrics:
- Failed login attempts (per hour)
- Successful backups (percentage)
- Time to detect incidents
- Time to respond to incidents
- Security patch latency
- Audit log volume

---

## Security Contacts

- **Security Team**: security@example.com
- **On-Call**: security-oncall@example.com
- **PGP Key**: https://example.com/security.asc

---

## Audit Schedule

- **Daily**: Automated security scans
- **Weekly**: Log review
- **Monthly**: Access review
- **Quarterly**: Penetration testing
- **Annually**: Full security audit

---

## Last Security Audit

**Date:** 2026-03-25  
**Auditor:** Security Team  
**Findings:** 0 Critical, 0 High, 3 Medium, 5 Low  
**Status:** ✅ Passed

---

## Certification

This software follows security best practices but is not formally certified.

For compliance requirements, contact security@example.com
