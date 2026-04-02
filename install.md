# Інсталяція агента (Agent Installation)

## Швидка установка

### Windows (PowerShell)
```powershell
iwr -useb https://get.backupsystem.com/agent/install.ps1 | iex
```

### Linux (Bash)
```bash
curl -fsSL https://get.backupsystem.com/agent/install.sh | sudo bash
```

## Параметри інсталяції

```bash
# Повна інсталяція з параметрами
./install.sh \
  --server 10.0.0.1:50051 \
  --token "AAA-BBB-CCC-DDD" \
  --agent-type "hyperv" \
  --auto-start
```

```powershell
# Windows з параметрами
.\install.ps1 -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AgentType "hyperv" -AutoStart
```

## Параметри

| Параметр | Опис | Приклад |
|----------|------|---------|
| `--server` | Адреса Management Server | `10.0.0.1:50051` |
| `--token` | Токен реєстрації агента | `AAA-BBB-CCC-DDD` |
| `--agent-type` | Тип агента | `hyperv`, `vmware`, `kvm`, `mssql`, `postgres`, `oracle` |
| `--install-dir` | Директорія інсталяції | `/opt/backup-agent` |
| `--auto-start` | Автозапуск після інсталяції | `true/false` |
| `--user` | Користувач для запуску | `root` (Linux) |
| `--service-name` | Ім'я служби | `BackupAgent` |
| `--ssl` | Використовувати TLS | `true/false` |

## Типи агентів

| Тип | Опис | Потребує |
|-----|------|----------|
| `hyperv` | Hyper-V агент | Windows Server + Hyper-V |
| `vmware` | VMware агент | VDDK |
| `kvm` | KVM агент | libvirt |
| `mssql` | MS SQL агент | SQL Server instance |
| `postgres` | PostgreSQL агент | PostgreSQL |
| `oracle` | Oracle агент | Oracle Client |

## Демонстрація інсталяції (Silent Install)

```bash
# Linux - повна автоматична інсталяція
curl -fsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- \
  --server $SERVER_ADDR \
  --token $AGENT_TOKEN \
  --agent-type hyperv \
  --auto-start
```

```powershell
# Windows - повна автоматична інсталяція
.\install.ps1 -Server $env:SERVER_ADDR -Token $env:AGENT_TOKEN -AgentType "hyperv" -AutoStart
```

## Перевірка інсталяції

```bash
# Linux
systemctl status backup-agent
# або
./backup-agent --version

# Windows
Get-Service -Name "BackupAgent"
# або
& "C:\Program Files\BackupAgent\backup-agent.exe" --version
```

## Діагностика

```bash
# Linux
./backup-agent diag --collect-logs

# Windows
.\backup-agent.exe diag --collect-logs
```

## Оновлення

```bash
# Linux
./backup-agent update --check

# Windows
.\backup-agent.exe update --check
```

## Видалення

```bash
# Linux
./backup-agent uninstall

# Windows
.\backup-agent.exe uninstall
```

## Docker-інсталяція (альтернатива)

```bash
# Linux
docker run -d \
  --name backup-agent \
  -e SERVER_ADDR=10.0.0.1:50051 \
  -e AGENT_TOKEN=AAA-BBB-CCC-DDD \
  -e AGENT_TYPE=hyperv \
  -v /var/lib/backup-agent:/data \
  -v /mnt/backup-data:/backup-data \
  backupsystem/agent:latest
```
