#!/usr/bin/env python3
# NovaBackup v6.0 - Modern GUI Interface
# Powered by Flask + React

from flask import Flask, render_template, jsonify, request, send_from_directory
from flask_cors import CORS
import os
import sys
import json
import subprocess
import threading
import time
from datetime import datetime
import webbrowser

app = Flask(__name__)
CORS(app)

# Global state
backup_jobs = []
system_status = {
    "backup_server": True,
    "proxy_server": True,
    "repository_server": True,
    "deduplication": True,
    "encryption": False,
    "agent_running": False,
    "current_backup": None,
    "progress": 0
}

@app.route('/')
def index():
    return render_template('veeam-style.html')

@app.route('/api/status')
def get_status():
    return jsonify(system_status)

@app.route('/api/jobs')
def get_jobs():
    return jsonify(backup_jobs)

@app.route('/api/jobs', methods=['POST'])
def create_job():
    data = request.json
    new_job = {
        "id": len(backup_jobs) + 1,
        "name": data.get('name', ''),
        "type": data.get('type', 'Files'),
        "source": data.get('source', ''),
        "destination": data.get('destination', ''),
        "schedule": data.get('schedule', 'Daily'),
        "status": "Active",
        "last_run": "Never",
        "next_run": "Next scheduled run",
        "enabled": data.get('enabled', True)
    }
    backup_jobs.append(new_job)
    return jsonify({"success": True, "job": new_job})

@app.route('/api/jobs/<int:job_id>', methods=['DELETE'])
def delete_job(job_id):
    global backup_jobs
    backup_jobs = [job for job in backup_jobs if job['id'] != job_id]
    return jsonify({"success": True})

@app.route('/api/backup/start', methods=['POST'])
def start_backup():
    job_id = request.json.get('job_id')
    job = next((job for job in backup_jobs if job['id'] == job_id), None)
    
    if job:
        system_status['current_backup'] = job['name']
        system_status['progress'] = 0
        system_status['agent_running'] = True
        
        # Start backup in background thread
        threading.Thread(target=run_backup_simulation, args=(job,)).start()
        
        return jsonify({"success": True, "message": f"Started backup: {job['name']}"})
    
    return jsonify({"success": False, "message": "Job not found"})

@app.route('/api/agent/start', methods=['POST'])
def start_agent():
    system_status['agent_running'] = True
    return jsonify({"success": True, "message": "Background agent started"})

@app.route('/api/agent/stop', methods=['POST'])
def stop_agent():
    system_status['agent_running'] = False
    system_status['current_backup'] = None
    system_status['progress'] = 0
    return jsonify({"success": True, "message": "Background agent stopped"})

@app.route('/api/settings', methods=['GET', 'POST'])
def settings():
    if request.method == 'GET':
        return jsonify({
            "compression": "Optimal",
            "encryption": system_status['encryption'],
            "deduplication": system_status['deduplication'],
            "retention_days": 30,
            "max_concurrent_jobs": 4,
            "repository_path": "C:\\NovaBackup\\repository"
        })
    else:
        data = request.json
        system_status['encryption'] = data.get('encryption', False)
        system_status['deduplication'] = data.get('deduplication', True)
        return jsonify({"success": True, "message": "Settings updated"})

def run_backup_simulation(job):
    """Simulate backup pipeline"""
    stages = [
        "Job Scheduler & Init",
        "Snapshot Creation", 
        "Application Consistency",
        "Change Block Tracking",
        "Data Read (Proxy Stage)",
        "Compression",
        "Deduplication",
        "Encryption",
        "Transport & Storage Write",
        "Metadata & Indexing"
    ]
    
    for i, stage in enumerate(stages):
        system_status['progress'] = int((i + 1) / len(stages) * 100)
        time.sleep(1)  # Simulate processing time
    
    # Update job status
    job['last_run'] = datetime.now().strftime("%Y-%m-%d %H:%M")
    job['status'] = "Completed"
    
    system_status['current_backup'] = None
    system_status['progress'] = 0

def initialize_sample_jobs():
    """Initialize sample backup jobs"""
    global backup_jobs
    backup_jobs = [
        {
            "id": 1,
            "name": "Daily Documents Backup",
            "type": "Files",
            "status": "Active",
            "last_run": "2026-03-11 02:00",
            "next_run": "2026-03-12 02:00",
            "schedule": "Daily 2AM",
            "source": "C:\\Users\\Documents",
            "destination": "D:\\NovaBackups\\Documents",
            "enabled": True
        },
        {
            "id": 2,
            "name": "Weekly System Backup",
            "type": "System",
            "status": "Active",
            "last_run": "2026-03-08 22:00",
            "next_run": "2026-03-15 22:00",
            "schedule": "Weekly Sun 10PM",
            "source": "C:\\Windows",
            "destination": "D:\\NovaBackups\\System",
            "enabled": True
        },
        {
            "id": 3,
            "name": "Database Backup",
            "type": "SQL Server",
            "status": "Active",
            "last_run": "2026-03-11 01:00",
            "next_run": "2026-03-12 01:00",
            "schedule": "Daily 1AM",
            "source": "SQL Server Instance",
            "destination": "D:\\NovaBackups\\Database",
            "enabled": True
        }
    ]

if __name__ == '__main__':
    initialize_sample_jobs()
    
    # Create templates directory if it doesn't exist
    if not os.path.exists('templates'):
        os.makedirs('templates')
    
    # Create static directory if it doesn't exist
    if not os.path.exists('static'):
        os.makedirs('static')
    
    print("NovaBackup v6.0 GUI Server Starting...")
    print("Opening browser...")
    
    # Open browser after a short delay
    def open_browser():
        time.sleep(2)
        webbrowser.open('http://localhost:5000')
    
    threading.Thread(target=open_browser).start()
    
    # Start Flask server
    app.run(host='0.0.0.0', port=5000, debug=False)
