import json
import sqlite3
from pathlib import Path

from novabackup.migrate import migrate_json_to_db


def test_migrate_json_to_db(tmp_path: Path):
    # Prepare a small backups.json in a temp path
    json_path = tmp_path / "backups.json"
    data = {
        "bak1": {
            "backup_id": "bak1",
            "vm_id": "vm1",
            "dest": "./backups",
            "type": "full",
            "provider": "Mock",
            "status": "created",
            "created_at": "2026-01-01T00:00:00Z",
            "snapshot": {
                "snapshot_id": "snap1",
                "provider": "Mock",
                "vm_id": "vm1",
                "dest": "./backups",
                "type": "full",
                "name": "snap",
                "status": "created",
            },
        }
    }
    json_path.write_text(json.dumps(data))

    db_path = tmp_path / "novabackup.db"
    db_url = f"sqlite:///{db_path}"
    result = migrate_json_to_db(json_path=str(json_path), database_url=db_url)

    assert isinstance(result, dict)
    assert result.get("migrated_backups", 0) >= 1

    # Verify that data was written into the SQLite DB
    conn = sqlite3.connect(db_path)
    cur = conn.cursor()
    cur.execute("SELECT COUNT(*) FROM backups")
    count = cur.fetchone()[0]
    assert count >= 1
    conn.close()
