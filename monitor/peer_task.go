package monitor

import (
	"context"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/monitor/metrics"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

const (
	peerTaskCommandBufferSize = 8
)

var (
	_ (peerCommand) = (*peerCommandResetErrors)(nil)
)

type peerCommand interface {
	execute(*peerTask)
}

type peerTask struct {
	logger  *zap.Logger
	metrics *metrics.PeerTaskMetrics

	// Safe for use outside task
	pid            peer.ID
	host           host.Host
	opts           *options
	exporter       Exporter
	cancel         context.CancelFunc
	command_sender chan<- peerCommand
	monitor        *Monitor

	// Unsafe for use outside task
	consecutive_errors int
	command_receiver   <-chan peerCommand
	collect_ticker     *time.Ticker
	bandwidth_ticker   *time.Ticker
	client_state       *telemetry.ClientState
}

func newPeerTask(pid peer.ID, host host.Host, opts *options, exporter Exporter, monitor *Monitor, logger *zap.Logger, m *metrics.PeerTaskMetrics) *peerTask {
	ctx, cancel := context.WithCancel(context.Background())
	command_channel := make(chan peerCommand, peerTaskCommandBufferSize)
	pt := &peerTask{
		logger:  logger,
		metrics: m,

		pid:            pid,
		host:           host,
		opts:           opts,
		exporter:       exporter,
		cancel:         cancel,
		command_sender: command_channel,
		monitor:        monitor,

		consecutive_errors: 0,
		command_receiver:   command_channel,
		collect_ticker:     time.NewTicker(opts.CollectPeriod),
		bandwidth_ticker:   time.NewTicker(opts.BandwidthPeriod),
		client_state:       nil,
	}
	go pt.run(ctx)
	return pt
}

func (p *peerTask) run(ctx context.Context) {

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case cmd := <-p.command_receiver:
			cmd.execute(p)
		case <-p.collect_ticker.C:
			p.collectTelemetry(ctx)
		case <-p.bandwidth_ticker.C:
			p.bandwidthTest(ctx)
		}
	}

}

func (p *peerTask) sendCommand(cmd peerCommand) {
	p.command_sender <- cmd
}

func (p *peerTask) collectTelemetry(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, p.opts.CollectTimeout)
	defer cancel()
	if err := p.tryCollectTelemetry(ctx); err != nil {
		p.logger.Warn("failed to collect telemetry", zap.Error(err))
		p.fail(err)
	} else {
		p.logger.Info("successfully collected telemetry")
		p.success()
	}
}

func (p *peerTask) tryCollectTelemetry(ctx context.Context) error {
	client, err := p.createClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		p.client_state = client.GetClientState()
	}()

	sess, err := client.GetSession(ctx)
	if err != nil {
		p.logger.Warn("failed to get session")
		return err
	}

	p.logger.Info("exporting session", zap.Any("session", sess))
	if err := p.tryExportSession(ctx, client, sess); err != nil {
		p.logger.Warn("failed to export session", zap.Error(err))
		return err
	}

	p.logger.Info("exporting properties", zap.Any("session", sess))
	if err := p.tryExportProperties(ctx, client, sess); err != nil {
		p.logger.Warn("failed to export properties", zap.Error(err))
		return err
	}

	p.logger.Info("exporting metrics", zap.Any("session", sess))
	if err := p.tryExportMetrics(ctx, client, sess); err != nil {
		p.logger.Warn("failed to export metrics", zap.Error(err))
		return err
	}

	p.logger.Info("exporting events", zap.Any("session", sess))
	if err := p.tryExportEvents(ctx, client, sess); err != nil {
		p.logger.Warn("failed to export events", zap.Error(err))
		return err
	}

	return nil
}

func (p *peerTask) tryExportSession(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	p.exporter.Session(p.pid, sess)
	return nil
}

func (p *peerTask) tryExportProperties(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	properties, err := client.GetProperties(ctx)
	if err != nil {
		p.logger.Warn("failed to get properties", zap.Error(err))
		return err
	}
	p.exporter.Properties(p.pid, sess, properties)
	return nil
}

func (p *peerTask) tryExportMetrics(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	cmetrics, err := client.GetMetrics(ctx)
	if err != nil {
		p.logger.Warn("failed to get metrics", zap.Error(err))
		return err
	}

	p.exporter.Metrics(p.pid, sess, cmetrics)

	return nil
}

func (p *peerTask) tryExportEvents(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	descriptors, err := client.GetStreamDescriptors(ctx)
	if err != nil {
		p.logger.Warn("failed to get stream descriptors", zap.Error(err))
		return err
	}

	for _, descriptor := range descriptors {
		if ed, ok := descriptor.Type.(*telemetry.StreamTypeEvent); ok {
			events, err := client.GetEvents(ctx, descriptor.ID)
			if err != nil {
				p.logger.Warn("failed to get events", zap.Error(err))
				return err
			}

			if len(events) > 0 {
				p.exporter.Events(p.pid, sess, telemetry.EventDescriptor{
					Scope:       ed.Scope,
					Name:        ed.Name,
					Description: ed.Description,
				}, events)
			}
		}
	}

	return nil
}

func (p *peerTask) bandwidthTest(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, p.opts.BandwidthTimeout)
	defer cancel()
	if err := p.tryBandwidthTest(ctx); err != nil {
		p.logger.Warn("failed to test bandwidth", zap.Error(err))
		p.fail(err)
		p.metrics.CollectFailure.Add(ctx, 1, metrics.KeyPeerID.String(p.pid.String()))
	} else {
		p.logger.Info("successfully tested bandwidth")
		p.success()
		p.metrics.CollectCompleted.Add(ctx, 1, metrics.KeyPeerID.String(p.pid.String()))
	}
}

func (p *peerTask) tryBandwidthTest(ctx context.Context) error {
	client, err := p.createClient(ctx)
	if err != nil {
		return err
	}

	p.logger.Info("starting bandwidth test")
	result, err := client.Bandwidth(ctx, telemetry.DEFAULT_BANDWIDTH_PAYLOAD_SIZE)
	if err != nil {
		return err
	}

	p.logger.Info("exporting bandwidth test result", zap.Any("result", result))
	p.exporter.Bandwidth(p.pid, result)

	return nil
}

func (p *peerTask) createClient(ctx context.Context) (*telemetry.Client, error) {
	p.logger.Info("creating telemetry client", zap.Any("state", p.client_state))
	client, err := telemetry.NewClient(
		ctx,
		telemetry.WithClientLibp2pDial(p.host, p.pid),
		telemetry.WithClientState(p.client_state),
	)
	if err != nil {
		p.logger.Warn("failed to create telemetry client", zap.Error(err))
		return nil, err
	}
	return client, nil
}

func (p *peerTask) fail(err error) {
	p.consecutive_errors++
	if p.consecutive_errors >= p.opts.MaxFailedAttemps {
		p.failFast(err)
	}
}

func (p *peerTask) success() {
	p.consecutive_errors = 0
}

func (p *peerTask) failFast(err error) {
	p.logger.Error("peer failure", zap.Error(err))
	p.cancel()
	p.monitor.sendCommand(newMonitorCommandPeerFailed(p.pid))
}

type peerCommandResetErrors struct{}

func newPeerCommandResetErrors() *peerCommandResetErrors {
	return &peerCommandResetErrors{}
}

// execute implements peerCommand
func (*peerCommandResetErrors) execute(p *peerTask) {
	p.consecutive_errors = 0
}
