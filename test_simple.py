import json
import os
import sys
import time

import requests

BASE = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"
HDRS = {"Authorization": TOKEN, "Content-Type": "application/json"}

print("=== BACKUP/RESTORE TEST ===\n")

# 1. Create job
print("1. Creating job...")
r = requests.post(
    f"{BASE}/api/jobs",
    headers=HDRS,
    json={
        "name": "Test Job",
        "sources": ["D:\\DOCUMENTS\\MY_SOFT\\KeepActive.exe"],
        "destination": "D:\\Backup\\TestRestore",
        "type": "file",
        "compression": True,
        "enabled": True,
    },
)
job = r.json().get("job", {})
jid = job.get("id")
print(f"   Created: {jid}, Name: {job.get('name')}")

# 2. Run backup
print("\n2. Running backup...")
time.sleep(1)
r = requests.post(f"{BASE}/api/jobs/{jid}/run", headers=HDRS)
res = r.json()
if not res.get("success"):
    print(f"   ERROR: {res.get('error')}")
    sys.exit(1)

sess = res.get("session", {})
bpath = sess.get("backup_path")
print(f"   Status: {sess.get('status')}")
print(f"   Files: {sess.get('files_processed')}/{sess.get('files_total')}")
print(f"   Path: {bpath}")

# 3. Check ZIP
print("\n3. Checking ZIP file...")
if bpath:
    zfile = os.path.join(bpath, "backup.zip")
    if os.path.exists(zfile):
        print(f"   EXISTS: {zfile} ({os.path.getsize(zfile)} bytes)")
        import zipfile

        try:
            zf = zipfile.ZipFile(zfile, "r")
            print(f"   VALID ZIP with {len(zf.namelist())} files")
            zf.close()
        except Exception as e:
            print(f"   INVALID ZIP: {e}")
    else:
        print(f"   NOT FOUND: {zfile}")

# 4. Restore
print("\n4. Restoring...")
rdest = "D:\\RESTORE_FINAL_TEST"
r = requests.post(
    f"{BASE}/api/restore/files",
    headers=HDRS,
    json={
        "backup_path": bpath,
        "files": [],
        "destination": rdest,
        "restore_original": False,
    },
)
res = r.json()
if res.get("success"):
    rsess = res.get("session", {})
    print(f"   Status: {rsess.get('status')}")
    print(f"   Files: {rsess.get('files_restored', 0)}")
else:
    print(f"   ERROR: {res.get('error')}")

# 5. Verify
print("\n5. Verifying restore...")
if os.path.exists(rdest):
    files = []
    for root, d, fs in os.walk(rdest):
        for f in fs:
            files.append(os.path.join(root, f))
    print(f"   FOUND {len(files)} files in {rdest}")
    for f in files[:3]:
        print(f"     - {f}")
else:
    print(f"   NOT FOUND: {rdest}")

print("\n=== TEST COMPLETE ===")
