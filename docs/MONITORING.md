# NovaBackup Monitoring Configuration

## Prometheus Metrics Export

### Setup Prometheus Scraper

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'novabackup'
    static_configs:
      - targets: ['localhost:8050']
    metrics_path: '/api/metrics'
    scrape_interval: 30s
```

### Available Metrics

```
# HELP novabackup_jobs_total Total number of backup jobs
# TYPE novabackup_jobs_total gauge
novabackup_jobs_total 15

# HELP novabackup_jobs_enabled Number of enabled backup jobs
# TYPE novabackup_jobs_enabled gauge
novabackup_jobs_enabled 12

# HELP novabackup_backup_running Number of currently running backups
# TYPE novabackup_backup_running gauge
novabackup_backup_running 2

# HELP novabackup_backup_total Total number of backup sessions
# TYPE novabackup_backup_total counter
novabackup_backup_total 1543

# HELP novabackup_backup_success_total Total number of successful backups
# TYPE novabackup_backup_success_total counter
novabackup_backup_success_total 1498

# HELP novabackup_backup_failed_total Total number of failed backups
# TYPE novabackup_backup_failed_total counter
novabackup_backup_failed_total 45

# HELP novabackup_backup_duration_seconds Backup duration in seconds
# TYPE novabackup_backup_duration_seconds histogram
novabackup_backup_duration_seconds_bucket{le="60"} 234
novabackup_backup_duration_seconds_bucket{le="300"} 876
novabackup_backup_duration_seconds_bucket{le="1800"} 1234
novabackup_backup_duration_seconds_bucket{le="3600"} 1456
novabackup_backup_duration_seconds_bucket{le="+Inf"} 1543
novabackup_backup_duration_seconds_sum 456789.12
novabackup_backup_duration_seconds_count 1543

# HELP novabackup_backup_size_bytes Total size of backups in bytes
# TYPE novabackup_backup_size_bytes counter
novabackup_backup_size_bytes 5678901234567

# HELP novabackup_storage_used_bytes Storage used in bytes
# TYPE novabackup_storage_used_bytes gauge
novabackup_storage_used_bytes 3456789012345

# HELP novabackup_storage_total_bytes Total storage capacity in bytes
# TYPE novabackup_storage_total_bytes gauge
novabackup_storage_total_bytes 10995116277760

# HELP novabackup_users_total Total number of users
# TYPE novabackup_users_total gauge
novabackup_users_total 25

# HELP novabackup_audit_logs_total Total number of audit log entries
# TYPE novabackup_audit_logs_total gauge
novabackup_audit_logs_total 45678

# HELP novabackup_http_requests_total Total HTTP requests
# TYPE novabackup_http_requests_total counter
novabackup_http_requests_total{method="GET",endpoint="/api/jobs",status="200"} 1234
novabackup_http_requests_total{method="POST",endpoint="/api/jobs",status="201"} 567

