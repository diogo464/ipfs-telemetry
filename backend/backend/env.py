import os
import logging
import redis
import nats
import nats.aio.client

from minio import Minio

from backend.db.connection import DbConnectionInfo

logger = logging.getLogger(__name__)

NATS_ENDPOINT = os.environ.get("NATS_ENDPOINT", "nats://nats:4222")

# Not in use right now
S3_ENDPOINT = os.environ.get("S3_ENDPOINT", "minio:9000")
S3_USE_SSL = os.environ.get("S3_USE_SSL", "false") == "true"
S3_ACCESS_KEY = os.environ.get("S3_ACCESS_KEY", "minio")
S3_SECRET_KEY = os.environ.get("S3_SECRET_KEY", "minio123")
S3_BUCKET_TELEMETRY = os.environ.get("S3_BUCKET_TELEMETRY", "telemetry")

POSTGRES_HOST = os.environ.get("POSTGRES_HOST", "localhost")
POSTGRES_USER = os.environ.get("POSTGRES_USER", "postgres")
POSTGRES_PASSWORD = os.environ.get("POSTGRES_PASSWORD", "postgres")
POSTGRES_DATABASE = os.environ.get("POSTGRES_DB", "postgres")

# Not in use right now
REDIS_HOST = os.environ.get("REDIS_HOST", "redis")
REDIS_PORT = int(os.environ.get("REDIS_PORT", "6379"))
REDIS_PASSWORD = os.environ.get("REDIS_PASSWORD", "redis")

# VictoriaMetrics
VM_ENDPOINT = os.environ.get("VM_ENDPOINT", "http://localhost:8428")

BACKEND_PREFIX = os.environ.get("BACKEND_PREFIX", "/api/v1")


def create_minio_client() -> Minio:
    """
    Create a Minio client.
    """
    logger.info(f"connecting to S3 at {S3_ENDPOINT}")
    return Minio(
        S3_ENDPOINT, access_key=S3_ACCESS_KEY, secret_key=S3_SECRET_KEY, secure=False
    )


def create_redis_client() -> redis.Redis:
    """
    Create a Redis client.
    """
    logger.info(f"connecting to Redis at {REDIS_HOST}:{REDIS_PORT}")
    return redis.Redis(host=REDIS_HOST, port=REDIS_PORT, password=REDIS_PASSWORD)


async def create_nats_client() -> nats.aio.client.Client:
    """
    Create a NATS client.
    """
    logger.info(f"connecting to NATS at {NATS_ENDPOINT}")
    return await nats.connect(NATS_ENDPOINT)


def create_db_conn_info() -> DbConnectionInfo:
    username = POSTGRES_USER
    password = POSTGRES_PASSWORD
    database = POSTGRES_DATABASE
    host = POSTGRES_HOST
    return DbConnectionInfo(username, password, database, host)


def setup_logging():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
