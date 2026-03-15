# NovaBackup Enterprise v6.0

Сучасна система резервного копіювання для Windows.

## Швидкий старт

### Вимоги
- Windows 10/11 або Windows Server 2019+
- Go 1.25+
- .NET 8.0 (для GUI)

### Збірка

```bash
# Зібрати службу Windows
go build -o nova.exe ./cmd/nova-service/

# Зібрати GUI
cd cmd/nova-wpf
dotnet build -c Release
```

### Встановлення

1. Відкрийте командний рядок **від імені адміністратора**
2. Встановіть службу:
   ```cmd
   nova.exe install
   ```
3. Запустіть службу:
   ```cmd
   nova.exe start
   ```

### Використання

Запустіть `NovaBackup.exe` для доступу до графічного інтерфейсу.

## Структура

```
├── cmd/
│   └── nova-service/     # Служба Windows
│   └── nova-wpf/         # GUI додаток
├── internal/             # Внутрішні пакунки Go
└── go.mod                # Go модуль
```

## Ліцензія

MIT
