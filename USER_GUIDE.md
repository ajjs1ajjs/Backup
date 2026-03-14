# NovaBackup - Посібник користувача

## Швидкий старт

### 1. Запуск програми

```bash
dotnet run --project NovaBackup.GUI/NovaBackup.GUI.csproj
```

Або після встановлення:
- Пуск → NovaBackup → NovaBackup Backup Console

### 2. Створення першого Backup Job

**Спосіб 1: Через Ribbon**
1. Натисніть вкладку **Home** → **Backup Job** (💾)
2. Введіть назву джобу
3. Додайте файли або папки для backup
4. Виберіть сховище (Repository)
5. Налаштуйте розклад (Schedule)
6. Натисніть **Save**

**Спосіб 2: Через вкладку Jobs**
1. Натисніть вкладку **Jobs** на Ribbon
2. Натисніть **New Job**
3. Заповніть форму як вище

### 3. Запуск Backup

1. Виберіть джоб у таблиці або дереві
2. На Ribbon натисніть **Jobs** → **Run** (▶)
3. Або права кнопка на джобі → Run

### 4. Перегляд результатів

- **Dashboard** - загальна статистика
- **Monitoring** → **Sessions** - історія сесій
- **Properties Panel** (справа) - деталі вибраного об'єкта

## Опис інтерфейсу

### Ribbon меню

#### Home
- **Backup Job** (💾) - створити новий джоб
- **Restore** (🔄) - перейти до відновлення файлів

#### Jobs
- **New Job** (➕) - створити джоб
- **Edit** (✏️) - редагувати вибраний джоб
- **Delete** (🗑️) - видалити вибраний джоб
- **Run** (▶) - запустити вибраний джоб
- **Disable** (⏸️) - вимкнути/увімкнути джоб

#### Infrastructure
- **Repository** (📁) - додати сховище
- **Server** (🖥️) - додати сервер
- **Rescan** (🔄) - оновити сховище

#### Monitoring
- **Sessions** (📊) - перегляд сесій
- **Logs** (📋) - журнали
- **Alarms** (⚠️) - сповіщення

### Навігаційне дерево (зліва)

```
🏠 Home
  └── 📊 Dashboard
💾 Backup Jobs
  ├── 📁 Active Jobs
  └── 📜 Job Sessions
🔄 Restore
📁 Infrastructure
  ├── 💿 Repositories
  ├── 🖥️ Servers
  └── ☁️ Cloud Connect
📊 Monitoring
  ├── 📋 Sessions
  ├── ⚠️ Alarms
  └── 📈 Events
```

### Робоча область (центр)

#### Dashboard
Показує:
- **Total Backups** - загальна кількість backup'ів
- **Storage Used** - використано місця
- **Success Rate** - відсоток успішних операцій
- **Next Job** - наступний запланований джоб
- **Recent Job Sessions** - таблиця останніх сесій

#### Backup Jobs
Таблиця з колонками:
- **Name** - назва джобу
- **Type** - тип (Full/Incremental/Differential)
- **Schedule** - розклад (Manual/Daily/Weekly)
- **Last Run** - останній запуск
- **Next Run** - наступний запуск
- **Enabled** - чи увімкнено

#### Repositories
Таблиця з колонками:
- **Name** - назва сховища
- **Path** - шлях до сховища
- **Type** - тип (Local/SMB/S3)
- **Used** - використано
- **Free** - вільно

### Панель властивостей (справа)

Показує деталі вибраного об'єкта:

**Для Backup Job:**
- Job Name
- Status (Enabled/Disabled)
- Schedule
- Last Run
- Next Run
- Description

**Для Repository:**
- Repository Name
- Path
- Type
- Storage (Usage bar + відсоток)

### Статус бар (знизу)

- **Status** - поточний стан
- **Jobs count** - кількість джобів
- **Connection** - статус підключення
- **Time** - поточний час
- **Version** - версія програми

## Гарячі клавіші

| Клавіша | Дія |
|---------|-----|
| F5 | Оновити (Refresh) |
| Ctrl+N | Новий джоб |
| Ctrl+E | Редагувати джоб |
| Delete | Видалити джоб |
| F9 | Запустити джоб |

## Приклади використання

### Створення щоденного backup

1. **Home** → **Backup Job**
2. Назва: `Daily Backup`
3. Додайте папки: `C:\ImportantData`
4. Сховище: `Local Repository`
5. Schedule: `Daily`
6. Час: `02:00`
7. **Save**

### Створення тижневого backup

1. **Jobs** → **New Job**
2. Назва: `Weekly Full Backup`
3. Тип: `Full`
4. Schedule: `Weekly`
5. Дні: `Sunday`
6. Час: `03:00`
7. **Save**

### Відновлення файлів

