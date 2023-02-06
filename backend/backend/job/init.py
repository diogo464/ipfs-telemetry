import logging
import asyncio
import datetime

from nats.js.api import StreamConfig
from nats.js.client import JetStreamContext

import backend.env


async def upsert_stream(js: JetStreamContext, config: StreamConfig):
    try:
        assert config.name is not None
        await js.stream_info(config.name)
        await js.update_stream(config)
    except:
        await js.add_stream(config)


async def main():
    nc = await backend.env.create_nats_client()
    js = nc.jetstream()

    print(await js.account_info())

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
