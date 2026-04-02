# Інсталяція агента (Agent Installation)

## Швидка установка

### Linux (Bash)
```bash
chmod +x install.sh
sudo ./install.sh --server 10.0.0.1:50051 --token "AAA-BBB-CCC-DDD" --agent-type "hyperv" --auto-start
```

### Windows (PowerShell)
```powershell
.\install.ps1 -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AgentType "hyperv" -AutoStart
```

### Універсальна установка (curl/iwr)
```bash
# Linux - з GitHub (рекомендовано)
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --server 10.0.0.1:50051 --token "AAA-BBB-CCC-DDD" --auto-start

# Linux - пропустити SSL (якщо проблеми з сертифікатами)
curl -kfsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- --skip-ssl --server 10.0.0.1:50051 --token "AAA-BBB-CCC-DDD" --auto-start

# Linux - зберегти скрипт і виконати
curl -fsSL -o install.sh https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh
sudo chmod +x install.sh && sudo ./install.sh --server 10.0.0.1:50051 --token "AAA-BBB-CCC-DDD" --auto-start
```

```powershell
# Windows - з GitHub (рекомендовано)
irm https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.ps1 | iex -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AutoStart

# Windows - пропустити SSL
iwr -useb -SkipCertificateCheck https://get.backupsystem.com/agent/install.ps1 | iex -SkipSSL -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AutoStart

# Windows - зберегти скрипт і виконати
irm -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.ps1" -OutFile install.ps1
.\install.ps1 -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AutoStart
```

## Серверна інсталяція (одною командою)

### Linux Server
```bash
sudo ./install.sh --mode server --install-dir /opt/backup-server
```

### Windows Server
```powershell
.\install.ps1 -Mode server -InstallDir "C:\BackupServer"
```

### Повна інсталяція (Server + Agent)
```bash
# Linux
sudo ./install.sh --mode all --server 10.0.0.1:50051 --token "AAA-BBB-CCC-DDD" --agent-type "hyperv" --auto-start
```

```powershell
# Windows
.\install.ps1 -Mode all -Server "10.0.0.1:50051" -Token "AAA-BBB-CCC-DDD" -AgentType "hyperv" -AutoStart
```

## Параметри інсталяції

```bash
# Повна інсталяція агента з параметрами
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
| `--mode` | Режим інсталяції | `agent`, `server`, `all` |
| `--auto-start` | Автозапуск після інсталяції | `true/false` |
| `--user` | Користувач для запуску | `root` (Linux) |
| `--service-name` | Ім'я служби | `BackupAgent` |
| `--skip-ssl` | Пропустити перевірку SSL | - |
| `--source-url` | Альтернативний URL скрипта | `http://server/install.sh` |
| `--local-source` | Локальний шлях до сирців | `/path/to/source` |

## Типи агентів

| Тип | Опис | Потребує |
|-----|------|----------|
| `hyperv` | Hyper-V агент | Windows Server + Hyper-V |
| `vmware` | VMware агент | VDDK |
| `kvm` | KVM агент | libvirt |
| `mssql` | MS SQL агент | SQL Server instance |
| `postgres` | PostgreSQL агент | PostgreSQL |
| `oracle` | Oracle агент | Oracle Client |

## Docker-інсталяція (альтернатива)

### Agent в Docker
```bash
docker run -d \
  --name backup-agent \
  -e SERVER_ADDR=10.0.0.1:50051 \
  -e AGENT_TOKEN=AAA-BBB-CCC-DDD \
  -e AGENT_TYPE=hyperv \
  -v /var/lib/backup-agent:/data \
  -v /mnt/backup-data:/backup-data \
  backupsystem/agent:latest
```

### Server в Docker
```bash
docker run -d \
  --name backup-server \
  -p 50051:50051 \
  -p 8080:80 \
  -v /backup/data:/data \
  -e DB_CONNECTION="Server=db;Database=backup;User=sa;Password=P@ssw0rd" \
  backupsystem/server:latest
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
./install.sh --uninstall

# Windows
.\install.ps1 -Uninstall
```
