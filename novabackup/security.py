import os
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional

import jwt  # PyJWT
from fastapi import HTTPException, Depends
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm

# JWT settings (use env var in production)
SECRET_KEY = os.environ.get("NOVABACKUP_JWT_SECRET", "change-me")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60

# Simple in-memory user store (for MVP/testing). In production, replace with a real user DB
USERS = {
    "alice": {"password": "secret", "roles": ["admin"]},
    "bob": {"password": "secret", "roles": ["user"]},
}

# OAuth2 scheme for token retrieval
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
        expire = datetime.utcnow() + timedelta(minutes=15)
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)


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
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid authentication")


def require_role(allowed: List[str]):
    async def _require(current_user: Dict[str, Any] = Depends(get_current_user)):
        roles = current_user.get("roles", [])
        if not any(r in roles for r in allowed):
            raise HTTPException(status_code=403, detail="Not enough privileges")
        return current_user

    return _require
