# 📊 NovaBackup Monitoring & Alerting Setup Guide

## Document Information

| Attribute | Value |
|-----------|-------|
| **Version** | 1.0 |
| **Last Updated** | March 27, 2026 |
| **Owner** | Operations Team |
| **Review Cycle** | Quarterly |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prometheus Metrics](#2-prometheus-metrics)
3. [Grafana Dashboard Setup](#3-grafana-dashboard-setup)
4. [Alert Rules Configuration](#4-alert-rules-configuration)
5. [Notification Channels](#5-notification-channels)
6. [Log Aggregation](#6-log-aggregation)
7. [Health Checks](#7-health-checks)
8. [Custom Monitoring](#8-custom-monitoring)

---

## 1. Overview

### 1.1 Purpose

This guide describes how to set up monitoring, alerting, and observability for NovaBackup. It covers Prometheus metrics, Grafana dashboards, alert rules, and notification integrations.

### 1.2 Monitoring Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   NovaBackup    │────▶│  Prometheus  │────▶│   Grafana   │
│   Application   │     │   Server     │     │  Dashboard  │
└─────────────────┘     └──────────────┘     └─────────────┘
                              │
                              ▼
                        ┌──────────────┐
                        │  Alertmanager │
                        └──────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐     ┌──────────┐    ┌──────────┐
         │ Email  │     │ Telegram │    │ Webhook  │
         └────────┘     └──────────┘    └──────────┘
```

### 1.3 Key Metrics Categories

| Category | Metrics |
|----------|---------|
| **Backup** | Success rate, duration, size, count |
| **Restore** | Success rate, duration, count |
| **Storage** | Usage, available space, growth rate |
| **API** | Request rate, latency, error rate |
| **System** | CPU, memory, disk I/O |
| **Security** | Failed logins, RBAC violations |

---

## 2. Prometheus Metrics

### 2.1 Enabling Metrics

Metrics are enabled by default. Verify in `.env`:

```ini
NOVABACKUP_METRICS_ENABLED=true
NOVABACKUP_METRICS_PATH=/metrics
```

### 2.2 Accessing Metrics

```bash
# Access metrics endpoint
curl http://localhost:8050/metrics

# Example output
# HELP novabackup_backups_total Total number of backups
# TYPE novabackup_backups_total counter
novabackup_backups_total{status="success"} 150
novabackup_backups_total{status="failed"} 5

# HELP novabackup_backup_duration_seconds Backup duration
# TYPE novabackup_backup_duration_seconds histogram
novabackup_backup_duration_seconds_bucket{le="60.0"} 120
novabackup_backup_duration_seconds_bucket{le="300.0"} 145
novabackup_backup_duration_seconds_bucket{le="+Inf"} 155
```

### 2.3 Available Metrics

#### Backup Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_backups_total` | Counter | Total backups created |
| `novabackup_backups_success_total` | Counter | Successful backups |
| `novabackup_backups_failed_total` | Counter | Failed backups |
| `novabackup_backup_duration_seconds` | Histogram | Backup duration |
| `novabackup_backup_size_bytes` | Gauge | Last backup size |
| `novabackup_backup_pending` | Gauge | Pending backups |

#### Restore Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_restores_total` | Counter | Total restores |
| `novabackup_restores_success_total` | Counter | Successful restores |
| `novabackup_restores_failed_total` | Counter | Failed restores |
| `novabackup_restore_duration_seconds` | Histogram | Restore duration |

#### Storage Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_storage_usage_bytes` | Gauge | Current storage usage |
| `novabackup_storage_available_bytes` | Gauge | Available storage |
| `novabackup_storage_total_bytes` | Gauge | Total storage capacity |
| `novabackup_storage_growth_bytes` | Counter | Storage growth rate |

#### API Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_api_requests_total` | Counter | Total API requests |
| `novabackup_api_requests_in_progress` | Gauge | In-progress requests |
| `novabackup_api_request_duration_seconds` | Histogram | Request latency |
| `novabackup_api_errors_total` | Counter | API errors |

#### Security Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_auth_attempts_total` | Counter | Authentication attempts |
| `novabackup_auth_failures_total` | Counter | Failed authentications |
| `novabackup_active_sessions` | Gauge | Active user sessions |
| `novabackup_rbac_violations_total` | Counter | RBAC violations |

#### Scheduler Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `novabackup_scheduler_jobs_total` | Gauge | Total scheduled jobs |
| `novabackup_scheduler_jobs_enabled` | Gauge | Enabled jobs |
| `novabackup_scheduler_executions_total` | Counter | Job executions |
| `novabackup_scheduler_execution_errors_total` | Counter | Job execution errors |

### 2.4 Prometheus Configuration

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'novabackup'
    static_configs:
      - targets: ['localhost:8050']
    metrics_path: '/metrics'
    scrape_interval: 30s
    
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

rule_files:
  - 'alert_rules.yml'
```

Start Prometheus:

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v prometheus_data:/prometheus \
  prom/prometheus:latest
```

---

## 3. Grafana Dashboard Setup

### 3.1 Installing Grafana

#### Docker Installation

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -v grafana_data:/var/lib/grafana \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana:latest
```

#### Linux Installation (Ubuntu)

```bash
# Add Grafana repository
sudo apt-get install -y apt-transport-https
sudo apt-get install -y software-properties-common wget
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list

# Install
sudo apt-get update
sudo apt-get install grafana

# Start
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

### 3.2 Configuring Data Source

1. **Access Grafana**: `http://localhost:3000`
2. **Login**: admin/admin (change password!)
3. **Add Data Source**:
   - Go to Configuration → Data Sources
   - Click "Add data source"
   - Select "Prometheus"
   - URL: `http://localhost:9090`
   - Click "Save & Test"

### 3.3 Import Dashboard

#### Option A: Import JSON Dashboard

Save `novabackup-dashboard.json`:

```json
{
  "dashboard": {
    "title": "NovaBackup Overview",
    "panels": [
      {
        "id": 1,
        "title": "Backup Success Rate",
        "type": "gauge",
        "targets": [
          {
            "expr": "rate(novabackup_backups_success_total[1h]) / rate(novabackup_backups_total[1h]) * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100,
            "thresholds": {
              "steps": [
                {"value": null, "color": "red"},
                {"value": 80, "color": "yellow"},
                {"value": 95, "color": "green"}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Total Backups",
        "type": "stat",
        "targets": [
          {
            "expr": "novabackup_backups_total"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Storage Usage",
        "type": "gauge",
        "targets": [
          {
            "expr": "novabackup_storage_usage_bytes / novabackup_storage_total_bytes * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "steps": [
                {"value": null, "color": "green"},
                {"value": 70, "color": "yellow"},
                {"value": 85, "color": "red"}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 12, "y": 0}
      },
      {
        "id": 4,
        "title": "Active Sessions",
        "type": "stat",
        "targets": [
          {
            "expr": "novabackup_active_sessions"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 5,
        "title": "Backup Duration (p95)",
        "type": "stat",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(novabackup_backup_duration_seconds_bucket[1h]))"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "s"
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 8}
      },
      {
        "id": 6,
        "title": "API Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(novabackup_api_requests_total[5m])"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 6, "y": 8}
      },
      {
        "id": 7,
        "title": "Failed Logins (24h)",
        "type": "stat",
        "targets": [
          {
            "expr": "increase(novabackup_auth_failures_total[24h])"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "thresholds": {
              "steps": [
                {"value": null, "color": "green"},
                {"value": 10, "color": "yellow"},
                {"value": 50, "color": "red"}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 8}
      },
      {
        "id": 8,
        "title": "Backup Trend (7 days)",
        "type": "graph",
        "targets": [
          {
            "expr": "increase(novabackup_backups_total[1d])"
          }
        ],
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 16}
      }
    ],
    "time": {"from": "now-6h", "to": "now"},
    "refresh": "30s"
  }
}
```

Import via Grafana UI:
1. Dashboard → Import
2. Upload JSON file
3. Select Prometheus data source
4. Click Import

#### Option B: Use Grafana CLI

```bash
# Import dashboard
grafana-cli --adminUrl http://localhost:3000 \
  --adminUser admin \
  --adminPassword admin \
  dashboards import-from-file novabackup-dashboard.json
```

---

## 4. Alert Rules Configuration

### 4.1 Alert Rules File

Create `alert_rules.yml`:

```yaml
groups:
  - name: novabackup_alerts
    interval: 30s
    rules:
      # Backup Failure Rate
      - alert: HighBackupFailureRate
        expr: |
          rate(novabackup_backups_failed_total[1h]) 
          / rate(novabackup_backups_total[1h]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High backup failure rate"
          description: "Backup failure rate is {{ $value | humanizePercentage }} over the last hour"
          
      # No Recent Backups
      - alert: NoRecentBackups
        expr: |
          time() - novabackup_backup_duration_seconds_timestamp > 86400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "No backups in last 24 hours"
          description: "No successful backups have been created in the last 24 hours"
          
      # Storage High
      - alert: StorageUsageHigh
        expr: |
          novabackup_storage_usage_bytes / novabackup_storage_total_bytes > 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Storage usage above 80%"
          description: "Storage usage is {{ $value | humanizePercentage }}"
          
      # Storage Critical
      - alert: StorageUsageCritical
        expr: |
          novabackup_storage_usage_bytes / novabackup_storage_total_bytes > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Storage usage above 90%"
          description: "Storage usage is {{ $value | humanizePercentage }} - immediate action required"
          
      # Long Running Backup
      - alert: LongRunningBackup
        expr: |
          histogram_quantile(0.99, rate(novabackup_backup_duration_seconds_bucket[1h])) > 3600
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Backups taking longer than 1 hour"
          description: "99th percentile backup duration is {{ $value | humanizeDuration }}"
          
      # API Error Rate
      - alert: HighAPIErrorRate
        expr: |
          rate(novabackup_api_errors_total[5m]) 
          / rate(novabackup_api_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API error rate"
          description: "API error rate is {{ $value | humanizePercentage }}"
          
      # Service Down
      - alert: NovaBackupDown
        expr: |
          up{job="novabackup"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "NovaBackup service is down"
          description: "NovaBackup instance {{ $labels.instance }} is not responding"
          
      # Multiple Failed Logins
      - alert: MultipleFailedLogins
        expr: |
          increase(novabackup_auth_failures_total[15m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Multiple failed login attempts"
          description: "{{ $value }} failed login attempts in the last 15 minutes"
          
      # RBAC Violations
      - alert: RBACViolations
        expr: |
          increase(novabackup_rbac_violations_total[1h]) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "RBAC violations detected"
          description: "{{ $value }} RBAC violations in the last hour"
          
      # Scheduler Job Failures
      - alert: SchedulerJobFailures
        expr: |
          increase(novabackup_scheduler_execution_errors_total[1h]) > 3
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Scheduled backup job failures"
          description: "{{ $value }} scheduled job failures in the last hour"
```

### 4.2 Alertmanager Configuration

Create `alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alertmanager@yourdomain.com'
  smtp_auth_username: 'alerts@yourdomain.com'
  smtp_auth_password: 'YOUR_PASSWORD'
  
  # Slack
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
  
  # Telegram
  telegram_api_url: 'https://api.telegram.org'

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default'
  
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'default'
    email_configs:
      - to: 'ops-team@yourdomain.com'
        send_resolved: true
        
  - name: 'critical-alerts'
    email_configs:
      - to: 'ops-team@yourdomain.com'
        send_resolved: true
    slack_configs:
      - channel: '#alerts-critical'
        send_resolved: true
    telegram_configs:
      - bot_token: 'YOUR_BOT_TOKEN'
        chat_id: -1001234567890
        send_resolved: true
        
  - name: 'warning-alerts'
    email_configs:
      - to: 'ops-team@yourdomain.com'
        send_resolved: true
    slack_configs:
      - channel: '#alerts-warning'
        send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname']
```

Start Alertmanager:

```bash
docker run -d \
  --name alertmanager \
  -p 9093:9093 \
  -v $(pwd)/alertmanager.yml:/etc/alertmanager/alertmanager.yml \
  prom/alertmanager:latest
```

---

## 5. Notification Channels

### 5.1 Email Notifications

Configure in `.env`:

```ini
NOVABACKUP_SMTP_HOST=smtp.gmail.com
NOVABACKUP_SMTP_PORT=587
NOVABACKUP_SMTP_USE_TLS=true
NOVABACKUP_SMTP_USER=alerts@yourdomain.com
NOVABACKUP_SMTP_PASSWORD=YOUR_APP_PASSWORD
NOVABACKUP_SMTP_FROM=novabackup@yourdomain.com
NOVABACKUP_SMTP_TO=ops-team@yourdomain.com,admin@yourdomain.com
```

Test email notification:

```bash
curl -X POST http://localhost:8050/notifications/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"channel": "email", "message": "Test notification from NovaBackup"}'
```

### 5.2 Telegram Notifications

Configure in `.env`:

```ini
NOVABACKUP_TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
NOVABACKUP_TELEGRAM_CHAT_IDS=-1001234567890,-1009876543210
```

**Setup Steps:**

1. Create bot via @BotFather
2. Get bot token
3. Add bot to group/channel
4. Get chat ID: `curl https://api.telegram.org/botTOKEN/getUpdates`
5. Add chat ID to `.env`

Test Telegram notification:

```bash
curl -X POST http://localhost:8050/notifications/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"channel": "telegram", "message": "Test notification from NovaBackup"}'
```

### 5.3 Slack Notifications (Webhook)

Configure in `.env`:

```ini
NOVABACKUP_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
NOVABACKUP_WEBHOOK_AUTH_TOKEN=optional-token
```

**Setup Steps:**

1. Go to Slack Workspace Settings
2. Add → Incoming Webhooks
3. Create webhook URL
4. Copy URL to `.env`

### 5.4 Microsoft Teams (Webhook)

```ini
NOVABACKUP_WEBHOOK_URL=https://outlook.office.com/webhook/YOUR-WEBHOOK-URL
```

---

## 6. Log Aggregation

### 6.1 Local Log Configuration

Configure in `.env`:

```ini
NOVABACKUP_LOG_LEVEL=INFO
NOVABACKUP_LOG_FORMAT=json
NOVABACKUP_AUDIT_ENABLED=true
NOVABACKUP_AUDIT_LOG_FILE=/var/log/novabackup/audit.log
```

### 6.2 Log Rotation

Create `/etc/logrotate.d/novabackup`:

```
/var/log/novabackup/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 novabackup novabackup
    postrotate
        systemctl reload novabackup > /dev/null 2>&1 || true
    endscript
}
```

### 6.3 ELK Stack Integration

#### Docker Compose for ELK

```yaml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"
      
  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
      - /var/log/novabackup:/var/log/novabackup
    ports:
      - "5044:5044"
      
  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
    ports:
      - "5601:5600"
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

#### Logstash Configuration

Create `logstash.conf`:

```
input {
  file {
    path => "/var/log/novabackup/*.log"
    start_position => "beginning"
    type => "novabackup"
  }
}

filter {
  json {
    source => "message"
  }
  
  date {
    match => ["timestamp", "ISO8601"]
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "novabackup-logs-%{+YYYY.MM.dd}"
  }
}
```

### 6.4 Loki Integration

#### Docker Compose for Loki

```yaml
version: '3.8'
services:
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    
  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - /var/log/novabackup:/var/log/novabackup
      - ./promtail-config.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
```

Create `promtail-config.yml`:

```yaml
server:
  http_listen_port: 9080
  
positions:
  filename: /tmp/positions.yaml
  
clients:
  - url: http://loki:3100/loki/api/v1/push
  
scrape_configs:
  - job_name: novabackup
    static_configs:
      - targets:
          - localhost
        labels:
          job: novabackup
          __path__: /var/log/novabackup/*.log
```

---

## 7. Health Checks

### 7.1 Application Health Endpoint

```bash
# Basic health check
curl http://localhost:8050/health

# Expected response
{
  "status": "healthy",
  "timestamp": "2026-03-27T10:00:00Z",
  "version": "8.5.0",
  "checks": {
    "database": "healthy",
    "storage": "healthy",
    "cloud_providers": "healthy"
  }
}
```

### 7.2 Docker Health Check

Already configured in `docker-compose.yml`:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8000/docs"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 7.3 Uptime Monitoring

#### Uptime Kuma Setup

```bash
docker run -d \
  --name uptime-kuma \
  -p 3001:3001 \
  -v uptime-kuma:/app/data \
  louislam/uptime-kuma:1

# Access: http://localhost:3001
# Add monitor: HTTP(s) → http://localhost:8050/health
# Set interval: 60 seconds
# Configure notifications
```

### 7.4 Synthetic Monitoring

Create health check script:

```bash
#!/bin/bash
# health-check.sh

BASE_URL="http://localhost:8050"

# Check API health
HEALTH=$(curl -s $BASE_URL/health | jq -r '.status')
if [ "$HEALTH" != "healthy" ]; then
    echo "CRITICAL: API health check failed"
    exit 1
fi

# Check metrics endpoint
METRICS=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/metrics)
if [ "$METRICS" != "200" ]; then
    echo "WARNING: Metrics endpoint returned $METRICS"
fi

# Check dashboard
DASHBOARD=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/static/index.html)
if [ "$DASHBOARD" != "200" ]; then
    echo "WARNING: Dashboard returned $DASHBOARD"
fi

echo "OK: All health checks passed"
exit 0
```

---

## 8. Custom Monitoring

### 8.1 Custom Metrics Export

```python
# Add custom metrics in your application
from prometheus_client import Gauge, Counter

# Define custom metrics
active_backups = Gauge('novabackup_active_backups', 'Number of active backup jobs')
total_storage_used = Gauge('novabackup_total_storage_used', 'Total storage used in bytes')

# Update metrics
active_backups.set(get_active_backup_count())
total_storage_used.set(get_total_storage_used())
```

### 8.2 Business Metrics

Track business-relevant metrics:

```ini
# Backup coverage
novabackup_vms_protected_total

# Compliance
novabackup_retention_compliance_percent

# Cost tracking
novabackup_cloud_storage_cost_dollars

# SLA tracking
novabackup_sla_compliance_percent
```

### 8.3 Custom Dashboards

Create custom Grafana panels for:

- Backup coverage by VM
- Storage growth trends
- Cost analysis
- SLA compliance
- Recovery time metrics

---

## Appendix A: Quick Start Monitoring Stack

```bash
# Start complete monitoring stack
docker-compose -f monitoring-compose.yml up -d

# monitoring-compose.yml includes:
# - Prometheus
# - Grafana
# - Alertmanager
# - Uptime Kuma
```

---

## Appendix B: Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| No metrics | Endpoint not scraped | Check Prometheus targets |
| Alerts not firing | Rules not loaded | Check rule_files in prometheus.yml |
| Notifications not sent | Receiver misconfigured | Test notification channel |
| Dashboard empty | Data source wrong | Select correct Prometheus DS |

### Diagnostic Commands

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check alert rules
curl http://localhost:9090/api/v1/rules

# Check active alerts
curl http://localhost:9090/api/v1/alerts

# Check Alertmanager status
curl http://localhost:9093/api/v1/status
```

---

**END OF DOCUMENT**
