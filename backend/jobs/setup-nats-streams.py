import env
import logging
import datetime
import asyncio

from nats.js.api import StreamConfig
from nats.js.client import JetStreamContext


async def upsert_stream(js: JetStreamContext, config: StreamConfig):
    logging.info(f"setting up stream {config.name} with config {config}")
    try:
        assert config.name is not None
        await js.stream_info(config.name)
        logging.info(f"stream {config.name} already exists, updating config")
        await js.update_stream(config)
    except:
        logging.info(f"stream {config.name} not found, creating")
        await js.add_stream(config)


async def main():
    nc = await env.create_nats_client()
    js = nc.jetstream()

    # Configure telemetry JetStream
    telemetry_config = StreamConfig(
        name="telemetry",
        description="Telemetry exports from monitors",
        subjects=["telemetry.export"],
        max_age=datetime.timedelta(days=90).total_seconds(),
        max_bytes=200 * (1024**3),  # 200 GiB
    )
    await upsert_stream(js, telemetry_config)


if __name__ == "__main__":
    env.setup_logging()
    asyncio.run(main())

