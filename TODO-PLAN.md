TODO PLAN (для майбутніх патчів)

Поточний статус (після readiness)
- Ready (поточна точка):
  - Windows Hyper-V провайдер: стабільний збір VM та маппінг станів
  - Linux Libvirt провайдер: інтегрований та повертає реальні VM за наявності libvirt
  - SA DB бекенд з підтримкою діалектів (PostgreSQL, MSSQL, Oracle) через SA_DBManager; BackupManager може з SA DB або з JSON/SQLite бекендом
  - Cloud каркас: AWS, Azure, GCP cloud providers (кільцеві каркаси) з Mock‑провайдером для CI; можливість підключення реальних SDK
  - API на FastAPI: ендпойнти /vms, /normalize, /backups, /backups/{backup_id}/restore; API ключ (NOVABACKUP_API_KEY)
  - CLI: backup create/list/restore; інтеграція з SA DB та cloud шляхами
  - Dashboard: /static/dashboard.html
  - Інсталяція: install.sh (Linux/macOS), install.bat (Windows)
  - CI: pytest, mypy, flake8; діалектні тести для SA DB
  - Міграція: backups.json → DB

- Not Ready (не готово):
  - Реальні cloud backup/restore операції через AWS/Azure/GCP SDK
  - Повна RBAC/OAuth2/JWT авторизація та розширені ролі
  - Розширений фронтенд UI/framework UI
  - Контейнеризація CI для діалектів БД та cloud‑провайдерів
  - Детальна документація для міграції та багатодіалектних кейсів

Next steps (після цієї збірки):
- Реалізувати реальні cloud‑бекенди (AWS/Azure/GCP) з відповідними SDK та прикладами використання
- Додати RBAC/OAuth2/JWT та більш детальну обробку помилок
- Розробити end-to-end тести для cloud сценаріїв
- Оновити CI з повноцінними окруженнями діалектів БД та cloud
- Оновити документацію та релізний процес
