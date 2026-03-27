"""
NovaBackup Backup Scheduler

Supports:
- Cron-like scheduling
- Interval-based scheduling
- One-time scheduled backups
- Persistent schedule storage
"""

import asyncio
import json
import logging
from typing import Dict, Any, Optional, List, Callable
from datetime import datetime, timedelta
from dataclasses import dataclass, asdict
from pathlib import Path
import os

from novabackup.backup import BackupManager
from novabackup.notifications import notify

logger = logging.getLogger("novabackup.scheduler")


@dataclass
class ScheduledJob:
    """Represents a scheduled backup job."""
    id: str
    name: str
    vm_id: str
    schedule_type: str  # cron, interval, once
    schedule_value: str  # cron expression or interval in seconds
    backup_type: str = "full"
    destination_type: str = "local"
    cloud_provider: Optional[str] = None
    cloud_region: Optional[str] = None
    cloud_dest: Optional[str] = None
    enabled: bool = True
    last_run: Optional[str] = None
    next_run: Optional[str] = None
    total_runs: int = 0
    failed_runs: int = 0
    created_at: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ScheduledJob':
        return cls(**data)


class BackupScheduler:
    """Manages scheduled backup jobs."""
    
    def __init__(self, schedule_file: Optional[str] = None):
        self.schedule_file = schedule_file or "schedules.json"
        self.jobs: Dict[str, ScheduledJob] = {}
        self.running = False
        self._task: Optional[asyncio.Task] = None
        self._backup_manager = BackupManager()
        self.load_schedules()
    
    def load_schedules(self):
        """Load schedules from file."""
        try:
            if os.path.exists(self.schedule_file):
                with open(self.schedule_file, 'r') as f:
                    data = json.load(f)
                    for job_data in data.get("jobs", []):
                        job = ScheduledJob.from_dict(job_data)
                        self.jobs[job.id] = job
                logger.info(f"Loaded {len(self.jobs)} scheduled jobs")
        except Exception as e:
            logger.error(f"Failed to load schedules: {e}")
    
    def save_schedules(self):
        """Save schedules to file."""
        try:
            data = {
                "jobs": [job.to_dict() for job in self.jobs.values()],
                "updated_at": datetime.utcnow().isoformat(),
            }
            with open(self.schedule_file, 'w') as f:
                json.dump(data, f, indent=2)
            logger.debug("Schedules saved")
        except Exception as e:
            logger.error(f"Failed to save schedules: {e}")
    
    def add_job(self, job: ScheduledJob) -> ScheduledJob:
        """Add a new scheduled job."""
        if not job.created_at:
            job.created_at = datetime.utcnow().isoformat()
        
        job.next_run = self._calculate_next_run(job)
        self.jobs[job.id] = job
        self.save_schedules()
        logger.info(f"Added scheduled job: {job.id} - {job.name}")
        return job
    
    def remove_job(self, job_id: str) -> bool:
        """Remove a scheduled job."""
        if job_id in self.jobs:
            del self.jobs[job_id]
            self.save_schedules()
            logger.info(f"Removed scheduled job: {job_id}")
            return True
        return False
    
    def enable_job(self, job_id: str) -> bool:
        """Enable a scheduled job."""
        if job_id in self.jobs:
            self.jobs[job_id].enabled = True
            self.jobs[job_id].next_run = self._calculate_next_run(self.jobs[job_id])
            self.save_schedules()
            return True
        return False
    
    def disable_job(self, job_id: str) -> bool:
        """Disable a scheduled job."""
        if job_id in self.jobs:
            self.jobs[job_id].enabled = False
            self.jobs[job_id].next_run = None
            self.save_schedules()
            return True
        return False
    
    def get_job(self, job_id: str) -> Optional[ScheduledJob]:
        """Get a scheduled job by ID."""
        return self.jobs.get(job_id)
    
    def list_jobs(self, enabled_only: bool = False) -> List[ScheduledJob]:
        """List all scheduled jobs."""
        jobs = list(self.jobs.values())
        if enabled_only:
            jobs = [j for j in jobs if j.enabled]
        return sorted(jobs, key=lambda j: j.next_run or "")
    
    def _calculate_next_run(self, job: ScheduledJob) -> Optional[str]:
        """Calculate next run time for a job."""
        now = datetime.utcnow()
        
        if job.schedule_type == "once":
            try:
                run_time = datetime.fromisoformat(job.schedule_value)
                if run_time > now:
                    return run_time.isoformat()
            except:
                pass
            return None
        
        elif job.schedule_type == "interval":
            try:
                seconds = int(job.schedule_value)
                next_run = now + timedelta(seconds=seconds)
                return next_run.isoformat()
            except:
                pass
        
        elif job.schedule_type == "cron":
            # Simple cron implementation (minute hour day month weekday)
            # For production, use 'croniter' library
            next_run = self._parse_cron(job.schedule_value, now)
            if next_run:
                return next_run.isoformat()
        
        return None
    
    def _parse_cron(self, cron_expr: str, now: datetime) -> Optional[datetime]:
        """Parse simple cron expression and return next run time."""
        try:
            parts = cron_expr.split()
            if len(parts) != 5:
                return None
            
            minute, hour, day, month, weekday = parts
            
            # Calculate next run (simplified)
            next_run = now.replace(second=0, microsecond=0)
            
            # Add 1 minute minimum
            next_run += timedelta(minutes=1)
            
            # Set hour and minute if specified
            if hour != "*":
                next_run = next_run.replace(hour=int(hour))
            if minute != "*":
                next_run = next_run.replace(minute=int(minute))
            
            return next_run
        except:
            return None
    
    async def start(self):
        """Start the scheduler."""
        if self.running:
            return
        
        self.running = True
        self._task = asyncio.create_task(self._run_loop())
        logger.info("Backup scheduler started")
    
    async def stop(self):
        """Stop the scheduler."""
        self.running = False
        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        logger.info("Backup scheduler stopped")
    
    async def _run_loop(self):
        """Main scheduler loop."""
        while self.running:
            try:
                await self._check_and_run_jobs()
                await asyncio.sleep(10)  # Check every 10 seconds
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Scheduler error: {e}")
                await asyncio.sleep(30)
    
    async def _check_and_run_jobs(self):
        """Check and run due jobs."""
        now = datetime.utcnow()
        
        for job in self.jobs.values():
            if not job.enabled or not job.next_run:
                continue
            
            next_run = datetime.fromisoformat(job.next_run)
            if next_run <= now:
                await self._run_job(job)
    
    async def _run_job(self, job: ScheduledJob):
        """Execute a scheduled backup job."""
        logger.info(f"Running scheduled job: {job.id} - {job.name}")
        
        try:
            # Create backup
            backup_result = self._backup_manager.create_backup(
                vm_id=job.vm_id,
                dest=job.cloud_dest or "./backups",
                backup_type=job.backup_type,
                snapshot_name=f"{job.name}-{datetime.utcnow().strftime('%Y%m%d-%H%M%S')}",
                destination_type=job.destination_type,
                cloud_provider=job.cloud_provider,
                cloud_region=job.cloud_region,
                cloud_dest=job.cloud_dest,
            )
            
            # Update job stats
            job.last_run = datetime.utcnow().isoformat()
            job.next_run = self._calculate_next_run(job)
            job.total_runs += 1
            
            if backup_result.get("status") == "completed":
                logger.info(f"Scheduled backup completed: {job.id}")
                await notify(
                    f"✅ Scheduled backup completed\n\n"
                    f"Job: {job.name}\n"
                    f"VM: {job.vm_id}\n"
                    f"Backup ID: {backup_result.get('backup_id')}",
                    level="success"
                )
            else:
                job.failed_runs += 1
                logger.warning(f"Scheduled backup failed: {job.id}")
                await notify(
                    f"⚠️ Scheduled backup failed\n\n"
                    f"Job: {job.name}\n"
                    f"VM: {job.vm_id}\n"
                    f"Status: {backup_result.get('status')}",
                    level="warning"
                )
            
            self.save_schedules()
            
        except Exception as e:
            job.failed_runs += 1
            logger.error(f"Scheduled job error: {job.id} - {e}")
            await notify(
                f"❌ Scheduled job error\n\n"
                f"Job: {job.name}\n"
                f"VM: {job.vm_id}\n"
                f"Error: {str(e)}",
                level="error"
            )
            self.save_schedules()


