#!/usr/bin/env python3
import json

import requests

BASE_URL = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"

headers = {"Authorization": TOKEN, "Content-Type": "application/json"}

# Test restore
backup_path = r"D:\DOCUMENTS\MY_SOFT\1\2026-03-16_100812"
dest = r"D:\RESTORE_TEST"

data = {
    "backup_path": backup_path,
    "files": [],  # Empty = restore all
    "destination": dest,
    "restore_original": False,
}

print("=" * 60)
print("Testing Restore API")
print("=" * 60)
print(f"Backup Path: {backup_path}")
print(f"Destination: {dest}")
print()

try:
    response = requests.post(
        f"{BASE_URL}/api/restore/files", headers=headers, json=data, timeout=30
    )

    print(f"Status: {response.status_code}")
    result = response.json()
    print(f"Response: {json.dumps(result, indent=2, ensure_ascii=False)}")

    if result.get("success"):
        print("\n✅ RESTORE SUCCESSFUL!")
        session = result.get("session", {})
        print(f"   Session ID: {session.get('id')}")
        print(f"   Status: {session.get('status')}")
        print(f"   Files Restored: {session.get('files_restored', 0)}")
    else:
        print(f"\n❌ RESTORE FAILED: {result.get('error', 'Unknown error')}")

except Exception as e:
    print(f"\n❌ ERROR: {e}")

print("=" * 60)
