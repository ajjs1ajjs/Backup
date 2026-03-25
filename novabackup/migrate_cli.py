import argparse
import json
from novabackup.migrate import migrate_json_to_db


def main():
    parser = argparse.ArgumentParser(description="Migrate backups.json to SQLite DB")
    parser.add_argument("--db-url", help="DB URL, e.g. sqlite:///./novabackup.db")
    parser.add_argument("--json-path", help="Path to backups.json")
    parser.add_argument(
        "--dry-run", action="store_true", help="Dry-run: do not write changes"
    )
    args = parser.parse_args()
    summary = migrate_json_to_db(json_path=args.json_path, database_url=args.db_url)
    if args.dry_run:
        print(json.dumps({"dry_run_summary": summary}, indent=2))
    else:
        print(json.dumps({"migration_summary": summary}, indent=2))


if __name__ == "__main__":
    main()
