import logging
import asyncio
import backend.db
import backend.env
import backend.prom
import backend.monitor
import backend.exporter
import backend.discovery

from nats.js.api import ConsumerConfig
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session


class PostgresTelemetryExporter(backend.exporter.TelemetryExporter):
    def __init__(
        self, session: Session, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.session = session
        self.logger = logger

    async def export(self, export: backend.monitor.MonitorExport):
        try:
            model_data = backend.db.convert_export(export)
            model_data.bulk_save(self.session)
            self.session.commit()
        except IntegrityError:
            self.session.rollback()
            self.logger.exception("Integrity error, skipping export", exc_info=True)


class PostgresDiscoveryExporter(backend.exporter.DiscoveryExporter):
    def __init__(
        self, session: Session, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.session = session
        self.logger = logger

    async def export(self, export: backend.discovery.DiscoveryNotification):
        print(export)


async def main():
    backend.env.setup_logging()
    client = await backend.env.create_nats_client()
    conn_info = backend.env.create_db_conn_info()
    engine = backend.db.create_engine(conn_info)

    telem_stream = await backend.exporter.NatsTelemetryStream.new(
        client,
        consumer=ConsumerConfig(
            durable_name="exporter-pg",
            description="Export telemetry to VictoriaMetrics",
        ),
    )
    telem_exporter = PostgresTelemetryExporter(backend.db.create_session(engine))
    telem_task = asyncio.create_task(
        backend.exporter.export(telem_stream, telem_exporter)
    )

    disc_stream = await backend.exporter.NatsDiscoveryStream.new(
        client, queue="exporter-pg"
    )
    disc_exporter = PostgresDiscoveryExporter(backend.db.create_session(engine))
    disc_task = asyncio.create_task(backend.exporter.export(disc_stream, disc_exporter))

    done, _ = await asyncio.wait(
        [telem_task, disc_task], return_when=asyncio.FIRST_EXCEPTION
    )
    raise next(iter(done)).exception()


if __name__ == "__main__":
    asyncio.run(main())
