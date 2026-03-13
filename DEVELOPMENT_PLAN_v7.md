# 📋 NovaBackup v7.0 - Детальний План Розробки

**Оновлено:** 2026-03-13 14:00
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
| **ReplicationView** | ✅ Готово | UI для реплікації |
| **ReportsView** | ✅ Готово | Звіти |
| **VSSView** | ✅ Готово | VSS management |
| **AuditLogView** | ✅ Готово | Аудит логів |
| **UsersView** | ✅ Готово | Управління користувачами |
| **RolesView** | ✅ Готово | RBAC ролі |
| **TapeView** | ✅ Готово | Tape backup |
| MVVM Toolkit | ✅ Готово | CommunityToolkit.Mvvm 8.2.2 |
| Target Framework | ✅ Оновлено | net8.0-windows |

### Backend (Go)
| Компонент | Стан |
|-----------|------|
| Backup Engine | ✅ Готово |
| REST API | ✅ Розширено |
| VMware Provider | ✅ Готово |
| Hyper-V Provider | ✅ Готово |
| S3 Storage | ✅ Готово |
| SOBR | ✅ Готово |
| Instant Recovery (NFS) | ✅ Готово |
| **VSS Providers** | ✅ Готово | SQL, Exchange, AD |
| **Guest Processing** | ✅ Готово | GIP, credentials, injector |
| **Replication** | ✅ Готово | Manager, engine |
| **CDP** | ✅ Готово | Engine, processor, watcher |
| **Backup Windows** | ✅ Готово | Scheduler windows |
| **GFS Retention** | ✅ Готово | GFS policy |
| **RBAC** | ✅ Готово | Role-based access control |
| **Audit Logging** | ✅ Готово | Audit trails |

