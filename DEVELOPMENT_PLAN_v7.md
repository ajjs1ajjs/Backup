# 📋 NovaBackup v7.0 - Детальний План Розробки

**Оновлено:** 2026-03-12  
**Версія:** 7.0  
**Мета:** Повноцінна заміна Veeam Backup & Replication

---

## ✅ Поточний Стан (Що Зроблено)

### WPF Desktop Application
| Компонент | Стан | Нотатки |
|-----------|------|---------|
| HomeView | ✅ Готово | Dashboard з статусом, активними сесіями |
| JobsView | ✅ Готово | Таблиця джобів + restore points |
| JobWizard | ✅ Готово | 6-кроковий wizard |
| StorageView | ✅ Готово | Таблиця репозиторіїв |
| InfrastructureView | ✅ Готово | Дерево серверів |
| RecoverySessions | ✅ Готово | MVVM інтеграція |
| MVVM Toolkit | ✅ Готово | CommunityToolkit.Mvvm 8.2.2 |
| Target Framework | ✅ Оновлено | net8.0-windows |

### Backend (Go)
| Компонент | Стан |
|-----------|------|
| Backup Engine | ✅ Готово |
| REST API | ⚠️ Базовий |
| VMware Provider | ✅ Готово |
| Hyper-V Provider | ✅ Готово |
| S3 Storage | ✅ Готово |
| SOBR | ✅ Готово |
| Instant Recovery (NFS) | ✅ Готово |

