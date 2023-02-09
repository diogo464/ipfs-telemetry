import logging
import asyncio
import requests
import backend.env
import backend.prom
import backend.monitor
import backend.exporter

from nats.js.api import ConsumerConfig


class VictoriaMetricsExporter(backend.exporter.TelemetryExporter):
    def __init__(
        self, vm_import_url: str, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.vm_import_url = vm_import_url
        self.logger = logger

    async def export(self, export: backend.monitor.MonitorExport):
        output = backend.prom.convert_export(export)
        self.logger.info(
            f"exporting from {export.peer}/{export.session} observed at {export.observed_at} with {len(output)} bytes"
        )
        requests.post(self.vm_import_url, data=output).raise_for_status()


async def main():
    backend.env.setup_logging()
    client = await backend.env.create_nats_client()
    stream = await backend.exporter.NatsTelemetryStream.new(
        client,
        consumer=ConsumerConfig(
            durable_name="exporter-vm",
            description="Export telemetry to VictoriaMetrics",
        ),
    )
    exporter = VictoriaMetricsExporter(
        f"{backend.env.VM_ENDPOINT}/api/v1/import/prometheus"
    )
    await backend.exporter.export(stream, exporter)


if __name__ == "__main__":
    asyncio.run(main())
