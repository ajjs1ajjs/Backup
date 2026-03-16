#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import subprocess
import sys
import time

import requests

BASE_URL = "http://localhost:8050"


def login():
    """Login and return token"""
    try:
        resp = requests.post(
            f"{BASE_URL}/api/auth/login",
            json={"username": "admin", "password": "admin123"},
            timeout=5,
        )
        if resp.status_code == 200:
            data = resp.json()
            return data.get("token")
    except Exception as e:
        print(f"Error: {e}")
    return None


def get_jobs(token):
    """Get all jobs"""
    headers = {"Authorization": token}
    try:
        resp = requests.get(f"{BASE_URL}/api/jobs", headers=headers, timeout=5)
        if resp.status_code == 200:
            return resp.json().get("jobs", [])
    except Exception as e:
        print(f"Error: {e}")
    return []


def delete_job(token, job_id):
    """Delete a job"""
    headers = {"Authorization": token}
    try:
        resp = requests.delete(
            f"{BASE_URL}/api/jobs/{job_id}", headers=headers, timeout=5
        )
        return resp.status_code == 200
    except:
        pass
    return False


def has_unicode_chars(text):
    """Check if text contains Unicode BOM or zero-width chars"""
    if not text:
        return False
    for char in ["\ufeff", "\u200b", "\u200c", "\u200d", "\u2060"]:
        if char in text:
            return True
    return False


def main():
    print("=" * 60)
    print("NovaBackup - Unicode Path Cleaner")
    print("=" * 60)

    # Step 1: Login
    print("\n[1/4] Logging in...")
    token = login()
    if not token:
        print("[ERROR] Failed to login. Is server running?")
        print("\nPlease run:")
        print("  cd D:\\WORK_CODE\\Backup")
        print("  .\\nova-backup.exe server")
        raw_input("Press Enter to exit...")
        sys.exit(1)
    print("[OK] Logged in successfully")

    # Step 2: Get jobs
    print("\n[2/4] Getting jobs...")
    jobs = get_jobs(token)
    print("Found {} jobs".format(len(jobs)))

    # Step 3: Delete corrupted jobs
    print("\n[3/4] Deleting corrupted jobs...")
    deleted = 0
    kept = 0

    for job in jobs:
        name = job.get("name", "Unknown")
        sources = job.get("sources", [])
        dest = job.get("destination", "")

        # Check if any path has Unicode chars
        is_corrupted = False
        for src in sources:
            if has_unicode_chars(src):
                is_corrupted = True
                break

        if has_unicode_chars(dest) or has_unicode_chars(name):
            is_corrupted = True

        if is_corrupted:
            print("  DEL: {}".format(name))
            if delete_job(token, job["id"]):
                deleted += 1
            else:
                print("      [WARN] Failed to delete")
        else:
            print("  KEEP: {}".format(name))
            kept += 1

    print("\nDeleted {} corrupted jobs, kept {} clean jobs".format(deleted, kept))

    # Step 4: Restart server
    print("\n[4/4] Restarting server...")
    print("  Stopping old server...")

    os.system("taskkill /F /IM nova-backup.exe >nul 2>&1")

    time.sleep(2)

    print("  Starting new server...")
    subprocess.Popen(
        ["nova-backup.exe", "server"],
        cwd=r"D:\WORK_CODE\Backup",
        creationflags=subprocess.CREATE_NEW_CONSOLE,
    )

    time.sleep(3)

    print("\n" + "=" * 60)
    print("DONE!")
    print("=" * 60)
    print("\nNext steps:")
    print("1. Wait 5 seconds for server to fully start")
    print("2. Open http://localhost:8050/quick-backup.html")
    print("3. Create NEW backup job with clean paths")
    print("   Example: D:/Documents/MySoft (use forward slashes)")
    print("4. Test the backup")
    print("\nWARNING: DO NOT reuse old corrupted jobs from dropdown!")
    print("=" * 60)

    raw_input("\nPress Enter to exit...")


if __name__ == "__main__":
    main()
