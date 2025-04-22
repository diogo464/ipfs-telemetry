from __future__ import annotations

import uuid
import datetime

from typing import Any, Union

from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column
from sqlalchemy.dialects.postgresql import JSONB, ARRAY, REAL, INTEGER

# https://docs.sqlalchemy.org/en/20/faq/ormconfiguration.html#how-do-i-map-a-table-that-has-no-primary-key


class Base(DeclarativeBase):
    pass


class Session(Base):
    __tablename__ = "sessions"

    session: Mapped[uuid.UUID] = mapped_column(primary_key=True)
    peer: Mapped[str] = mapped_column(primary_key=True)
    first_seen: Mapped[datetime.datetime] = mapped_column()
    last_seen: Mapped[datetime.datetime] = mapped_column()

    def __repr__(self) -> str:
        return f"<Session(session={self.session}, peer={self.peer}, first_seen={self.first_seen}, last_seen={self.last_seen})>"


class Metric(Base):
    __tablename__ = "metrics"

    session: Mapped[uuid.UUID] = mapped_column(primary_key=True)
    peer: Mapped[str] = mapped_column()
    scope: Mapped[str] = mapped_column()
    version: Mapped[str] = mapped_column()
    name: Mapped[str] = mapped_column()
    attributes: Mapped[dict[str, Any]] = mapped_column(JSONB)
    timestamp: Mapped[datetime.datetime] = mapped_column()
    value: Mapped[float] = mapped_column()

    def __repr__(self) -> str:
        return f"<Metric(session={self.session}, peer={self.peer}, scope={self.scope}, version={self.version}, name={self.name}, attributes={self.attributes}, timestamp={self.timestamp}, value={self.value})>"


class Histogram(Base):
    __tablename__ = "histograms"

    session: Mapped[uuid.UUID] = mapped_column(primary_key=True)
    peer: Mapped[str] = mapped_column()
    scope: Mapped[str] = mapped_column()
    version: Mapped[str] = mapped_column()
    name: Mapped[str] = mapped_column()
    attributes: Mapped[dict[str, Any]] = mapped_column(JSONB)
    timestamp: Mapped[datetime.datetime] = mapped_column()
    count: Mapped[int] = mapped_column()
    sum: Mapped[float] = mapped_column()
    min: Mapped[float] = mapped_column()
    max: Mapped[float] = mapped_column()
    bounds: Mapped[list[float]] = mapped_column(ARRAY(REAL))
    counts: Mapped[list[int]] = mapped_column(ARRAY(INTEGER))

    def __repr__(self) -> str:
        return f"<Histogram(session={self.session}, peer={self.peer}, scope={self.scope}, version={self.version}, name={self.name}, attributes={self.attributes}, timestamp={self.timestamp}, count={self.count}, sum={self.sum}, min={self.min}, max={self.max}, bounds={self.bounds}, counts={self.counts})>"


class Property(Base):
    __tablename__ = "properties"

    session: Mapped[uuid.UUID] = mapped_column(primary_key=True)
    peer: Mapped[str] = mapped_column()
    scope: Mapped[str] = mapped_column()
    version: Mapped[str] = mapped_column()
    name: Mapped[str] = mapped_column()
    value: Mapped[str] = mapped_column()

    def __repr__(self) -> str:
        return f"<Property(session={self.session}, peer={self.peer}, scope={self.scope}, version={self.version}, name={self.name}, value={self.value})>"


class Event(Base):
    __tablename__ = "events"

    session: Mapped[uuid.UUID] = mapped_column(primary_key=True)
    peer: Mapped[str] = mapped_column()
    scope: Mapped[str] = mapped_column()
    version: Mapped[str] = mapped_column()
    name: Mapped[str] = mapped_column()
    timestamp: Mapped[datetime.datetime] = mapped_column()
    value: Mapped[str] = mapped_column(JSONB)

    def __repr__(self) -> str:
        return f"<Event(session={self.session}, peer={self.peer}, scope={self.scope}, version={self.version}, name={self.name}, timestamp={self.timestamp}, value={self.value})>"


class Discovery(Base):
    __tablename__ = "discovery"

    peer: Mapped[str] = mapped_column(primary_key=True)
    timestamp: Mapped[datetime.datetime] = mapped_column(primary_key=True)
    latitude: Mapped[float] = mapped_column()
    longitude: Mapped[float] = mapped_column()
    location: Mapped[str] = mapped_column()

    def __repr__(self) -> str:
        return f"<Discovery(peer={self.peer}, timestamp={self.timestamp}, latitude={self.latitude}, longitude={self.longitude}, location={self.location})>"


class Description(Base):
    __tablename__ = "descriptions"

    scope: Mapped[str] = mapped_column(primary_key=True)
    version: Mapped[str] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(primary_key=True)
    type: Mapped[str] = mapped_column(primary_key=True)
    description: Mapped[str] = mapped_column()
