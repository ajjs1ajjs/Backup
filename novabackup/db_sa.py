from sqlalchemy import create_engine, Column, String, Text
from sqlalchemy.orm import declarative_base, sessionmaker
from typing import List, Dict, Any
import uuid
import datetime

Base = declarative_base()


class BackupSA(Base):
    __tablename__ = "backups"
    backup_id = Column(String, primary_key=True)
    vm_id = Column(String)
    dest = Column(String)
    type = Column(String)
    provider = Column(String)
    status = Column(String)
    created_at = Column(String)
    snapshot_id = Column(String)
    snapshot_provider = Column(String)
    snapshot_type = Column(String)
    snapshot_name = Column(String)


class RestoreSA(Base):
    __tablename__ = "restores"
    restore_id = Column(String, primary_key=True)
    backup_id = Column(String)
    vm_id = Column(String)
    dest = Column(String)
    status = Column(String)
    started_at = Column(String)
    finished_at = Column(String)


class SA_DBManager:
    def __init__(self, database_url: str):
        self.engine = create_engine(database_url, future=True, echo=False)
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine, future=True)

    def create_backup(
        self, vm_id: str, dest: str, backup_type: str, snapshot_name: str = None
    ) -> Dict[str, Any]:
        import datetime, uuid

        backup_id = str(uuid.uuid4())
        snapshot_id = str(uuid.uuid4())
        now = datetime.datetime.utcnow().isoformat() + "Z"
        provider = "DB"
        with self.Session() as session:
            b = BackupSA(
                backup_id=backup_id,
                vm_id=vm_id,
                dest=dest,
                type=backup_type,
                provider=provider,
                status="created",
                created_at=now,
                snapshot_id=snapshot_id,
                snapshot_provider=provider,
                snapshot_type=backup_type,
                snapshot_name=snapshot_name,
            )
            session.add(b)
            session.commit()
        snapshot = {
            "snapshot_id": snapshot_id,
            "provider": provider,
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "name": snapshot_name,
            "status": "created",
        }
        return {
            "backup_id": backup_id,
            "vm_id": vm_id,
            "dest": dest,
            "type": backup_type,
            "snapshot": snapshot,
            "provider": provider,
            "status": "created",
            "created_at": now,
        }

    def list_backups(self) -> List[Dict[str, Any]]:
        with self.Session() as session:
            rows = session.query(BackupSA).all()
            res: List[Dict[str, Any]] = []
            for r in rows:
                snapshot = {
                    "snapshot_id": r.snapshot_id,
                    "provider": r.snapshot_provider,
                    "vm_id": r.vm_id,
                    "dest": r.dest,
                    "type": r.snapshot_type,
                    "name": r.snapshot_name,
                    "status": r.status,
                }
                res.append(
                    {
                        "backup_id": r.backup_id,
                        "vm_id": r.vm_id,
                        "dest": r.dest,
                        "type": r.type,
                        "snapshot": snapshot,
                        "provider": r.provider,
                        "status": r.status,
                        "created_at": r.created_at,
                    }
                )
            return res

    def restore_backup(self, backup_id: str, dest: str) -> Dict[str, Any]:
        import datetime

        with self.Session() as session:
            b = session.query(BackupSA).filter_by(backup_id=backup_id).first()
            if not b:
                raise KeyError(backup_id)
            now = datetime.datetime.utcnow().isoformat() + "Z"
            restore_id = str(uuid.uuid4())
            r = RestoreSA(
                restore_id=restore_id,
                backup_id=backup_id,
                vm_id=b.vm_id,
                dest=dest,
                status="restored",
                started_at=now,
                finished_at=now,
            )
            session.add(r)
            b.status = "restored"
            session.commit()
        return {
            "backup_id": backup_id,
            "vm_id": b.vm_id,
            "dest": dest,
            "status": "restored",
            "restored_at": now,
        }