### Тестування
| Тип | Стан |
|-----|------|
| Unit Tests (C#) | ⚠️ 8 тестів |
| Integration Tests | ⚠️ Базові |
| CI/CD | ⚠️ Базовий workflow |

---

## 🎯 ПРІОРИТЕТНІ НАПРЯМКИ

### ПРІОРИТЕТ 1: API РОЗШИРЕННЯ (2 тижні)

**Мета:** Повноцінний REST API для всіх функцій

```
Тиждень 1:
├── /api/v1/proxies          - Управління backup proxies
├── /api/v1/credentials      - Менеджер паролів
├── /api/v1/backup/sessions  - Активні сесії бекапу
└── /api/v1/jobs/:id/history - Історія виконання

Тиждень 2:
├── /api/v1/restore/*       - Повний restore API
├── /api/v1/reports          - Звіти
├── /api/v1/notifications    - Нотифікації
└── /api/v1/settings         - Налаштування системи
```

**Файли для реалізації:**
- `internal/api/server.go` - додати нові ендпоінти
- `desktop/wpf/Services/IApiClient.cs` - розширити інтерфейс
- `desktop/wpf/Services/ApiClient.cs` - реалізувати клієнт

---

### ПРІОРИТЕТ 2: ГОСТЬОВА ОБРОБКА (3 тижні)

**Мета:** Application-Aware backup для SQL, Exchange, AD

```
Тиждень 1: VSS Integration
├── pkg/providers/vss/vss.go           - Core VSS writer
├── pkg/providers/vss/sqlserver.go     - SQL Server VSS
├── pkg/providers/vss/exchange.go      - Exchange VSS
└── pkg/providers/vss/activedirectory.go - AD VSS

Тиждень 2: Guest Processing Service
├── internal/guest/gip.go              - Guest Interaction Proxy
├── internal/guest/credentials.go      - Credential management
└── internal/guest/injector.go        - Agent injection

Тиждень 3: Integration
├── Додати GuestProcessing до JobModel
├── Розширити JobWizard (крок Guest)
└── Тестування
```

---

### ПРІОРИТЕТ 3: REPLICATION & CDP (2 тижні)

**Мета:** Реплікація VM та Continuous Data Protection

```
Тиждень 1: Replication
├── pkg/replication/manager.go         - Реплікація VM
├── pkg/replication/scheduler.go      - Планування реплікацій
└── cmd/nova-cli/replication.go       - CLI команди

Тиждень 2: CDP
├── internal/cdp/engine.go            - CDP ядро
├── internal/cdp/processor.go         - Обробка змін
└── CDP retention policy
```

---

### ПРІОРИТЕТ 4: ENTERPRISE FEATURES (2 тижні)

**Мета:** Масштабованість та безпека

```
Тиждень 1: Backup Windows & GFS
├── internal/scheduler/windows.go      - Backup windows
├── internal/retention/gfs.go          - GFS retention
└── synthetic full backups

Тиждень 2: Security
├── internal/rbac/                    - Розширений RBAC
├── internal/audit/                   - Аудит логів
└── Encryption at rest
```

---

### ПРІОРИТЕТ 5: TESTING & CI (2 тижні)

**Мета:** Повне покриття тестами

```
Тиждень 1: Unit Tests (Go)
├── internal/backup/*_test.go
├── internal/providers/*_test.go
└── internal/storage/*_test.go

Тиждень 2: Integration & E2E
├── tests/integration/api_test.go
├── tests/e2e/backup_restore_test.go
└── .github/workflows/ - розширити CI
```

---

## 📅 ГРАФІК ВИКОНАННЯ

| Фаза | Тижні | Дата |
|------|-------|------|
| API Розширення | 2 | Тиж 1-2 |
| Guest Processing | 3 | Тиж 3-5 |
| Replication & CDP | 2 | Тиж 6-7 |
| Enterprise Features | 2 | Тиж 8-9 |
| Testing & CI | 2 | Тиж 10-11 |

**Всього:** ~11 тижнів до стабільної v7.0

---

## 🗂️ ФАЙЛИ ДЛЯ РОЗРОБКИ

### API Extensions
```
internal/api/
├── server.go              [РОЗШИРИТИ]
├── handlers/
│   ├── proxies.go         [НОВИЙ]
│   ├── credentials.go     [НОВИЙ]
│   ├── sessions.go        [НОВИЙ]
│   └── reports.go        [НОВИЙ]
```

### Guest Processing
```
internal/guest/
├── gip.go                 [НОВИЙ]
├── credentials.go         [НОВИЙ]
├── injector.go            [НОВИЙ]
└── processor.go          [НОВИЙ]

pkg/providers/vss/
├── vss.go                [РОЗШИРИТИ]
├── sqlserver.go          [ДОПРАЦЮВАТИ]
├── exchange.go           [ДОПРАЦЮВАТИ]
└── activedirectory.go    [ДОПРАЦЮВАТИ]
```

### Frontend (WPF)
```
desktop/wpf/
├── Services/
│   ├── IApiClient.cs     [РОЗШИРИТИ]
│   ├── ApiClient.cs      [ОНОВИТИ]
│   └── CredentialService.cs [НОВИЙ]
├── ViewModels/
│   ├── CredentialsViewModel.cs  [НОВИЙ]
│   ├── ProxiesViewModel.cs      [НОВИЙ]
│   └── ReportsViewModel.cs      [НОВИЙ]
├── Views/
│   ├── CredentialsWindow.xaml    [НОВИЙ]
│   ├── ProxiesWindow.xaml       [НОВИЙ]
│   └── ReportsView.xaml         [НОВИЙ]
└── Models/
    ├── CredentialModel.cs        [НОВИЙ]
    └── ProxyModel.cs            [НОВИЙ]
```

### Tests
```
tests/
├── Go/
│   ├── unit/
│   │   ├── backup_test.go
│   │   └── storage_test.go
│   └── integration/
│       └── api_test.go
└── C#/
    ├── Integration/
    │   └── ApiClientIntegrationTests.cs
    └── Unit/
        └── ViewModelTests.cs
```

---

## 🚀 НАСТУПНІ КРОКИ

1. **Негайно:** Розширити REST API (пріоритет 1)
2. **Після API:** Guest Processing (пріоритет 2)
3. **Паралельно:** Додати Go unit тести
4. **Перед релізом:** Full E2E тестування

---

## 📊 МЕТРИКИ УСПІХУ

- [ ] API coverage > 80%
- [ ] Go unit tests > 50% 
- [ ] C# unit tests > 60%
- [ ] Zero blocking bugs
- [ ] CI pipeline green
- [ ] Demo: Backup → Restore працює

---

*План створено на основі аналізу коду та Roadmap*
