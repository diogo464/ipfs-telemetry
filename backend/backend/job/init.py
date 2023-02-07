import logging
import asyncio
import datetime

from nats.js.api import StreamConfig
from nats.js.client import JetStreamContext

import backend.env
import backend.db


async def upsert_stream(js: JetStreamContext, config: StreamConfig):
    try:
        assert config.name is not None
        await js.stream_info(config.name)
        await js.update_stream(config)
    except:
        await js.add_stream(config)


async def main():
    conn_info = backend.env.create_db_conn_info()
    engine = backend.db.create_engine(conn_info)
    nc = await backend.env.create_nats_client()
    js = nc.jetstream()

    if backend.db.requires_setup_database(engine):
        logging.info("Setting up database")
        backend.db.setup_database(engine)

    # Configure telemetry JetStream
    telemetry_config = StreamConfig(
        name="telemetry",
        description="Telemetry exports from monitors",
        subjects=["telemetry.export"],
        max_age=datetime.timedelta(days=90).total_seconds(),
        max_bytes=2 * (1024**3),  # 200 GiB
    )
    await upsert_stream(js, telemetry_config)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(main())
