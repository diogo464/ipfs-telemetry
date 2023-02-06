import logging
import datetime
import dataclasses
import json

from typing import Any, Sequence
from sqlalchemy.dialects.postgresql import insert
from sqlalchemy.orm import Session as SqlSession

from backend.monitor import Export

from . import Session, Property, Event, Metric, Histogram
from .model import Base

logger = logging.getLogger(__name__)


@dataclasses.dataclass
class FromRawResults:
    sessions: list[Session]
    properties: list[Property]
    events: list[Event]
    metrics: list[Metric]
    histograms: list[Histogram]

    def bulk_save(self, session: SqlSession) -> None:
        logger.info(
            f"Saving {len(self.sessions)} sessions, {len(self.properties)} properties, {len(self.events)} events, {len(self.metrics)} metrics, {len(self.histograms)} histograms"
        )
        _store_db_objects(session, self.sessions)
        _store_db_objects(session, self.properties)
        _store_db_objects(session, self.events)
        _store_db_objects(session, self.metrics)
        _store_db_objects(session, self.histograms)


def convert_export_object(export: Export) -> FromRawResults:
    db_sessions = []
    db_properties = []
    db_events = []
    db_metrics = []
    db_histograms = []

    db_sessions.append(_convert_export_session(export))
    db_properties.extend(_convert_export_properties(export))
    db_events.extend(_convert_export_events(export))
    metrics, histograms = _convert_export_metrics(export)
    db_metrics.extend(metrics)
    db_histograms.extend(histograms)

    return FromRawResults(
        sessions=db_sessions,
        properties=db_properties,
        events=db_events,
        metrics=db_metrics,
        histograms=db_histograms,
    )


def _store_db_objects(session: SqlSession, db_objects: Sequence[Base]):
    sessions = [ss for ss in db_objects if isinstance(ss, Session)]
    others = [obj for obj in db_objects if not isinstance(obj, Session)]
    for ss in sessions:
        ins = insert(Session).values(
            {
                "session": ss.session,
                "peer": ss.peer,
                "first_seen": ss.first_seen,
                "last_seen": ss.last_seen,
            }
        )
        ins = ins.on_conflict_do_update(
            constraint=Session.__table__.primary_key,
            set_={"last_seen": ins.excluded.last_seen},
            where=Session.last_seen < ins.excluded.last_seen,
        )
        session.execute(ins)
    session.bulk_save_objects(others)


def _convert_export_session(export: Export) -> Session:
    return Session(
        session=export.session,
        peer=export.peer,
        first_seen=export.observed_at,
        last_seen=export.observed_at,
    )


def _convert_export_properties(export: Export) -> list[Property]:
    db_props = []
    for property in export.properties:
        db_props.append(
            Property(
                session=export.session,
                peer=export.peer,
                scope=property.scope.name,
                version=property.scope.version,
                name=property.name,
                value=str(property.value),
            )
        )
    return db_props


def _convert_export_events(export: Export) -> list[Event]:
    db_events = []
    for event_export in export.events:
        descriptor = event_export.descriptor
        for event in event_export.events:
            try:
                event_data = event.decode().decode("utf-8")
                _ = json.loads(event_data)  # Check that it is valid JSON

                db_events.append(
                    Event(
                        session=export.session,
                        peer=export.peer,
                        scope=descriptor.scope.name,
                        version=descriptor.scope.version,
                        name=descriptor.name,
                        timestamp=event.timestamp,
                        value=event_data,
                    )
                )
            except Exception as e:
                logger.warn(f"Failed to decode event {descriptor}: {e}")
    return db_events


def _convert_export_metrics(
    export: Export,
) -> tuple[list[Metric], list[Histogram]]:
    db_metrics = []
    db_histograms = []

    resource_metrics = [m.decode_otlp() for m in export.metrics]
    peer = export.peer
    session = export.session
    for resmetrics in resource_metrics:
        for smetrics in resmetrics.scope_metrics:
            scope = smetrics.scope.name
            version = smetrics.scope.version
            for metric in smetrics.metrics:
                name = metric.name
                _ = metric.unit  # TODO: Store unit

                if metric.HasField("gauge"):
                    for dp in metric.gauge.data_points:
                        attributes = _convert_otel_attributes(dp.attributes)
                        timestamp = _convert_otel_timestamp(dp.time_unix_nano)
                        value = _convert_otel_numberdatapoint(dp)
                        db_metrics.append(
                            Metric(
                                session=session,
                                peer=peer,
                                scope=scope,
                                version=version,
                                name=name,
                                attributes=attributes,
                                timestamp=timestamp,
                                value=value,
                            )
                        )
                elif metric.HasField("sum"):
                    for dp in metric.sum.data_points:
                        attributes = _convert_otel_attributes(dp.attributes)
                        timestamp = _convert_otel_timestamp(dp.time_unix_nano)
                        value = _convert_otel_numberdatapoint(dp)
                        db_metrics.append(
                            Metric(
                                session=session,
                                peer=peer,
                                scope=scope,
                                version=version,
                                name=name,
                                attributes=attributes,
                                timestamp=timestamp,
                                value=value,
                            )
                        )
                elif metric.HasField("histogram"):
                    for dp in metric.histogram.data_points:
                        attributes = _convert_otel_attributes(dp.attributes)
                        timestamp = _convert_otel_timestamp(dp.time_unix_nano)
                        count = dp.count
                        sum = dp.sum
                        min = dp.min
                        max = dp.max
                        counts = [x for x in dp.bucket_counts]
                        bounds = [x for x in dp.explicit_bounds]
                        db_histograms.append(
                            Histogram(
                                session=session,
                                peer=peer,
                                scope=scope,
                                version=version,
                                name=name,
                                attributes=attributes,
                                timestamp=timestamp,
                                count=count,
                                sum=sum,
                                min=min,
                                max=max,
                                counts=counts,
                                bounds=bounds,
                            )
                        )
                else:
                    logger.warn("Unhandled metric type")

    return (db_metrics, db_histograms)


def _convert_otel_attributes(attrs) -> dict[str, Any]:
    attributes = {}
    for attr in attrs:
        key = attr.key
        value = None
        if attr.value.HasField("string_value"):
            value = attr.value.string_value
        elif attr.value.HasField("int_value"):
            value = attr.value.int_value
        elif attr.value.HasField("double_value"):
            value = attr.value.double_value
        else:
            logger.warn("Unhandled attribute type")
        if value is not None:
            attributes[key] = value
    return attributes


def _convert_otel_timestamp(time_unix_nano: int) -> datetime.datetime:
    return datetime.datetime.fromtimestamp(time_unix_nano / 1e9)


def _convert_otel_numberdatapoint(dp) -> float:
    if dp.HasField("as_int"):
        return float(dp.as_int)
    else:
        return dp.as_double
