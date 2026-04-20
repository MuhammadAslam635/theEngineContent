import os
import threading
from contextlib import contextmanager
from typing import Any, Dict, Iterator, Optional, Union

import psycopg2
from psycopg2.pool import ThreadedConnectionPool


_POOL: Optional[ThreadedConnectionPool] = None
_POOL_LOCK = threading.Lock()


def _truthy(value: Optional[str], default: bool = False) -> bool:
    if value is None:
        return default
    return value.strip().lower() in {"1", "true", "yes", "y", "on"}


AUDIT_LOGGING_ENABLED = _truthy(os.getenv("AUDIT_LOGGING_ENABLED"), default=True)


def _build_database_url_from_parts() -> Optional[str]:
    host = os.getenv("POSTGRES_HOST")
    port = os.getenv("POSTGRES_PORT", "5432")
    dbname = os.getenv("POSTGRES_DB")
    user = os.getenv("POSTGRES_USER")
    password = os.getenv("POSTGRES_PASSWORD")

    if not host or not dbname or not user or password is None:
        return None

    return f"postgresql://{user}:{password}@{host}:{port}/{dbname}"


def get_database_url() -> Optional[str]:
    return os.getenv("DATABASE_URL") or _build_database_url_from_parts()


def get_audit_logs_pool() -> ThreadedConnectionPool:
    global _POOL

    if _POOL is not None:
        return _POOL

    with _POOL_LOCK:
        if _POOL is not None:
            return _POOL

        database_url = get_database_url()
        if not database_url:
            raise RuntimeError(
                "Missing DATABASE_URL or POSTGRES_* environment variables required for audit logging."
            )

        minconn = int(os.getenv("AUDIT_DB_POOL_MIN", "1"))
        maxconn = int(os.getenv("AUDIT_DB_POOL_MAX", "5"))
        _POOL = ThreadedConnectionPool(minconn=minconn, maxconn=maxconn, dsn=database_url)
        return _POOL


@contextmanager
def _audit_conn() -> Iterator[Any]:
    pool = get_audit_logs_pool()
    conn = pool.getconn()
    try:
        yield conn
    finally:
        pool.putconn(conn)


def close_audit_logs_pool() -> None:
    global _POOL

    with _POOL_LOCK:
        if _POOL is None:
            return
        _POOL.closeall()
        _POOL = None


def write_audit_log(
    *,
    agent: str,
    user_id: Union[int, str],
    input_query: Optional[str] = None,
    agent_prompt: Optional[str] = None,
    input_tokens: int = 0,
    output_response: Optional[str] = None,
    output_usage_tokens: int = 0,
    error: Optional[str] = None,
    line_number: Optional[int] = None,
) -> Optional[int]:
    if not AUDIT_LOGGING_ENABLED:
        return None

    if user_id is None or str(user_id).strip() == "":
        raise ValueError("user_id is required")

    user_id_int = int(user_id)

    sql = """
        INSERT INTO audit_logs (
            agent,
            input_query,
            agent_prompt,
            input_tokens,
            output_response,
            output_usage_tokens,
            error,
            line_number,
            user_id
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        RETURNING id;
    """

    params = (
        agent,
        input_query,
        agent_prompt,
        int(input_tokens or 0),
        output_response,
        int(output_usage_tokens or 0),
        error,
        line_number,
        user_id_int,
    )

    with _audit_conn() as conn:
        with conn.cursor() as cur:
            cur.execute(sql, params)
            inserted_id_row = cur.fetchone()
        conn.commit()

    return int(inserted_id_row[0]) if inserted_id_row else None


def try_write_audit_log(**kwargs: Any) -> Optional[int]:
    try:
        return write_audit_log(**kwargs)
    except (psycopg2.Error, ValueError, TypeError, RuntimeError):
        return None


def normalize_audit_payload(payload: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "agent": payload.get("agent", ""),
        "user_id": payload.get("user_id", ""),
        "input_query": payload.get("input_query"),
        "agent_prompt": payload.get("agent_prompt"),
        "input_tokens": int(payload.get("input_tokens") or 0),
        "output_response": payload.get("output_response"),
        "output_usage_tokens": int(payload.get("output_usage_tokens") or 0),
        "error": payload.get("error"),
        "line_number": payload.get("line_number"),
    }
