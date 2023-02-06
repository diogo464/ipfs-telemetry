import logging
import requests

from nats.js.api import ConsumerConfig

import backend.env
import backend.monitor
import backend.prom

logger = logging.getLogger(__name__)

CONSUMER_NAME = "exporter-vm"


async def run():
    nc = await backend.env.create_nats_client()
    js = nc.jetstream()
    consumer = ConsumerConfig(
        durable_name=CONSUMER_NAME,
        description="Export telemetry to VictoriaMetrics",
    )
    sub = await js.subscribe(
        "telemetry.export", queue=CONSUMER_NAME, config=consumer, stream="telemetry"
    )

    vm_import_url = f"{backend.env.VM_ENDPOINT}/api/v1/import/prometheus"
    logger.info(f"exporting to Victoria Metrics at {vm_import_url}")

    async for msg in sub.messages:
        export = backend.monitor.Export.parse_raw(msg.data)
        output = backend.prom.convert_export(export)
        logger.info(f"Sending {len(output)} bytes to Victoria Metrics")
        requests.post(vm_import_url, data=output).raise_for_status()
        await msg.ack()
