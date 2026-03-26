# Migration Guide: From JSON Store to SQL Database

This guide describes how to migrate your existing Novabackup data from the default JSON store (`backups.json`) to a SQL database (SQLite, PostgreSQL, Microsoft SQL Server, or Oracle) using the provided migration script.

## Overview

Novabackup supports two storage backends:
1. **JSON store** (default): Stores backups in a JSON file (`backups.json`) in the working directory.
2. **SQL store**: Uses SQLAlchemy to store backups in a relational database. Supported dialects: SQLite, PostgreSQL, MSSQL, Oracle.

The migration script (`novabackup/migrate.py`) performs a one-time migration from the JSON store to the SQL store. After migration, you can configure Novabackup to use the SQL store by setting the `NOVABACKUP_DATABASE_URL` environment variable.

## Prerequisites

- Python 3.10 or newer.
- The `novabackup` package installed (with the `db` extra for SQL support).
- Access to the existing `backups.json` file (if you have been using the JSON store).
- For PostgreSQL, MSSQL, or Oracle: appropriate database server running and accessible.

## Step-by-Step Migration

### 1. Install Required Dependencies

If you haven't already, install the Novabackup package with the `db` extra:

```bash
pip install novabackup[db]
```

This installs SQLAlchemy and the necessary DB-API drivers for SQLite (built-in), PostgreSQL (requires `psycopg2-binary`), MSSQL (requires `pyodbc`), and Oracle (requires `cx_Oracle`).

> **Note**: For production use, you may want to install the specific driver for your target database. The `db` extra includes the basic SQLAlchemy package; you may need to install the driver separately.

### 2. Prepare the Target Database

#### Option A: SQLite (file-based)

No preparation needed; the migration script will create the database file if it doesn't exist.

#### Option B: PostgreSQL

1. Create a database and user:
   ```sql
   CREATE DATABASE novabackup;
   CREATE USER novabackup_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE novabackup TO novabackup_user;
   ```
2. Note the connection URL format:
   ```
   postgresql://novabackup_user:your_password@localhost:5432/novabackup
   ```

#### Option C: Microsoft SQL Server

1. Create a database and login:
   ```sql
   CREATE DATABASE novabackup;
   GO
   CREATE LOGIN novabackup_user WITH PASSWORD = 'your_password';
   GO
   USE novabackup;
   GO
   CREATE USER novabackup_user FOR LOGIN novabackup_user;
   GO
   EXEC sp_addrolemember 'db_owner', 'novabackup_user';
   ```
2. Note the connection URL format (using pyodbc):
   ```
   mssql+pyodbc://novabackup_user:your_password@localhost:1433/novabackup?driver=ODBC+Driver+17+for+SQL+Server
   ```
   Adjust the driver as needed.

#### Option D: Oracle

1. Create a user and grant privileges:
   ```sql
   CREATE USER novabackup_user IDENTIFIED BY your_password;
   GRANT CONNECT, RESOURCE TO novabackup_user;
   ```
2. Note the connection URL format:
   ```
   oracle+cx_oracle://novabackup_user:your_password@localhost:1521/?service_name=orcl
   ```

### 3. Run the Migration Script

Set the `NOVABACKUP_DATABASE_URL` environment variable to point to your target database, then run the migration script.

#### Example: SQLite (file)
```bash
export NOVABACKUP_DATABASE_URL="sqlite:///./novabackup.db"  # Linux/macOS
set NOVABACKUP_DATABASE_URL=sqlite:///.\novabackup.db       # Windows
python -m novabackup.migrate
```

#### Example: PostgreSQL
```bash
export NOVABACKUP_DATABASE_URL="postgresql://novabackup_user:your_password@localhost:5432/novabackup"
python -m novabackup.migrate
```

#### Example: MSSQL
```bash
export NOVABACKUP_DATABASE_URL="mssql+pyodbc://novabackup_user:your_password@localhost:1433/novabackup?driver=ODBC+Driver+17+for+SQL+Server"
python -m novabackup.migrate
```

#### Example: Oracle
```bash
export NOVABACKUP_DATABASE_URL="oracle+cx_oracle://novabackup_user:your_password@localhost:1521/?service_name=orcl"
python -m novabackup.migrate
```

The script will output a summary like:
```
{"migrated_backups": 42, "migrated_restores": 0}
```

### 4. Verify the Migration

After migration, you can verify that the data is in the database by checking the `backups` table.

#### Example: Using sqlite3 CLI (for SQLite)
```bash
sqlite3 novabackup.db "SELECT * FROM backups LIMIT 5;"
```

#### Example: Using psql (for PostgreSQL)
```bash
psql "postgresql://novabackup_user:your_password@localhost:5432/novabackup" -c "SELECT * FROM backups LIMIT 5;"
```

### 5. Configure Novabackup to Use the SQL Store

To make Novabackup use the SQL store by default, set the `NOVABACKUP_DATABASE_URL` environment variable in your deployment environment (e.g., in your shell, Docker compose, or systemd service).

#### Example: Docker Compose (local development)
Update `docker-compose.yml`:
```yaml
api:
  # ...
  environment:
    - NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db
```

#### Example: Systemd Service
Edit the service file (`deploy/systemd/novabackup.service`):
```ini
Environment="NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db"
```

### 6. (Optional) Remove the Old JSON File

After verifying that the migration was successful and Novabackup is working with the SQL store, you can safely remove or archive the old `backups.json` file.

## Troubleshooting

- **Migration script returns zero migrated backups**: Ensure that the `backups.json` file exists in the current working directory (or provide the path via `--json-path` if you modify the script; currently the script uses the current working directory). Also, check that the JSON file is in the expected format (a dictionary mapping backup IDs to job objects).

- **Database connection errors**: Verify the `NOVABACKUP_DATABASE_URL` format and that the database server is accessible. Test the connection with a simple script or CLI tool.

- **Permission errors**: Ensure the user running the migration has read access to `backups.json` and write access to the target database file (for SQLite) or sufficient privileges in the target database.

## Automating Migration in Production

For production deployments, you can run the migration script as part of your deployment process before starting the application. Ensure that the migration is idempotent (the script uses `INSERT OR IGNORE` so running it multiple times will not create duplicates).

## Further Reading

- The migration script is located at `novabackup/migrate.py`.
- The SQLAlchemy backend is in `novabackup/db_sa.py`.
- The main `BackupManager` class (in `novabackup/backup.py`) automatically selects the SQL store if `NOVABACKUP_DATABASE_URL` is set and the SQLAlchemy dependencies are available.

---
*This guide applies to Novabackup version 0.2.0 and later.*