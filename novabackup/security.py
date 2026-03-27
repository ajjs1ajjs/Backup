from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List, Set
import os
import secrets
import json
import logging

import jwt  # PyJWT
from fastapi import HTTPException, Depends
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm

# Logger for audit events
audit_logger = logging.getLogger("novabackup.audit")
audit_logger.setLevel(logging.INFO)

# Simple in‑memory store for MVP (replace with DB in production)
USERS = {
    "alice": {
        "password": "secret",
        "roles": ["admin"],
        "scopes": ["read", "write", "delete", "manage_users"]
    },
    "bob": {
        "password": "secret",
        "roles": ["user"],
        "scopes": ["read", "write"]
    },
    # New test user with no privileges to validate RBAC boundaries in tests
    "charlie": {"password": "secret", "roles": [], "scopes": []},
    # Service account for automated tasks
    "service": {
        "password": "service-secret",
        "roles": ["service"],
        "scopes": ["read", "backup", "restore"]
    },
}

# JWT secret from environment variable or generate secure random fallback
# IMPORTANT: Set NOVABACKUP_JWT_SECRET environment variable in production!
SECRET_KEY = os.environ.get("NOVABACKUP_JWT_SECRET")
if not SECRET_KEY:
    # Generate a secure random key for development only
    SECRET_KEY = secrets.token_hex(32)
    print("⚠️  WARNING: Using auto-generated JWT secret. Set NOVABACKUP_JWT_SECRET in production!")

ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60
REFRESH_TOKEN_EXPIRE_DAYS = 7

# In-memory stores (replace with Redis/DB in production)
REFRESH_TOKENS: Dict[str, str] = {}  # token -> username
TOKEN_BLACKLIST: Set[str] = set()  # Blacklisted tokens (for logout)
AUDIT_LOGS: List[Dict[str, Any]] = []  # Recent audit logs

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/token")


def verify_password(plain: str, password: str) -> bool:
    return plain == password


def authenticate_user(username: str, password: str) -> Optional[Dict[str, Any]]:
    user = USERS.get(username)
    if not user:
        audit_log("login_failed", username, "user_not_found")
        return None
    if not verify_password(password, user["password"]):
        audit_log("login_failed", username, "invalid_password")
        return None
    audit_log("login_success", username, f"roles={user['roles']}")
    return {"username": username, "roles": user["roles"], "scopes": user.get("scopes", [])}


def create_access_token(
    data: Dict[str, Any], expires_delta: Optional[timedelta] = None, scopes: Optional[List[str]] = None
) -> str:
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({
        "exp": expire,
        "typ": "access",
        "iat": datetime.utcnow(),
        "jti": secrets.token_hex(16),  # Unique token ID
    })
    if scopes:
        to_encode["scopes"] = scopes
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)


def create_refresh_token(
    data: Dict[str, Any], expires_delta: Optional[timedelta] = None
) -> str:
    to_encode = data.copy()
    if expires_delta is None:
        expires_delta = timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS)
    expire = datetime.utcnow() + expires_delta
    to_encode.update({
        "exp": expire,
        "typ": "refresh",
        "iat": datetime.utcnow(),
        "jti": secrets.token_hex(16),
    })
    token = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    username = data.get("sub") or data.get("username")
    if username:
        REFRESH_TOKENS[token] = username
        audit_log("refresh_token_created", username, f"expires={expire.isoformat()}")
    return token


