package telemetry

var globalTelemetry Telemetry = NewNoOpTelemetry()

func SetGlobalTelemetry(t Telemetry) { globalTelemetry = t }

func GetGlobalTelemetry() Telemetry { return globalTelemetry }
