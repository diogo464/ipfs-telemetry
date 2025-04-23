import env
import requests
import logging
import asyncio

import export
import export.monitor

from nats.js.api import ConsumerConfig
from opentelemetry.proto.collector.metrics.v1.metrics_service_pb2 import (
    ExportMetricsServiceRequest,
)


class VictoriaMetricsExporter(export.TelemetryExporter):
    def __init__(
        self, vm_import_url: str, logger: logging.Logger = logging.getLogger(__name__)
    ):
        self.vm_import_url = vm_import_url
        self.logger = logger

    async def export(self, export: export.monitor.MonitorExport):
        for metrics in export.metrics:
            resource_metrics = metrics.decode_otlp()
            export_request = ExportMetricsServiceRequest(
                resource_metrics=[resource_metrics]
            )
            request_data = export_request.SerializeToString()
            requests.post(self.vm_import_url, data=request_data).raise_for_status()


async def main():
    client = await env.create_nats_client()
    stream = await export.NatsTelemetryStream.new(
        client,
        consumer=ConsumerConfig(
            durable_name="exporter-vm",
            description="Export telemetry to VictoriaMetrics",
        ),
    )
    exporter = VictoriaMetricsExporter(f"{env.VM_ENDPOINT}/opentelemetry/v1/metrics")
    await export.export(stream, exporter)


if __name__ == "__main__":
    env.setup_logging()
    asyncio.run(main())

