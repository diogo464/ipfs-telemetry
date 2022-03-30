package main

import (
	"fmt"

	"git.d464.sh/adc/telemetry/plugin"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/plugin"
)

type TelemetryPlugin struct {
	service *telemetry.TelemetryService
}

var _ plugin.PluginDaemonInternal = (*TelemetryPlugin)(nil)

func (*TelemetryPlugin) Name() string {
	return "telemetry"
}

func (*TelemetryPlugin) Version() string {
	return "0.0.0"
}

func (*TelemetryPlugin) Init(env *plugin.Environment) error {
	return nil
}

func (p *TelemetryPlugin) Start(node *core.IpfsNode) error {
	fmt.Println("Starting telemetry plugin")
	service, err := telemetry.NewTelemetryService(node)
	if err != nil {
		return err
	}
	p.service = service
	return nil
}

func (*TelemetryPlugin) Close() error {
	return nil
}

// Plugins is an exported list of plugins that will be loaded by go-ipfs.
var Plugins = []plugin.Plugin{
	&TelemetryPlugin{},
}
