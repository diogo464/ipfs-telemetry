import logging
import json
import opentelemetry.proto.metrics.v1.metrics_pb2 as otlp

from re import IGNORECASE, UNICODE, compile

from backend.monitor import Export, ExportMetrics

logger = logging.getLogger(__name__)


_non_letters_digits_underscore_re = compile(r"[^\w]", UNICODE | IGNORECASE)


def _sanitize(key: str) -> str:
    # https://github.com/open-telemetry/opentelemetry-python/blob/main/exporter/opentelemetry-exporter-prometheus/src/opentelemetry/exporter/prometheus/__init__.py
    """sanitize the given metric name or label according to Prometheus rule.
    Replace all characters other than [A-Za-z0-9_] with '_'.
    """
    return _non_letters_digits_underscore_re.sub("_", key)


def _convert_numeric_dp(dp: otlp.NumberDataPoint) -> float:
    if dp.HasField("as_int"):
        return float(dp.as_int)
    else:
        return dp.as_double


def _milli_timestamp_from_dp(dp) -> int:
    return int(dp.time_unix_nano / 1000000)


def _attributes_to_labels(attrs) -> dict[str, str]:
    attributes = {}
    for attr in attrs:
        key = attr.key
        value = None
        if attr.value.HasField("string_value"):
            value = attr.value.string_value
        elif attr.value.HasField("int_value"):
            value = attr.value.int_value
        elif attr.value.HasField("double_value"):
            value = attr.value.double_value
        else:
            logger.warn("Unhandled attribute type")
        if value is not None:
            attributes[key] = str(value)
    return attributes


class PromMetric:
    def __init__(self, labels: dict[str, str], value: float, suffix: str = ""):
        self.labels = labels
        self.value = value
        self.suffix = suffix


class PromGroup:
    TYPE_COUNTER = "counter"
    TYPE_GAUGE = "gauge"
    TYPE_HISTOGRAM = "histogram"
    TYPE_SUMMARY = "summary"
    TYPE_UNTYPED = "untyped"

    def __init__(
        self,
        name: str,
        type: str,
        help: str = "",
        timestamp: int = 0,
        labels: dict[str, str] = {},
    ):
        self.name = name
        self.type = type
        self.help = help
        self.timestamp = timestamp
        self.labels = labels
        self.metrics: list[PromMetric] = []

    def add_metric(self, labels: dict[str, str], value: float, suffix: str = ""):
        self.metrics.append(PromMetric(self.labels | labels, value, suffix))

    def build(self) -> str:
        lines = []
        lines.append(f"# HELP {self.name} {self.help}")
        lines.append(f"# TYPE {self.name} {self.type}")
        for metric in self.metrics:
            # Use json.dumps to escape the label values
            labels = ",".join(
                [f"{k}={json.dumps(v)}" for k, v in metric.labels.items()]
            )
            if labels:
                labels = "{" + labels + "}"
            lines.append(
                f"{self.name}{metric.suffix}{labels} {metric.value} {self.timestamp}"
            )
        return "\n".join(lines) + "\n"


def convert_export(export: Export) -> str:
    output = ""
    metrics = [m.decode_otlp() for m in export.metrics]
    for resource_metrics in metrics:
        resource = resource_metrics.resource
        base_labels = _attributes_to_labels(resource.attributes)
        base_labels["peerid"] = export.peer
        base_labels["session"] = str(export.session)
        for sm in resource_metrics.scope_metrics:
            output += _convert_scope_metrics(base_labels, sm) + "\n"
    return output


def _convert_scope_metrics(l: dict[str, str], sm: otlp.ScopeMetrics) -> str:
    metrics: list[otlp.Metric] = [m for m in sm.metrics]
    output = ""
    for metric in metrics:
        output += _convert_metric(l, metric) + "\n"
    return output


def _convert_metric(l: dict[str, str], metric: otlp.Metric) -> str:
    name = _sanitize(metric.name)
    help = metric.description

    if metric.HasField("gauge") or metric.HasField("sum"):
        metric_type = (
            PromGroup.TYPE_GAUGE if metric.HasField("gauge") else PromGroup.TYPE_COUNTER
        )
        datapoints = (
            metric.gauge.data_points
            if metric.HasField("gauge")
            else metric.sum.data_points
        )
        ts = _milli_timestamp_from_dp(datapoints[0])
        group = PromGroup(name, metric_type, help, timestamp=ts, labels=l)
        for dp in datapoints:
            extra_labels = _attributes_to_labels(dp.attributes)
            value = _convert_numeric_dp(dp)
            group.add_metric(extra_labels, value)
        return group.build()

    elif metric.HasField("histogram"):
        ts = _milli_timestamp_from_dp(metric.histogram.data_points[0])
        group = PromGroup(name, PromGroup.TYPE_HISTOGRAM, help, timestamp=ts, labels=l)
        for dp in metric.histogram.data_points:
            acum = 0  # https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#histogram
            extra_labels = _attributes_to_labels(dp.attributes)
            for i, count in enumerate(dp.bucket_counts):
                le = (
                    str(dp.explicit_bounds[i])
                    if i < len(dp.explicit_bounds)
                    else "+Inf"
                )
                acum += count
                group.add_metric({**extra_labels, "le": le}, acum, "_bucket")
            group.add_metric(extra_labels, dp.count, "_count")
            group.add_metric(extra_labels, dp.sum, "_sum")
        return group.build()

    raise Exception(f"Unhandled metric type: {metric}")
