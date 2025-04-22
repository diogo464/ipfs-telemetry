import os
import logging
import nats
import nats.aio.client

from db.connection import DbConnectionInfo

logger = logging.getLogger(__name__)

NATS_ENDPOINT = os.environ.get("NATS_ENDPOINT", "nats://localhost:4222")

POSTGRES_HOST = os.environ.get("POSTGRES_HOST", "localhost")
POSTGRES_USER = os.environ.get("POSTGRES_USER", "postgres")
POSTGRES_PASSWORD = os.environ.get("POSTGRES_PASSWORD", "postgres")
POSTGRES_DATABASE = os.environ.get("POSTGRES_DB", "postgres")

# VictoriaMetrics
VM_ENDPOINT = os.environ.get("VM_ENDPOINT", "http://localhost:8428")

BACKEND_PREFIX = os.environ.get("BACKEND_PREFIX", "/api/v1")


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

