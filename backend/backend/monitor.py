import base64
import datetime
import uuid

from opentelemetry.proto.metrics.v1.metrics_pb2 import ResourceMetrics
from typing import Optional
from pydantic import BaseModel, Field


class Scope(BaseModel):
    name: str = Field(description="Scope name", alias="Name")
    version: str = Field(description="Scope version", alias="Version")
    schema_url: str = Field(description="Scope schema URL", alias="SchemaURL")


class EventDescriptor(BaseModel):
    scope: Scope = Field(description="Scope of the event")
    name: str = Field(description="Name of the event")
    description: str = Field(description="Description of the event")


class Event(BaseModel):
    timestamp: datetime.datetime = Field(description="Timestamp of the event")
    data: str = Field(description="Data of the event")

    def decode(self) -> bytes:
        return base64.b64decode(self.data)


class ExportBandwidth(BaseModel):
    upload_rate: int = Field(description="Upload rate in bytes per second")
    download_rate: int = Field(description="Download rate in bytes per second")


class ExportEvents(BaseModel):
    descriptor: EventDescriptor = Field(description="Event descriptor")
    events: list[Event] = Field(description="Events")


class ExportMetrics(BaseModel):
    otlp: bytes = Field(description="OTLP metrics")

    def decode_otlp(self) -> ResourceMetrics:
        return ResourceMetrics.FromString(base64.b64decode(self.otlp))


class ExportProperty(BaseModel):
    scope: Scope = Field(description="Scope of the property")
    name: str = Field(description="Property name")
    description: str = Field(description="Property description")
    value: str = Field(description="Property value")


class MonitorExport(BaseModel):
    observed_at: datetime.datetime = Field(description="Timestamp of the export")
    peer: str = Field(description="Peer ID")
    session: uuid.UUID = Field(description="Session UUID")

    properties: list[ExportProperty] = Field(description="Properties")
    metrics: list[ExportMetrics] = Field(description="Metrics")
    events: list[ExportEvents] = Field(description="Events")
    bandwidth: Optional[ExportBandwidth] = Field(description="Bandwidth")