# Global scheduler instance
_scheduler: Optional[BackupScheduler] = None


def get_scheduler() -> BackupScheduler:
    """Get or create global scheduler."""
    global _scheduler
    if _scheduler is None:
        _scheduler = BackupScheduler()
    return _scheduler


def create_scheduled_backup(
    name: str,
    vm_id: str,
    schedule_type: str = "interval",
    schedule_value: str = "3600",
    backup_type: str = "full",
    destination_type: str = "local",
    cloud_provider: Optional[str] = None,
    cloud_region: Optional[str] = None,
    cloud_dest: Optional[str] = None,
) -> ScheduledJob:
    """
    Create a new scheduled backup.
    
    Args:
        name: Job name
        vm_id: VM to backup
        schedule_type: 'cron', 'interval', or 'once'
        schedule_value: Cron expression, interval seconds, or ISO datetime
        backup_type: 'full', 'incremental', or 'differential'
        destination_type: 'local' or 'cloud'
        cloud_provider: AWS, Azure, or GCP (if cloud)
        cloud_region: Cloud region
        cloud_dest: Cloud destination path
    
    Returns:
        Created ScheduledJob
    """
    import uuid
    
    job = ScheduledJob(
        id=f"job-{uuid.uuid4().hex[:8]}",
        name=name,
        vm_id=vm_id,
        schedule_type=schedule_type,
        schedule_value=schedule_value,
        backup_type=backup_type,
        destination_type=destination_type,
        cloud_provider=cloud_provider,
        cloud_region=cloud_region,
        cloud_dest=cloud_dest,
        enabled=True,
    )
    
    scheduler = get_scheduler()
    return scheduler.add_job(job)