# HELP novabackup_http_request_duration_seconds HTTP request latency
# TYPE novabackup_http_request_duration_seconds histogram
novabackup_http_request_duration_seconds_bucket{le="0.1"} 2345
novabackup_http_request_duration_seconds_bucket{le="0.5"} 3456
novabackup_http_request_duration_seconds_bucket{le="1"} 4567
novabackup_http_request_duration_seconds_bucket{le="5"} 5678
novabackup_http_request_duration_seconds_bucket{le="+Inf"} 6789
```

---

## Grafana Dashboard

### Import Dashboard

Import dashboard from `grafana/novabackup-dashboard.json`:

```json
{
  "dashboard": {
    "title": "NovaBackup Enterprise",
    "panels": [
      {
        "title": "Backup Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(novabackup_backup_success_total[1h]) / rate(novabackup_backup_total[1h]) * 100"
          }
        ],
        "thresholds": [
          {"value": 95, "color": "green"},
          {"value": 90, "color": "yellow"},
          {"value": 0, "color": "red"}
        ]
      },
      {
        "title": "Running Backups",
        "type": "gauge",
        "targets": [
          {
            "expr": "novabackup_backup_running"
          }
        ]
      },
      {
        "title": "Storage Usage",
        "type": "timeseries",
        "targets": [
          {
            "expr": "novabackup_storage_used_bytes / novabackup_storage_total_bytes * 100"
          }
        ]
      },
      {
        "title": "Backup Duration (p95)",
        "type": "timeseries",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(novabackup_backup_duration_seconds_bucket[1h]))"
          }
        ]
      },
      {
        "title": "Failed Backups (24h)",
        "type": "stat",
        "targets": [
          {
            "expr": "increase(novabackup_backup_failed_total[24h])"
          }
        ],
        "thresholds": [
          {"value": 0, "color": "green"},
          {"value": 1, "color": "yellow"},
          {"value": 5, "color": "red"}
        ]
      },
      {
        "title": "API Request Rate",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(novabackup_http_requests_total[5m])"
          }
        ]
      }
    ]
  }
}
```

---

## Alert Rules

### Prometheus Alert Rules

Add to `alerts.yml`:

```yaml
groups:
  - name: novabackup
    rules:
      # Backup Failure Alert
      - alert: NovaBackupHighFailureRate
        expr: rate(novabackup_backup_failed_total[1h]) / rate(novabackup_backup_total[1h]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High backup failure rate"
          description: "Backup failure rate is {{ $value | humanizePercentage }} over the last hour"

      # No Backups in 24h
      - alert: NovaBackupNoRecentBackups
        expr: time() - novabackup_backup_duration_seconds_timestamp > 86400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "No successful backups in 24 hours"
          description: "Last successful backup was more than 24 hours ago"

      # Storage Usage Alert
      - alert: NovaBackupStorageHigh
        expr: novabackup_storage_used_bytes / novabackup_storage_total_bytes * 100 > 85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Backup storage usage high"
          description: "Storage usage is {{ $value | humanizePercentage }}"

      # Storage Critical
      - alert: NovaBackupStorageCritical
        expr: novabackup_storage_used_bytes / novabackup_storage_total_bytes * 100 > 95
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Backup storage almost full"
          description: "Storage usage is {{ $value | humanizePercentage }}"

      # Long Running Backup
      - alert: NovaBackupLongRunning
        expr: novabackup_backup_running > 0 and histogram_quantile(0.99, rate(novabackup_backup_duration_seconds_bucket[1h])) > 14400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Backup running longer than expected"
          description: "Backup has been running for more than 4 hours"

      # API Error Rate
      - alert: NovaBackupAPIErrors
        expr: rate(novabackup_http_requests_total{status=~"5.."}[5m]) / rate(novabackup_http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API error rate"
          description: "API error rate is {{ $value | humanizePercentage }}"

      # Service Down
      - alert: NovaBackupServiceDown
        expr: up{job="novabackup"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "NovaBackup service is down"
          description: "NovaBackup instance {{ $labels.instance }} is not responding"

      # Audit Log Security Alert
      - alert: NovaBackupMultipleFailedLogins
        expr: increase(novabackup_audit_logs_total{action="login",success="false"}[15m]) > 10
        for: 0m
        labels:
          severity: warning
        annotations:
          summary: "Multiple failed login attempts"
          description: "More than 10 failed login attempts in 15 minutes"

      # User Disabled
      - alert: NovaBackupUserDisabled
        expr: changes(novabackup_users_enabled[1h]) > 0
        for: 0m
        labels:
          severity: info
        annotations:
          summary: "User account status changed"
          description: "A user account was enabled or disabled"
```

---

## Health Check Endpoint

### Simple Health Check

```bash
curl http://localhost:8050/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "7.0.0",
  "time": "2026-03-25T10:30:00Z"
}
```

### Detailed Health Check

```bash
curl http://localhost:8050/api/health/detailed
```

**Response:**
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5
    },
    "storage": {
      "status": "healthy",
      "free_space_gb": 2500
    },
    "backup_engine": {
      "status": "healthy",
      "running_jobs": 2
    },
    "scheduler": {
      "status": "healthy",
      "next_job_in_seconds": 1200
    }
  }
}
```

---

## Log Aggregation

### ELK Stack Configuration

**Logstash Config:**
```ruby
input {
  file {
    path => "C:/ProgramData/NovaBackup/Logs/*.log"
    start_position => "beginning"
    sincedb_path => "C:/ProgramData/Logstash/sincedb"
  }
}

filter {
  grok {
    match => { "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:log_message}" }
  }
  date {
    match => [ "timestamp", "ISO8601" ]
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "novabackup-logs-%{+YYYY.MM.dd}"
  }
}
```

### Splunk Forwarder

**inputs.conf:**
```ini
[monitor://C:\ProgramData\NovaBackup\Logs]
disabled = false
index = novabackup
sourcetype = novabackup:logs
```

---

## Notification Integration

### Slack Alerts

```yaml
# In alertmanager.yml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#backup-alerts'
        title: 'NovaBackup Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        send_resolved: true
```

### Email Alerts

```yaml
receivers:
  - name: 'email'
    email_configs:
      - to: 'backup-team@example.com'
        from: 'novabackup@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'novabackup'
        auth_password: 'password'
        headers:
          Subject: 'NovaBackup Alert: {{ .GroupLabels.alertname }}'
```

### Microsoft Teams

```yaml
receivers:
  - name: 'teams'
    webhook_configs:
      - url: 'https://outlook.office.com/webhook/YOUR/WEBHOOK/URL'
```

---

## Systemd Service Status (Linux)

```bash
# Check service status
systemctl status novabackup

# View logs
journalctl -u novabackup -f

# Check resource usage
systemd-cgtop | grep novabackup
```

---

## Windows Performance Monitor

### Add Counters

1. Open Performance Monitor (`perfmon.msc`)
2. Add counters for NovaBackup:
   - Process → % Processor Time → novabackup
   - Process → Private Bytes → novabackup
   - Process → IO Read Bytes/sec → novabackup
   - Process → IO Write Bytes/sec → novabackup

### PowerShell Monitoring Script

```powershell
# monitor-novabackup.ps1
while ($true) {
    $response = Invoke-WebRequest -Uri "http://localhost:8050/api/health" -UseBasicParsing
    $health = $response.Content | ConvertFrom-Json
    
    if ($health.status -ne "healthy") {
        Write-EventLog -LogName Application -Source "NovaBackup" -EntryType Error `
            -EventId 1001 -Message "NovaBackup health check failed"
    }
    
    Start-Sleep -Seconds 60
}
```

---

## Resource Limits

### Linux (systemd)

Edit `/etc/systemd/system/novabackup.service`:

```ini
[Service]
LimitNOFILE=65535
LimitNPROC=4096
MemoryLimit=8G
CPUQuota=200%
```

### Windows (Job Objects)

Use Windows Job Objects to limit:
- Memory: 8 GB
- CPU: 200%
- IO: 500 MB/s

---

## Monitoring Best Practices

1. **Set up redundant monitoring** (Prometheus + CloudWatch/Azure Monitor)
2. **Configure alert escalation** (Warning → Critical → Page)
3. **Monitor backup success rate** (target: >95%)
4. **Track storage trends** (alert at 85%, critical at 95%)
5. **Monitor API latency** (p95 < 500ms)
6. **Watch for failed login attempts** (security)
7. **Track backup duration trends** (detect performance degradation)
8. **Monitor disk I/O** (detect bottlenecks)
9. **Set up synthetic transactions** (test backup/restore weekly)
10. **Review audit logs regularly** (security compliance)

---

## Support Contacts

- **On-Call**: backup-oncall@example.com
- **Slack**: #backup-support
- **PagerDuty**: novabackup-critical
