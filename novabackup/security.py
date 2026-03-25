from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import os

import jwt  # PyJWT
from fastapi import HTTPException, Depends
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm

# Simple in‑memory store for MVP (replace with DB in production)
USERS = {
    "alice": {"password": "secret", "roles": ["admin"]},
    "bob": {"password": "secret", "roles": ["user"]},
}

SECRET_KEY = os.environ.get("NOVABACKUP_JWT_SECRET", "change-me")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60
REFRESH_TOKEN_EXPIRE_DAYS = 7

REFRESH_TOKENS: Dict[str, str] = {}

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/token")


def verify_password(plain: str, password: str) -> bool:
    return plain == password


def authenticate_user(username: str, password: str) -> Optional[Dict[str, Any]]:
    user = USERS.get(username)
    if not user:
        return None
    if not verify_password(password, user["password"]):
        return None
    return {"username": username, "roles": user["roles"]}


def create_access_token(
    data: Dict[str, Any], expires_delta: Optional[timedelta] = None
) -> str:
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire, "typ": "access"})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)


def create_refresh_token(
    data: Dict[str, Any], expires_delta: Optional[timedelta] = None
) -> str:
    to_encode = data.copy()
    if expires_delta is None:
        expires_delta = timedelta(days=REFRESH_TOKEN_EXPIRE_DAYS)
    expire = datetime.utcnow() + expires_delta
    to_encode.update({"exp": expire, "typ": "refresh"})
    token = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    username = data.get("sub") or data.get("username")
    if username:
        REFRESH_TOKENS[token] = username
    return token


def refresh_access_token(refresh_token: str) -> Dict[str, Any]:
    try:
        payload = jwt.decode(refresh_token, SECRET_KEY, algorithms=[ALGORITHM])
        username = payload.get("sub") or payload.get("username")
        roles = payload.get("roles", [])
        if not username:
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        new_access = create_access_token(
            {"sub": username, "roles": roles}, expires_delta=timedelta(minutes=60)
        )
        new_refresh = create_refresh_token({"sub": username, "roles": roles})
        REFRESH_TOKENS.pop(refresh_token, None)
        REFRESH_TOKENS[new_refresh] = username
        return {
            "access_token": new_access,
            "token_type": "bearer",
            "refresh_token": new_refresh,
        }
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Refresh token expired")
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid refresh token")


async def get_current_user(token: str = Depends(oauth2_scheme)) -> Dict[str, Any]:
    if not token:
        raise HTTPException(status_code=401, detail="Not authenticated")
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username = payload.get("sub") or payload.get("username")
        roles = payload.get("roles", [])
        if not username:
            raise HTTPException(status_code=401, detail="Invalid authentication")
        return {"username": username, "roles": roles}
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Token expired")
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid authentication")


def require_role(allowed: List[str]):
    async def _require(current_user: Dict[str, Any] = Depends(get_current_user)):
        roles = current_user.get("roles", [])
        if not any(r in roles for r in allowed):
            raise HTTPException(status_code=403, detail="Not enough privileges")
        return current_user

    return _require
