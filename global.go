package telemetry

var globalTelemetry Telemetry = nil

func SetGlobalTelemetry(t Telemetry) { globalTelemetry = t }

func GetGlobalTelemetry() Telemetry { return globalTelemetry }
