package monitor

import (
	"context"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
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
	// Safe for use outside task
	pid            peer.ID
	host           host.Host
	opts           *options
	exporter       Exporter
	cancel         context.CancelFunc
	command_sender chan<- peerCommand
	monitor        *Monitor2

	// Unsafe for use outside task
	consecutive_errors int
	command_receiver   <-chan peerCommand
	collect_ticker     *time.Ticker
	bandwidth_ticker   *time.Ticker
	session_exported   bool
	metrics_seqn       int
	captures_seqn      map[string]int
	events_seqn        map[string]int
}

func newPeerTask(pid peer.ID, host host.Host, opts *options, exporter Exporter, monitor *Monitor2) *peerTask {
	ctx, cancel := context.WithCancel(context.Background())
	command_channel := make(chan peerCommand, peerTaskCommandBufferSize)
	pt := &peerTask{
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
		session_exported:   false,
		metrics_seqn:       0,
		captures_seqn:      make(map[string]int),
		events_seqn:        make(map[string]int),
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
		logrus.Warnf("Failed to collect telemetry for peer %s: %s", p.pid, err)
		p.fail(err)
	} else {
		logrus.Infof("Successfully collected telemetry for peer %s", p.pid)
	}
}

func (p *peerTask) tryCollectTelemetry(ctx context.Context) error {
	client, err := p.createClient()
	if err != nil {
		return err
	}

	sess, err := client.GetSession(ctx)
	if err != nil {
		logrus.Warnf("Failed to get session for peer %s: %s", p.pid, err)
		return err
	}

	if !p.session_exported {
		logrus.Infof("Exporting session for peer %s: %s", p.pid, sess)
		if err := p.tryExportSession(ctx, client, sess); err != nil {
			logrus.Warnf("Failed to export session for peer %s: %s", p.pid, err)
			return err
		}
	}

	logrus.Infof("Exporting metrics for peer %s: %s", p.pid, sess)
	if err := p.tryExportMetrics(ctx, client, sess); err != nil {
		logrus.Warnf("Failed to export metrics for peer %s: %s", p.pid, err)
		return err
	}

	logrus.Infof("Exporting events for peer %s: %s", p.pid, sess)
	if err := p.tryExportEvents(ctx, client, sess); err != nil {
		logrus.Warnf("Failed to export events for peer %s: %s", p.pid, err)
		return err
	}

	logrus.Infof("Exporting captures for peer %s: %s", p.pid, sess)
	if err := p.tryExportCaptures(ctx, client, sess); err != nil {
		logrus.Warnf("Failed to export captures for peer %s: %s", p.pid, err)
		return err
	}

	return nil
}

func (p *peerTask) tryExportSession(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	props, err := client.GetProperties(ctx)
	if err != nil {
		logrus.Warnf("Failed to get properties for peer %s: %s", p.pid, err)
		return err
	}

	p.exporter.Session(p.pid, sess, props)

	return nil
}

func (p *peerTask) tryExportMetrics(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	cmetrics, err := client.GetMetrics(ctx, p.metrics_seqn)
	if err != nil {
		logrus.Warnf("Failed to get metrics for peer %s: %s", p.pid, err)
		return err
	}

	p.metrics_seqn = cmetrics.SequenceNumber
	p.exporter.Metrics(p.pid, sess, cmetrics.Metrics)

	return nil
}

func (p *peerTask) tryExportEvents(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	descriptors, err := client.GetEventDescriptors(ctx)
	if err != nil {
		logrus.Warnf("Failed to get event descriptors for peer %s: %s", p.pid, err)
		return err
	}

	for _, descriptor := range descriptors {
		cevents, err := client.GetEvent(ctx, descriptor.Name, p.events_seqn[descriptor.Name])
		if err != nil {
			logrus.Warnf("Failed to get events for peer %s: %s", p.pid, err)
			return err
		}

		if len(cevents) > 0 {
			p.events_seqn[descriptor.Name] = cevents[len(cevents)-1].SequenceNumber

			exported_events := make([]Event, 0, len(cevents))
			for _, cevent := range cevents {
				exported_events = append(exported_events, Event{
					Timestamp: cevent.Timestamp,
					Data:      cevent.Data,
				})
			}

			p.exporter.Events(p.pid, sess, descriptor, exported_events)
		}
	}

	return nil
}

func (p *peerTask) tryExportCaptures(ctx context.Context, client *telemetry.Client, sess telemetry.Session) error {
	descriptors, err := client.GetCaptureDescriptors(ctx)
	if err != nil {
		logrus.Warnf("Failed to get capture descriptors for peer %s: %s", p.pid, err)
		return err
	}

	for _, descriptor := range descriptors {
		ccaptures, err := client.GetCapture(ctx, descriptor.Name, p.captures_seqn[descriptor.Name])
		if err != nil {
			logrus.Warnf("Failed to get captures for peer %s: %s", p.pid, err)
			return err
		}

		if len(ccaptures) > 0 {
			p.captures_seqn[descriptor.Name] = ccaptures[len(ccaptures)-1].SequenceNumber

			exported_captures := make([]Capture, 0, len(ccaptures))
			for _, ccapture := range ccaptures {
				exported_captures = append(exported_captures, Capture{
					Timestamp: ccapture.Timestamp,
					Data:      ccapture.Data,
				})
			}

			p.exporter.Captures(p.pid, sess, descriptor, exported_captures)
		}
	}

	return nil
}

func (p *peerTask) bandwidthTest(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, p.opts.BandwidthTimeout)
	defer cancel()
	if err := p.tryBandwidthTest(ctx); err != nil {
		logrus.Warnf("Failed to test bandwidth for peer %s: %s", p.pid, err)
		p.fail(err)
	} else {
		logrus.Infof("Successfully tested bandwidth for peer %s", p.pid)
	}
}

func (p *peerTask) tryBandwidthTest(ctx context.Context) error {
	client, err := p.createClient()
	if err != nil {
		return err
	}

	logrus.Infof("Starting bandwidth test for peer %s", p.pid)
	result, err := client.Bandwidth(ctx, telemetry.DEFAULT_BANDWIDTH_PAYLOAD_SIZE)
	if err != nil {
		return err
	}

	logrus.Infof("Exporting bandwidth test result for peer %s: %s", p.pid, result)
	p.exporter.Bandwidth(p.pid, result)

	return nil
}

func (p *peerTask) createClient() (*telemetry.Client, error) {
	logrus.Tracef("Creating telemetry client for peer %s", p.pid)
	client, err := telemetry.NewClient(p.host, p.pid)
	if err != nil {
		logrus.Warnf("Failed to create telemetry client for peer %s: %s", p.pid, err)
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

func (p *peerTask) failFast(err error) {
	logrus.Errorf("Peer %s failed: %s", p.pid, err)
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
