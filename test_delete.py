import requests

BASE = "http://localhost:8050"
TOKEN = "eba53c11ead2c28d0d55b3fdd80d6fb2d08385a50f5062d45f12c9661067d5bf"
HDRS = {"Authorization": TOKEN}

# Get jobs
r = requests.get(f"{BASE}/api/jobs", headers=HDRS)
jobs = r.json().get("jobs", [])

print(f"Found {len(jobs)} jobs")

# Try to delete first job
if jobs:
    job_id = jobs[0]["id"]
    job_name = jobs[0]["name"]
    print(f"\nTrying to delete: {job_name} ({job_id})")

    r = requests.delete(f"{BASE}/api/jobs/{job_id}", headers=HDRS)

    if r.status_code == 200:
        print(f"SUCCESS! Deleted {job_name}")
    else:
        print(f"ERROR {r.status_code}: {r.json().get('error', 'Unknown error')}")
else:
    print("No jobs to delete")