def refresh_access_token(refresh_token: str) -> Dict[str, Any]:
    try:
        if refresh_token in TOKEN_BLACKLIST:
            audit_log("token_refresh_blocked", "unknown", "blacklisted_token")
            raise HTTPException(status_code=401, detail="Token has been revoked")
        
        payload = jwt.decode(refresh_token, SECRET_KEY, algorithms=[ALGORITHM])
        username = payload.get("sub") or payload.get("username")
        roles = payload.get("roles", [])
        scopes = payload.get("scopes", [])
        
        if not username:
            audit_log("token_refresh_failed", "unknown", "invalid_token_payload")
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        
        # Verify token is in our store
        stored_username = REFRESH_TOKENS.get(refresh_token)
        if stored_username and stored_username != username:
            audit_log("token_refresh_failed", username, "token_mismatch")
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        
        new_access = create_access_token(
            {"sub": username, "roles": roles},
            expires_delta=timedelta(minutes=60),
            scopes=scopes
        )
        new_refresh = create_refresh_token({"sub": username, "roles": roles})
        
        # Revoke old token
        REFRESH_TOKENS.pop(refresh_token, None)
        TOKEN_BLACKLIST.add(refresh_token)
        
        audit_log("token_refreshed", username, "new_tokens_issued")
        return {
            "access_token": new_access,
            "token_type": "bearer",
            "refresh_token": new_refresh,
        }
    except jwt.ExpiredSignatureError:
        audit_log("token_refresh_expired", "unknown", "refresh_token_expired")
        raise HTTPException(status_code=401, detail="Refresh token expired")
    except Exception as e:
        if isinstance(e, HTTPException):
            raise
        audit_log("token_refresh_error", "unknown", str(e))
        raise HTTPException(status_code=401, detail="Invalid refresh token")


async def get_current_user(token: str = Depends(oauth2_scheme)) -> Dict[str, Any]:
    if not token:
        raise HTTPException(status_code=401, detail="Not authenticated")
    
    if token in TOKEN_BLACKLIST:
        raise HTTPException(status_code=401, detail="Token has been revoked")
    
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username = payload.get("sub") or payload.get("username")
        roles = payload.get("roles", [])
        scopes = payload.get("scopes", [])
        
        if not username:
            raise HTTPException(status_code=401, detail="Invalid authentication")
        return {
            "username": username,
            "roles": roles,
            "scopes": scopes,
            "token_id": payload.get("jti", "unknown")
        }
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Token expired")
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid authentication")


def require_role(allowed: List[str]):
    """Require user to have at least one of the allowed roles."""
    async def _require(current_user: Dict[str, Any] = Depends(get_current_user)):
        roles = current_user.get("roles", [])
        if not any(r in roles for r in allowed):
            audit_log(
                "access_denied_role",
                current_user.get("username", "unknown"),
                f"required={allowed}, has={roles}"
            )
            raise HTTPException(status_code=403, detail="Not enough privileges")
        return current_user

    return _require


def require_scope(required_scopes: List[str]):
    """Require user to have all of the required scopes."""
    async def _require_scope(current_user: Dict[str, Any] = Depends(get_current_user)):
        user_scopes = set(current_user.get("scopes", []))
        missing_scopes = set(required_scopes) - user_scopes
        
        if missing_scopes:
            audit_log(
                "access_denied_scope",
                current_user.get("username", "unknown"),
                f"required={required_scopes}, missing={list(missing_scopes)}"
            )
            raise HTTPException(
                status_code=403,
                detail=f"Missing required scopes: {list(missing_scopes)}"
            )
        return current_user

    return _require_scope


def revoke_token(token: str) -> bool:
    """Revoke a token (add to blacklist)."""
    if token in REFRESH_TOKENS:
        username = REFRESH_TOKENS.pop(token)
        audit_log("token_revoked", username, "refresh_token_revoked")
    
    TOKEN_BLACKLIST.add(token)
    audit_log("token_blacklisted", "unknown", f"token_id={token[:16]}...")
    return True


def audit_log(action: str, user: str, details: str = ""):
    """Log an audit event to both logger and in-memory store."""
    timestamp = datetime.utcnow().isoformat()
    log_entry = {
        "timestamp": timestamp,
        "action": action,
        "user": user,
        "details": details,
    }
    
    # Log to file via standard logger
    audit_logger.info(f"AUDIT: {json.dumps(log_entry)}")
    
    # Store in-memory (keep last 1000 entries)
    AUDIT_LOGS.append(log_entry)
    if len(AUDIT_LOGS) > 1000:
        AUDIT_LOGS.pop(0)


def get_audit_logs(limit: int = 100) -> List[Dict[str, Any]]:
    """Get recent audit logs."""
    return AUDIT_LOGS[-limit:]
