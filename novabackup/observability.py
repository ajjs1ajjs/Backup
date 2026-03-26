import logging
import time
from functools import wraps
from typing import Callable

from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Response

# Logger setup
logger = logging.getLogger("novabackup")
logger.setLevel(logging.INFO)
# In production, you would configure handlers and formatters, but for now we just use basic config if not configured.
if not logger.handlers:
    handler = logging.StreamHandler()
    formatter = logging.Formatter(
        "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)

# Prometheus metrics
REQUEST_COUNT = Counter(
    "novabackup_http_requests_total",
    "Total HTTP requests",
    ["method", "endpoint", "http_status"],
)
REQUEST_LATENCY = Histogram(
    "novabackup_http_request_duration_seconds",
    "HTTP request latency in seconds",
    ["method", "endpoint"],
)

BACKUPS_CREATED = Counter(
    "novabackup_backups_created_total",
    "Total backups created",
    ["destination_type", "provider"],
)
BACKUPS_RESTORED = Counter(
    "novabackup_backups_restored_total", "Total backups restored", ["destination_type"]
)


# Decorator to track requests
def track_requests(method: str, endpoint: str):
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            try:
                response = await func(*args, **kwargs)
                # If response is a Response object, we can get status code; otherwise assume 200
                status_code = getattr(response, "status_code", 200)
                REQUEST_COUNT.labels(
                    method=method, endpoint=endpoint, http_status=status_code
                ).inc()
                REQUEST_LATENCY.labels(method=method, endpoint=endpoint).observe(
                    time.time() - start_time
                )
                return response
            except Exception as e:
                REQUEST_COUNT.labels(
                    method=method, endpoint=endpoint, http_status=500
                ).inc()
                REQUEST_LATENCY.labels(method=method, endpoint=endpoint).observe(
                    time.time() - start_time
                )
                raise

        return wrapper

    return decorator


def metrics_response() -> Response:
    """Returns Prometheus metrics."""
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


# Audit logging helper
def audit_log(action: str, user: str, details: str = ""):
    """Logs an audit event."""
    logger.info(f"AUDIT: action={action} user={user} {details}")
