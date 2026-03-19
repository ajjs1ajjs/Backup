import time

import requests

BASE = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"
HDRS = {"Authorization": TOKEN}

print("=" * 60)
print("TESTING DELETE FIX")
print("=" * 60)

# Get jobs
r = requests.get(f"{BASE}/api/jobs", headers=HDRS)
jobs = r.json().get("jobs", [])

print(f"\nFound {len(jobs)} jobs")

if not jobs:
    print("No jobs to test with!")
    exit(1)

# Try to delete first job
job_id = jobs[0]["id"]
job_name = jobs[0]["name"]
print(f"\nAttempting to delete: {job_name} ({job_id})")

r = requests.delete(f"{BASE}/api/jobs/{job_id}", headers=HDRS)

print(f"Response Status: {r.status_code}")
print(f"Response Body: {r.text}")

if r.status_code == 200:
    print(f"\nSUCCESS! Deleted '{job_name}'")
    print("\nThe fix is working - server was restarted with new code.")
elif r.status_code == 500:
    print(f"\nFAILED with 500 error")
    print("Server is still running OLD code - needs restart!")
    print("\nTo fix this:")
    print("1. Close the terminal where nova-backup.exe is running (Ctrl+C)")
    print("2. Run: D:\\WORK_CODE\\Backup\\nova-backup.exe server")
    print("3. Try again")
else:
    print(f"\nUnexpected status: {r.status_code}")

print("=" * 60)
