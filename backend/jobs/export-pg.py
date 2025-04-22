import logging
import asyncio
import db
import env
import export
import export.monitor
import export.discovery

from nats.js.api import ConsumerConfig
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session


class PostgresTelemetryExporter(export.TelemetryExporter):
    def __init__(
        self, session: Session, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.session = session
        self.logger = logger

    async def export(self, export: export.monitor.MonitorExport):
        try:
            model_data = db.convert_export(export)
            model_data.bulk_save(self.session)
            self.session.commit()
        except IntegrityError:
            self.session.rollback()
            self.logger.exception("Integrity error, skipping export", exc_info=True)


class PostgresDiscoveryExporter(export.DiscoveryExporter):
    def __init__(
        self, session: Session, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.session = session
        self.logger = logger

    async def export(self, export: export.discovery.DiscoveryNotification):
        print(export)


async def main():
    env.setup_logging()
    client = await env.create_nats_client()
    conn_info = env.create_db_conn_info()
    engine = db.create_engine(conn_info)

    telem_stream = await export.NatsTelemetryStream.new(
        client,
        consumer=ConsumerConfig(
            durable_name="exporter-pg",
            description="Export telemetry to VictoriaMetrics",
        ),
    )
    telem_exporter = PostgresTelemetryExporter(db.create_session(engine))
    telem_task = asyncio.create_task(export.export(telem_stream, telem_exporter))

    disc_stream = await export.NatsDiscoveryStream.new(client, queue="exporter-pg")
    disc_exporter = PostgresDiscoveryExporter(db.create_session(engine))
    disc_task = asyncio.create_task(export.export(disc_stream, disc_exporter))

    done, _ = await asyncio.wait(
        [telem_task, disc_task], return_when=asyncio.FIRST_EXCEPTION
    )

    exception = next(iter(done)).exception()
    if exception is not None:
        raise exception


if __name__ == "__main__":
    asyncio.run(main())

