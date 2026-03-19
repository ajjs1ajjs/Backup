#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
import time

import requests

BASE_URL = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"

headers = {"Authorization": TOKEN, "Content-Type": "application/json"}

print("=" * 60)
print("NovaBackup - Full Backup/Restore Test")
print("=" * 60)

# Step 1: Create job
print("\n[1/5] Creating backup job...")
job_data = {
    "name": "Full Test Backup",
    "sources": ["D:\\DOCUMENTS\\MY_SOFT\\KeepActive.exe"],
    "destination": "D:\\Backup\\FullTest",
    "type": "file",
    "compression": True,
    "encryption": False,
    "schedule": "manual",
    "enabled": True,
}

resp = requests.post(f"{BASE_URL}/api/jobs", headers=headers, json=job_data)
result = resp.json()
if not result.get("success"):
    print(f"ERROR: {result.get('error')}")
    exit(1)

job = result.get("job", {})
job_id = job.get("id")
print(f"✓ Job created: {job_id}")
print(f"  Name: {job.get('name')}")
print(f"  Source: {job.get('sources')}")
print(f"  Dest: {job.get('destination')}")

# Step 2: Run backup
print("\n[2/5] Running backup...")
time.sleep(1)
resp = requests.post(f"{BASE_URL}/api/jobs/{job_id}/run", headers=headers)
result = resp.json()

if not result.get("success"):
    print(f"ERROR: {result.get('error')}")
    exit(1)

session = result.get("session", {})
backup_path = session.get("backup_path")
print(f"✓ Backup completed!")
print(f"  Session: {session.get('id')}")
print(f"  Status: {session.get('status')}")
print(f"  Files: {session.get('files_processed', 0)}/{session.get('files_total', 0)}")
print(
    f"  Size: {session.get('bytes_read', 0)} bytes read, {session.get('bytes_written', 0)} bytes written"
)
print(f"  Path: {backup_path}")

# Step 3: Verify backup file exists
print("\n[3/5] Verifying backup file...")
import os

if backup_path:
    zip_file = os.path.join(backup_path, "backup.zip")
    if os.path.exists(zip_file):
        size = os.path.getsize(zip_file)
        print(f"✓ Backup file exists: {zip_file}")
        print(f"  Size: {size} bytes")

        # Try to open as ZIP
        import zipfile

        try:
            with zipfile.ZipFile(zip_file, "r") as zf:
                files = zf.namelist()
                print(f"  Files in archive: {len(files)}")
                for f in files[:5]:  # Show first 5
                    print(f"    - {f}")
        except Exception as e:
            print(f"✗ Invalid ZIP: {e}")
    else:
        print(f"✗ Backup file NOT found: {zip_file}")

# Step 4: Restore
print("\n[4/5] Testing restore...")
restore_dest = "D:\\RESTORE_TEST_FULL"
restore_data = {
    "backup_path": backup_path,
    "files": [],  # All files
    "destination": restore_dest,
    "restore_original": False,
}

resp = requests.post(
    f"{BASE_URL}/api/restore/files", headers=headers, json=restore_data
)
result = resp.json()

if result.get("success"):
    restore_session = result.get("session", {})
    print(f"✓ Restore completed!")
    print(f"  Status: {restore_session.get('status')}")
    print(f"  Files restored: {restore_session.get('files_restored', 0)}")
else:
    print(f"✗ Restore failed: {result.get('error')}")

# Step 5: Verify restored files
print("\n[5/5] Verifying restored files...")
if os.path.exists(restore_dest):
    restored_files = []
    for root, dirs, files in os.walk(restore_dest):
        for file in files:
            restored_files.append(os.path.join(root, file))

    print(f"✓ Restored {len(restored_files)} files:")
    for f in restored_files[:5]:
        print(f"    - {f}")
else:
    print(f"✗ Restore directory not found: {restore_dest}")

print("\n" + "=" * 60)
print("TEST COMPLETE")
print("=" * 60)
