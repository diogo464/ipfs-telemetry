from __future__ import annotations

import abc
import logging

from typing import Optional, TypeVar, Generic
from nats.js.api import ConsumerConfig
from nats.aio.client import Client as NatsClient
from nats.aio.subscription import Subscription as NatsSubscription
from export.monitor import MonitorExport
from export.discovery import DiscoveryNotification

T = TypeVar("T")
logger = logging.getLogger(__name__)


class Exporter(Generic[T], abc.ABC):
    @abc.abstractmethod
    async def export(self, export: T):
        pass


class ExportStream(abc.ABC, Generic[T]):
    @abc.abstractmethod
    async def try_consume(self, exporter: Exporter[T]):
        """
        Tries to consume the next export from the stream.
        The exporter is called with the export as an argument.
        In case of an error the export should not be acknowledged and this method should rethrow the error.
        """
        pass


class NatsStream(ExportStream[T]):
    def __init__(self, subscription: NatsSubscription, is_jetstream: bool):
        self.sub = subscription
        self.is_jetstream = is_jetstream

    async def try_consume(self, exporter: Exporter[T]):
        msg = await self.sub.next_msg(timeout=None)
        export = self.parse_message(msg.data)
        await exporter.export(export)
        if self.is_jetstream:
            logger.info(
                f"acknowledging NATS message for JetStream stream '{msg.metadata.stream}' with sequence {msg.metadata.sequence}"
            )
            await msg.ack()

    @abc.abstractmethod
    def parse_message(self, msg: bytes) -> T:
        pass

    @staticmethod
    async def _new_js_sub(
        nc: NatsClient,
        stream: str,
        subject: str,
        consumer: Optional[ConsumerConfig] = None,
    ) -> NatsSubscription:
        """
        Creates a new NATS stream for exporting data.

        :param nc: NATS client
        :param stream: NATS stream name
        :param subject: NATS subject to filter on
        :param consumer: NATS consumer configuration. Defaults to an ephemeral consumer.
        """

        js = nc.jetstream()
        if consumer is None:
            consumer = ConsumerConfig(
                description=f"Ephemeral consumer for {subject} on {stream}",
            )
        sub = await js.subscribe(subject, queue=consumer.durable_name, stream=stream)
        logger.info(
            f"created NATS export stream for subject '{sub.subject}' with queue '{sub.queue}'"
        )
        return sub

    @staticmethod
    async def _new_sub(nc: NatsClient, subject: str, queue: str) -> NatsSubscription:
        """
        Creates a new NATS subscription.

        :param nc: NATS client
        :param subject: NATS subject to filter on
        :param queue: NATS queue to join
        """
        sub = await nc.subscribe(subject, queue=queue)
        logger.info(
            f"created NATS subscription for subject '{sub.subject}' on queue '{sub.queue}'"
        )
        return sub


class NatsTelemetryStream(NatsStream):
    def parse_message(self, msg: bytes) -> MonitorExport:
        return MonitorExport.parse_raw(msg)

    @staticmethod
    async def new(
        nc: NatsClient,
        stream: str = "telemetry",
        subject: str = "telemetry.export",
        consumer: Optional[ConsumerConfig] = None,
    ) -> NatsTelemetryStream:
        """
        Creates a new NATS stream for exporting telemetry data.

        :param nc: NATS client
        :param stream: NATS stream name
        :param subject: NATS subject to filter on
        :param consumer: NATS consumer configuration. Defaults to an ephemeral consumer.
        """
        subscription = await NatsStream._new_js_sub(nc, stream, subject, consumer)
        return NatsTelemetryStream(subscription, is_jetstream=True)


class NatsDiscoveryStream(NatsStream):
    def parse_message(self, msg: bytes) -> DiscoveryNotification:
        return DiscoveryNotification.parse_raw(msg)

    @staticmethod
    async def new(
        nc: NatsClient, subject: str = "discovery", queue: str = ""
    ) -> NatsDiscoveryStream:
        """
        Creates a new NATS stream for exporting discovery notifications.

        :param nc: NATS client
        :param stream: NATS stream name
        :param subject: NATS subject to filter on
        :param consumer: NATS consumer configuration. Defaults to an ephemeral consumer.
        """
        subscription = await NatsStream._new_sub(nc, subject, queue)
        return NatsDiscoveryStream(subscription, is_jetstream=False)


TelemetryExporter = Exporter[MonitorExport]
DiscoveryExporter = Exporter[DiscoveryNotification]

TelemetryExportStream = ExportStream[MonitorExport]
DiscoveryExportStream = ExportStream[DiscoveryNotification]


async def export(stream: ExportStream[T], exporter: Exporter[T]):
    """
    Consumes exports from a stream and exports them using the exporter.
    """
    while True:
        await stream.try_consume(exporter)