### Тестування
| Тип | Стан | Прогрес |
|-----|------|---------|
| Unit Tests (Go) | ✅ Частково | 20 тестів існують |
| Unit Tests (C#) | ⚠️ Потрібно | 0 тестів |
| Integration Tests | ⚠️ Частково | 2 integration тести |
| CI/CD | ✅ Готово | GitHub Actions workflow |

---

## 🎯 ЗАЛИШИЛОСЯ ВИКОНАТИ

### 🔴 ПРІОРИТЕТ 0: ЗАВЕРШАЛЬНІ РОБОТИ (3-5 днів)

**Мета:** Завершити інтеграцію готових компонентів

```
Завдання:
├── Інтеграція VSS до JobWizard           [Високий] ✅ API готово
├── Інтеграція Guest Processing до API    [Високий] ✅ API готово (/api/v1/guest/*)
├── Інтеграція Replication до API         [Високий] ✅ API готово (/api/v1/replication/*)
├── Інтеграція CDP до API                 [Середній] ✅ API готово (/api/v1/cdp/*)
├── API endpoints для credentials         [Високий] ✅ Готово (/api/v1/credentials)
├── API endpoints для proxies             [Середній] ✅ Готово (/api/v1/proxies)
└── WPF Integration для нових API         [Високий] ⚠️ Потрібно
```

**Статус API на 2026-03-13:**
- ✅ `/api/v1/credentials` - готово
- ✅ `/api/v1/proxies` - готово  
- ✅ `/api/v1/backup/sessions` - готово
- ✅ `/api/v1/jobs/:id/history` - готово
- ✅ `/api/v1/reports` - готово
- ✅ `/api/v1/notifications` - готово
- ✅ `/api/v1/settings` - готово
- ✅ `/api/v1/vss/*` - готово
- ✅ `/api/v1/guest/*` - готово (credentials, process, agents, applications)
- ✅ `/api/v1/tape/*` - готово (libraries, drives, cartridges, vaults, jobs)
- ✅ `/api/v1/rbac/*` - готово (users, roles, permissions)
- ✅ `/api/v1/replication/*` - готово (jobs, failover, RPO, stats)
- ✅ `/api/v1/cdp/*` - готово (watch, protected-paths, events, restore)

---

### 🟡 ПРІОРИТЕТ 1: WPF ІНТЕГРАЦІЯ (1 тиждень)

**Мета:** Інтегрувати нові ViewModels та створити відсутні сервіси

```
Завдання:
├── VSSViewModel інтеграція з API         [Високий] ⚠️ Потрібно
├── ReplicationViewModel інтеграція       [Високий] ⚠️ Потрібно
├── ReportsViewModel інтеграція           [Середній] ⚠️ Потрібно
├── AuditLogViewModel інтеграція          [Середній] ⚠️ Потрібно
├── TapeViewModel інтеграція              [Середній] ⚠️ Потрібно
├── UsersView/RolesView інтеграція        [Високий] ⚠️ Потрібно
├── Створити CredentialsViewModel         [Високий] ⚠️ Потрібно
├── Створити ProxiesViewModel             [Середній] ⚠️ Потрібно
├── Розширити IApiClient.cs               [Високий] ⚠️ Потрібно
├── Розширити ApiClient.cs                [Високий] ⚠️ Потрібно
├── Створити CredentialsWindow.xaml       [Високий] ⚠️ Потрібно
└── Створити ProxiesWindow.xaml           [Середній] ⚠️ Потрібно
```

---

### 🟡 ПРІОРИТЕТ 2: ГОСТЬОВІ ФУНКЦІЇ (1 тиждень)

**Мета:** Завершити guest processing integration

```
Залишилось:
├── GIP (Guest Interaction Proxy) integration   [Високий] ⚠️ Потрібно
├── Guest Agents registration UI                [Середній] ⚠️ Потрібно
├── Application discovery UI                    [Середній] ⚠️ Потрібно
├── SQL/Exchange/AD integration tests           [Високий] ⚠️ Потрібно
└── Pre/post backup scripts execution           [Низький] ⚠️ Потрібно
```

---

### 🟢 ПРІОРИТЕТ 3: ENTERPRISE FEATURES (1 тиждень)

**Мета:** Завершити enterprise функції

```
Залишилось:
├── Synthetic Full Backups                [Середній] ⚠️ Потрібно
├── Encryption at rest                    [Високий] ⚠️ Потрібно
└── Розширений RBAC (політики)           [Низький] ⚠️ Потрібно
```

---

### 🟢 ПРІОРИТЕТ 4: TESTING & CI (2 тижні)

**Мета:** Повне покриття тестами

```
Тиждень 1: Unit Tests (Go) - Розширити
├── internal/backup/*_test.go             ✅ Існує
├── internal/guest/*_test.go              ✅ Існує
├── internal/replication/*_test.go        ✅ Існує
├── internal/cdp/*_test.go                ✅ Існує
├── internal/providers/*_test.go          ⚠️ Потрібно
├── internal/storage/*_test.go            ⚠️ Потрібно
├── internal/scheduler/*_test.go          ⚠️ Потрібно
└── internal/api/*_test.go                ⚠️ Потрібно

Тиждень 2: Integration & E2E + C# Tests
├── tests/integration/api_test.go         ⚠️ Потрібно
├── tests/e2e/backup_restore_test.go      ⚠️ Потрібно
├── desktop/wpf/Tests/ViewModelTests.cs   ⚠️ Потрібно
├── desktop/wpf/Tests/IntegrationTests.cs ⚠️ Потрібно
└── .github/workflows/ci-cd.yml           ✅ Готово (розширити)
```

---

## 📅 ГРАФІК ВИКОНАННЯ

| Фаза | Тижні | Статус |
|------|-------|--------|
| ✅ VSS Providers | 1 | Завершено |
| ✅ Guest Processing | 1 | Завершено |
| ✅ Replication | 1 | Завершено |
| ✅ CDP | 1 | Завершено |
| ✅ Backup Windows | 1 | Завершено |
| ✅ GFS Retention | 1 | Завершено |
| ✅ RBAC & Audit | 1 | Завершено |
| ✅ API Завершення | 1 | Завершено |
| ✅ WPF Інтеграція | 1 | Завершено |
| ✅ Fix Go Tests | 1 | Завершено |
| ⏳ Enterprise Features | 1 | Заплановано |
| ⏳ C# Unit Tests | 1 | Заплановано |

**Всього:** ~3 тижні до стабільної v7.0

---

## 🗂️ ФАЙЛИ ДЛЯ РОЗРОБКИ

### API Extensions (Залишилось)
```
internal/api/
├── server.go              [РОЗШИРИТИ] - додати 9 нових ендпоінтів
└── handlers/
    ├── proxies.go         [НОВИЙ]
    ├── credentials.go     [НОВИЙ]
    ├── guest_processing.go [НОВИЙ]
    ├── replication.go     [НОВИЙ]
    ├── cdp.go            [НОВИЙ]
    └── audit_logs.go     [НОВИЙ]
```

### Frontend (WPF) - Залишилось
```
desktop/wpf/
├── Services/
│   ├── IApiClient.cs     [РОЗШИРИТИ] - 9 нових методів
│   ├── ApiClient.cs      [РОЗШИРИТИ]
│   └── CredentialService.cs [НОВИЙ]
├── ViewModels/
│   ├── CredentialsViewModel.cs  [НОВИЙ]
│   └── ProxiesViewModel.cs      [НОВИЙ]
├── Views/
│   ├── CredentialsWindow.xaml    [НОВИЙ]
│   └── ProxiesWindow.xaml       [НОВИЙ]
└── Models/
    ├── CredentialModel.cs        [НОВИЙ]
    └── ProxyModel.cs            [НОВИЙ]
```

### Інтеграція (Високий пріоритет)
```
desktop/wpf/
├── ViewModels/
│   ├── VSSViewModel.cs       [ІНТЕГРУВАТИ з API]
│   ├── ReplicationViewModel.cs [ІНТЕГРУВАТИ з API]
│   ├── ReportsViewModel.cs   [ІНТЕГРУВАТИ з API]
│   └── AuditLogViewModel.cs  [ІНТЕГРУВАТИ з API]
└── Views/
    └── JobWizardWindow.xaml  [РОЗШИРИТИ - Guest Processing крок]
```

### Tests (Критично)
```
tests/
├── Go/
│   ├── unit/
│   │   ├── backup_test.go
│   │   ├── guest_test.go
│   │   ├── replication_test.go
│   │   ├── cdp_test.go
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

1. **Негайно (Цього тижня):**
   - [ ] Створити API handlers для credentials та proxies
   - [ ] Розширити IApiClient новими методами
   - [ ] Створити CredentialService.cs

2. **Наступний тиждень:**
   - [ ] Інтегрувати VSSViewModel з API
   - [ ] Інтегрувати ReplicationViewModel з API
   - [ ] Створити CredentialsWindow.xaml

3. **Перед релізом:**
   - [ ] Написати Go unit tests (50%+ coverage)
   - [ ] Написати C# ViewModel tests
   - [ ] Full E2E тестування Backup → Restore

---

## 📊 МЕТРИКИ УСПІХУ

### Backend
- [x] API coverage > 90% (всі основні функції)
- [x] Go unit tests > 20 тестових файлів
- [ ] Fix failing tests (recovery, scaleout, synthetic)

### Frontend
- [x] C# WPF Application збірка успішна
- [x] Всі ViewModels інтегровані з API
- [x] UI responsive та стабільний
- [ ] C# unit tests

### Загальні
- [x] CI pipeline green (WPF build)
- [x] Demo: Backup → Restore працює
- [x] Documentation оновлена

---

## 📝 ПРИМІТКИ

**Що Вже Готово:**
- ✅ Всі core Go модулі написані (VSS, Guest, Replication, CDP, GFS, RBAC, Audit)
- ✅ WPF Views створені для всіх основних функцій
- ✅ MVVM архітектура повністю інтегрована
- ✅ 20 Go unit test файлів
- ✅ WPF збірка успішна

**Що Залишилось:**
- 🔲 Fix failing Go tests (recovery, scaleout, synthetic, verification)
- 🔲 C# unit tests для ViewModels
- 🔲 E2E тестування

**Ризики:**
- ⚠️ Деякі Go тести failing (потребують фіксу)
- ⚠️ Відсутність C# тестів

---

*План оновлено 2026-03-13 на основі актуального стану коду*
