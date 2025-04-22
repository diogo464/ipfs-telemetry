from __future__ import annotations

import os
import logging

from typing import Union
from dataclasses import dataclass

import sqlalchemy
import sqlalchemy.engine
import sqlalchemy.dialects.postgresql
from sqlalchemy.engine import Engine
from sqlalchemy import select, text
from sqlalchemy.orm import Session as SqlSession

from .model import Session

logger = logging.getLogger(__name__)


@dataclass
class DbConnectionInfo:
    username: str
    password: str
    database: str
    host: str


def _read_schema() -> str:
    schema_path = os.path.join(os.path.dirname(__file__), "telemetry.sql")
    with open(schema_path, "r") as f:
        return f.read()


def _execute_schema(engine: Engine):
    schema = _read_schema()
    with engine.connect() as conn:
        conn.execute(text(schema))
        conn.commit()


def create_engine(connection_info: DbConnectionInfo) -> sqlalchemy.engine.Engine:
    return sqlalchemy.create_engine(
        f"postgresql+psycopg://{connection_info.username}:{connection_info.password}@{connection_info.host}/{connection_info.database}"
    )


def create_session(engine: Engine) -> SqlSession:
    return SqlSession(engine)


def requires_setup_database(engine: Engine) -> bool:
    with engine.connect() as conn:
        return not engine.dialect.has_table(conn, Session.__tablename__)


def setup_database(engine: Engine):
    _execute_schema(engine)