1. **Restore** в дереві
2. Виберіть backup зі списку
3. Виберіть restore point
4. Виберіть режим:
   - Restore to original location
   - Restore to new location
5. Вкажіть шлях
6. **Restore**

### Вимкнення джобу

1. Виберіть джоб у таблиці
2. **Jobs** → **Disable**
3. Або права кнопка → Disable

## Усунення проблем

### Джоб не запускається
- Перевірте чи Enabled джоб
- Перевірте чи існує сховище
- Перевірте права доступу до файлів

### Немає вільного місця
- Додайте нове сховище
- Очистіть старі backup'и
- Налаштуйте Retention Policy

### Помилка доступу до сховища
- Перевірте шлях до сховища
- Для SMB перевірте credentials
- Натисніть **Rescan**

## Advanced Features

### 🔐 Encryption at Rest (Шифрування даних)

NovaBackup підтримує військове шифрування AES-256-GCM для захисту ваших резервних копій.

#### Увімкнення шифрування

**При створенні Backup Job:**
1. Відкрийте **Job Wizard**
2. На кроці **Settings** знайдіть розділ **Encryption**
3. Виберіть рівень шифрування:
   - **AES-256-GCM** (рекомендовано) - максимальний захист
   - **None** - без шифрування (швидше)
4. Введіть пароль шифрування (мінімум 8 символів)
5. Збережіть пароль у безпечному місці!

**Важливо:**
- ⚠️ Без правильного пароля неможливо відновити дані
- ⚠️ Пароль не зберігається в базі даних
- ⚠️ Для зміни пароля потрібно створити новий backup

#### Переваги шифрування
- ✅ Захист від несанкціонованого доступу
- ✅ Відповідність вимогам GDPR, HIPAA
- ✅ Безпека при зберіганні в хмарі
- ✅ Захист від фізичної крадіжки дисків

#### Продуктивність
- Шифрування додає ~5-10% до часу backup
- Використовується апаратне прискорення AES-NI
- Стиснення відбувається ПЕРЕД шифруванням

---

### 🔄 Synthetic Full Backups

Synthetic Full Backup створює повну резервну копію з інкрементальних, не навантажуючи джерело.

#### Як це працює

```
День 1: [Full Backup]
День 2: [Incremental] → зміни з Дня 1
День 3: [Incremental] → зміни з Дня 2
День 4: [Synthetic Full] ← об'єднує всі інкрементальні
```

**Переваги:**
- 🚀 Немає навантаження на джерело (production)
- 🚀 Швидше створення ніж повний backup
- 🚀 Менше мережевого трафіку
- 🚀 Оптимальне використання сховища

#### Створення Synthetic Full Backup

**Через UI:**
1. Відкрийте **Management** → **Synthetic Backup**
2. Натисніть **Create Synthetic Backup**
3. Виберіть параметри:
   - **Source Repo** - джерело інкрементальних backup'ів
   - **Target Repo** - сховище для synthetic full
   - **Backup Type** - Full або Incremental
   - **Compression** - увімкнути стиснення
   - **Retention Days** - скільки зберігати
4. Натисніть **Create**

**Через API:**
```bash
POST /api/v1/synthetic
{
  "source_repo": "Repo-01",
  "target_repo": "SOBR-01",
  "backup_type": "Full",
  "compression": true,
  "retention_days": 30
}
```

#### Merge Incrementals

Функія об'єднує інкрементальні backup'и в synthetic full:

1. Виберіть існуючий synthetic backup
2. Натисніть **Merge Incrementals**
3. Виберіть дату **Since** - з якої міняти інкрементальні
4. Натисніть **Merge**

**Коли використовувати:**
- Після серії інкрементальних backup'ів
- Перед видаленням старих точок відновлення
- Для оптимізації сховища

#### Моніторинг та статистика

**Dashboard Synthetic Backup показує:**
- Total Backups - загальна кількість
- Total Chains - ланцюжки backup'ів
- Active Backups - поточні операції
- Total Size - розмір у сховищі
- Compression Ratio - середнє стиснення
- Last Activity - остання операція

#### Best Practices

**Рекомендації:**
- 📅 Створюйте Synthetic Full щотижня
- 📅 Зберігайте 4-6 тижневих точок
- 📅 Використовуйте окремий repo для synthetic
- 📅 Увімкніть compression для економії місця

**Уникати:**
- ❌ Не створюйте synthetic full щодня (немає сенсу)
- ❌ Не видаляйте інкрементальні до завершення merge
- ❌ Не використовуйте для дуже малих backup'ів (<1GB)

---

## Додаткова допомога

- **F1** - довідка (в розробці)
- **?** - інформація про об'єкт
- **Права кнопка** - контекстне меню

## Контакти

- Сайт: https://novabackup.local
- Підтримка: support@novabackup.local
- Документація: https://novabackup.local/docs
