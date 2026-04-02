# Технічні вимоги до Backup-системи

## 1. Підтримка гіпервізорів

| Гіпервізор | Бекап | Відновлення | Notes |
|------------|-------|-------------|-------|
| Hyper-V | ✅ | ✅ | Пріоритет 1 |
| VMware | ✅ | ✅ | VDDK |
| KVM | ✅ | ✅ | libvirt |

## 2. Підтримка баз даних

| БД | Бекап | Відновлення | Метод |
|----|-------|-------------|-------|
| Microsoft SQL Server | ✅ | ✅ | VDI/VSS |
| PostgreSQL | ✅ | ✅ | pg_dump / native |
| Oracle Database | ✅ | ✅ | RMAN / expdp |

### 2.1. Особливості бекапу БД
- **VSS Writer** для MS SQL (кронштейн consistency)
- **Point-in-time recovery** для всіх БД
- **Transaction log backup** для MS SQL
- **WAL archiving** для PostgreSQL

## 3. Хмарна інтеграція

### 3.1. Azure
- Azure Blob Storage (Hot/Cool/Archive tiers)
- Azure Files (SMB 3.0)
- Azure VM Agent для бекапу
- Azure Archive Storage для довгострокового зберігання

### 3.2. AWS
- Amazon S3 (Standard/IA/Glacier)
- Amazon EBS snapshots
- AWS Storage Gateway

### 3.3. Google Cloud
- Google Cloud Storage
- Compute Engine snapshots

## 4. Методи бекапу

| Метод | Опис | Підтримка |
|-------|------|-----------|
| Full | Повний бекап всіх даних | ✅ |
| Incremental (CBT) | Тільки змінені блоки | ✅ |
| Differential | Зміни з моменту останнього Full | ✅ |
| Synthetic Full | Інкременти → повний без VM | ✅ |
| Continuous Data Protection (CDP) | Real-time захист | Future |

### 4.1. Дедuплікація
- **Source-side dedup**: стиснення до передачі по мережі
- **Target-side dedup**: зберігання на repository
- **Algorithm**: Zstd (швидкість) / SHA-256 (hash)

### 4.2. Компресія
- Zstd (рекомендовано)
- LZ4 (high speed)
- Gzip (compat)

## 5. Безпека

### 5.1. Шифрування
- **AES-256** для даних at rest
- **TLS 1.3** для передачі по мережі
- **Per-VM encryption keys** (KMIP support)

### 5.2. Авторизація
- Role-Based Access Control (RBAC)
- LDAP/Active Directory integration
- Two-Factor Authentication (2FA)
- Audit logging всіх операцій

### 5.3. Network Security
- VPN tunnel для offsite реплікації
- Bandwidth throttling
- Network encryption

## 6. Швидкість та продуктивність

### 6.1. Цілі
- **Speed**: > 1 GB/s на диск
- **Latency**: < 100ms для метаданих
- **Concurrency**: 100+ паралельних завдань

### 6.2. Оптимізації
- RDMA підтримка для high-speed мереж
- Multithreading (C++)
- RAM caching для метаданих
- Fast Clone (ReFS/XFS)

## 7. Планування (Scheduler)

### 7.1. Типи розкладів
- **Cron-like** синтаксис
- **GFS** (Grandfather-Father-Son)
- **3-2-1 Rule**: 3 копії, 2 носія, 1 offsite
- **Wimam** window (час доби)

### 7.2. Retention Policies
- Daily: 7 копій
- Weekly: 4 копії
- Monthly: 12 копій
- Yearly: 7 копій

### 7.3. Chain rotation
- GFS
- Tower of Hanoi
- Custom

## 8. Веб-інтерфейс (UI/UX)

### 8.1. Технології
- **Frontend**: React + TypeScript / Blazor
- **Design System**: Material UI / Fluent UI
- **Charts**: Recharts / Chart.js
- **State**: Redux Toolkit / Zustand

### 8.2. Сторінки
1. **Dashboard**: Загальний стан, метрики
2. **Jobs**: Список завдань, статус
3. **Backup Store**: Репозиторії, використання
4. **VMs/DBs**: Список захищених об'єктів
5. **Restore**: Відновлення
6. **Settings**: Налаштування
7. **Reports**: Звіти
8. **Users**: Управління користувачами

### 8.3. Mobile-first responsive
- Desktop: 1920px+
- Tablet: 768px - 1919px
- Mobile: < 768px

## 9. API

### 9.1. RESTful
- OpenAPI 3.0 specification
- JWT authentication
- Rate limiting

### 9.2. gRPC (internally)
- C# ↔ C++ communication
- Streaming для великих даних

## 10. Структура компонентів

```
┌─────────────────────────────────────────────────────────────┐
│                    Web UI (React/Blazor)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Management Server (ASP.NET Core)                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │  Scheduler  │ │   REST API  │ │  Repository Manager    │ │
│  │ (Quartz.NET)│ │   (Swagger) │ │  (Retention Policy)    │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
    ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
    │  Hyper-V    │   │    VMware   │   │    KVM      │
    │   Agent     │   │    Agent    │   │    Agent    │
    │    (C++)    │   │    (C++)    │   │    (C++)    │
    └─────────────┘   └─────────────┘   └─────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
    ┌─────────────────────────────────────────────────────────┐
    │              Storage Repository                         │
    │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
    │  │  Local   │ │   NFS    │ │  S3/Blob │ │  Tape      │ │
    │  └──────────┘ └──────────┘ └──────────┘ └────────────┘ │
    └─────────────────────────────────────────────────────────┘
```

## 11. Типи відновлення

| Тип | Опис |
|-----|------|
| Full VM Restore | Відновлення VM повністю |
| Instant Restore | Монтування бекапу як VM |
| File-Level Recovery (FLR) | Витяг файлів з бекапу |
| Database Restore | Відновлення БД з точки |
| Export | Експорт в інший формат |

## 12. Метрики та моніторинг

- **RPO** (Recovery Point Objective): налаштовується
- **RTO** (Recovery Time Objective): ціль < 1 год
- **Success Rate**: > 99.9%
- **Speed**: real-time tracking
- **Alerts**: Email, Telegram, Slack, Webhook
