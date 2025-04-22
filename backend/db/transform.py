import logging
import datetime
import dataclasses
import json

from typing import Any, Sequence
from sqlalchemy.dialects.postgresql import insert
from sqlalchemy.orm import Session as SqlSession

from export.monitor import MonitorExport

from . import Session, Property, Event, Metric, Histogram, Description
from .model import Base

logger = logging.getLogger(__name__)


@dataclasses.dataclass(eq=True, frozen=True)
class DescriptionData:
    scope: str
    version: str
    name: str
    description: str
    type: str


@dataclasses.dataclass
class FromRawResults:
    sessions: list[Session]
    properties: list[Property]
    events: list[Event]
    metrics: list[Metric]
    histograms: list[Histogram]
    descriptions: list[Description]

    def bulk_save(self, session: SqlSession) -> None:
        logger.info(
            f"Saving {len(self.sessions)} sessions, {len(self.properties)} properties, {len(self.events)} events, {len(self.metrics)} metrics, {len(self.histograms)} histograms, and {len(self.descriptions)} descriptions."
        )
        _store_db_objects(session, self.sessions)
        _store_db_objects(session, self.properties)
        _store_db_objects(session, self.events)
        _store_db_objects(session, self.metrics)
        _store_db_objects(session, self.histograms)
        _store_db_objects(session, self.descriptions)


def convert_export_object(export: MonitorExport) -> FromRawResults:
    db_sessions: list[Session] = []
    db_properties: list[Property] = []
    db_events: list[Event] = []
    db_metrics: list[Metric] = []
    db_histograms: list[Histogram] = []
    db_descriptions: list[Description] = []
    descriptions: set[DescriptionData] = set()

    db_sessions.append(_convert_export_session(export))
    db_properties.extend(_convert_export_properties(export, descriptions))
    db_events.extend(_convert_export_events(export, descriptions))
    metrics, histograms = _convert_export_metrics(export, descriptions)
    db_metrics.extend(metrics)
    db_histograms.extend(histograms)
    for desc in descriptions:
        db_descriptions.append(
            Description(
                scope=desc.scope,
                version=desc.version,
                name=desc.name,
                description=desc.description,
                type=desc.type,
            )
        )

    return FromRawResults(
        sessions=db_sessions,
        properties=db_properties,
        events=db_events,
        metrics=db_metrics,
        histograms=db_histograms,
        descriptions=db_descriptions,
    )


def _store_db_objects(session: SqlSession, db_objects: Sequence[Base]):
    sessions = [ss for ss in db_objects if isinstance(ss, Session)]
    others = [obj for obj in db_objects if not isinstance(obj, Session)]

    descriptions = [desc for desc in others if isinstance(desc, Description)]
    others = [obj for obj in others if not isinstance(obj, Description)]

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

    logger.info(f"saving {len(descriptions)} descriptions")
    for desc in descriptions:
        ins = insert(Description).values(
            {
                "scope": desc.scope,
                "version": desc.version,
                "name": desc.name,
                "description": desc.description,
                "type": desc.type,
            }
        )
        ins = ins.on_conflict_do_nothing(
            constraint=Description.__table__.primary_key,
        )
        session.execute(ins)
    session.bulk_save_objects(others)


def _convert_export_session(export: MonitorExport) -> Session:
    return Session(
        session=export.session,
        peer=export.peer,
        first_seen=export.observed_at,
        last_seen=export.observed_at,
    )


def _convert_export_properties(
    export: MonitorExport, descriptions: set[DescriptionData]
) -> list[Property]:
    db_props = []
    for property in export.properties:
        description = DescriptionData(
            scope=property.scope.name,
            version=property.scope.version,
            name=property.name,
            description=property.description,
            type="property",
        )
        prop = Property(
            session=export.session,
            peer=export.peer,
            scope=property.scope.name,
            version=property.scope.version,
            name=property.name,
            value=str(property.value),
        )
        db_props.append(prop)
        descriptions.add(description)
    return db_props


def _convert_export_events(
    export: MonitorExport, descriptions: set[DescriptionData]
) -> list[Event]:
    db_events = []
    for event_export in export.events:
        descriptor = event_export.descriptor
        for event in event_export.events:
            try:
                event_data = event.decode().decode("utf-8")
                _ = json.loads(event_data)  # Check that it is valid JSON

                description = DescriptionData(
                    scope=descriptor.scope.name,
                    version=descriptor.scope.version,
                    name=descriptor.name,
                    description=descriptor.description,
                    type="event",
                )
                ev = Event(
                    session=export.session,
                    peer=export.peer,
                    scope=descriptor.scope.name,
                    version=descriptor.scope.version,
                    name=descriptor.name,
                    timestamp=event.timestamp,
                    value=event_data,
                )
                db_events.append(ev)
                descriptions.add(description)
            except Exception as e:
                logger.warning(f"Failed to decode event {descriptor}: {e}")
    return db_events


def _convert_export_metrics(
    export: MonitorExport, descriptions: set[DescriptionData]
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
                        description = DescriptionData(
                            scope=scope,
                            version=version,
                            name=name,
                            description=metric.description,
                            type="metric",
                        )
                        descriptions.add(description)
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
                        description = DescriptionData(
                            scope=scope,
                            version=version,
                            name=name,
                            description=metric.description,
                            type="metric",
                        )
                        descriptions.add(description)
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
                        description = DescriptionData(
                            scope=scope,
                            version=version,
                            name=name,
                            description=metric.description,
                            type="histogram",
                        )
                        descriptions.add(description)
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
                    logger.warning("Unhandled metric type")

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
            logger.warning("Unhandled attribute type")
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

