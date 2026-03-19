import json
import time

import requests

BASE = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"
HDRS = {"Authorization": TOKEN, "Content-Type": "application/json"}

print("Creating and running backup job...\n")

# Create job
r = requests.post(
    f"{BASE}/api/jobs",
    headers=HDRS,
    json={
        "name": "Debug Test",
        "sources": ["D:\\DOCUMENTS\\MY_SOFT\\KeepActive.exe"],
        "destination": "D:\\Backup\\DebugTest",
        "type": "file",
        "compression": True,
        "enabled": True,
    },
)
job = r.json().get("job", {})
jid = job.get("id")
print(f"Job created: {jid}\n")

# Run backup
print("Running backup (check server logs for details)...")
time.sleep(1)
r = requests.post(f"{BASE}/api/jobs/{jid}/run", headers=HDRS)
res = r.json()

if res.get("success"):
    sess = res.get("session", {})
    print(f"\nResult:")
    print(f"  Status: {sess.get('status')}")
    print(f"  Files: {sess.get('files_processed')}/{sess.get('files_total')}")
    print(
        f"  Bytes: {sess.get('bytes_read')} read, {sess.get('bytes_written')} written"
    )
    print(f"  Path: {sess.get('backup_path')}")
else:
    print(f"\nError: {res.get('error')}")
