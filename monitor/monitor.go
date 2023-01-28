package monitor

import (
	"context"

	"github.com/diogo464/telemetry/monitor/metrics"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"go.uber.org/zap"
)

var (
	_ (monitorCommand) = (*monitorCommandDiscover)(nil)
	_ (monitorCommand) = (*monitorCommandDiscoverWithAddr)(nil)
	_ (monitorCommand) = (*monitorCommandPeerFailed)(nil)
)

type monitorCommand interface {
	execute(*Monitor)
}

type Monitor struct {
	logger    *zap.Logger
	metrics   *metrics.Metrics
	ptmetrics *metrics.PeerTaskMetrics

	// Safe for use outside task
	command_sender chan<- monitorCommand
	host           host.Host
	opts           *options
	exporter       Exporter

	// Unsafe for use outside task
	command_receiver <-chan monitorCommand
	peers            map[peer.ID]*peerTask
}

func Start(ctx context.Context, o ...Option) (*Monitor, error) {
	opts := defaults()
	if err := apply(opts, o...); err != nil {
		return nil, err
	}

	if opts.Host == nil {
		h, err := createDefaultHost(ctx)
		if err != nil {
			return nil, err
		}
		opts.Host = h
	}

	if opts.Exporter == nil {
		opts.Exporter = NewNoOpExporter()
	}

	mmetrics, err := metrics.New(opts.MeterProvider)
	if err != nil {
		return nil, err
	}
	ptmetrics, err := metrics.NewPeerTaskMetrics(opts.MeterProvider)
	if err != nil {
		return nil, err
	}
	emetrics, err := metrics.NewExportMetrics(opts.MeterProvider)
	if err != nil {
		return nil, err
	}

	command_channel := make(chan monitorCommand)
	m := &Monitor{
		logger:    opts.Logger,
		metrics:   mmetrics,
		ptmetrics: ptmetrics,

		command_sender: command_channel,
		host:           opts.Host,
		opts:           opts,
		exporter:       &observableExporter{m: emetrics, e: opts.Exporter},

		command_receiver: command_channel,
		peers:            map[peer.ID]*peerTask{},
	}

	mmetrics.RegisterCallback(func(ctx context.Context) {
		mmetrics.ActivePeers.Observe(ctx, int64(len(m.peers)))
	})

	go m.run(ctx)
	return m, nil
}

func (m *Monitor) Discover(ctx context.Context, pid peer.ID) {
	m.sendCommand(newMonitorCommandDiscover(pid))
}

func (m *Monitor) DiscoverWithAddr(ctx context.Context, paddr peer.AddrInfo) {
	m.sendCommand(newMonitorCommandDiscoverWithAddr(paddr))
}

func (m *Monitor) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-m.command_receiver:
			cmd.execute(m)
		}
	}
}

func (m *Monitor) sendCommand(cmd monitorCommand) {
	m.command_sender <- cmd
}

func (m *Monitor) discover(pid peer.ID) {
	m.metrics.DiscoveredPeers.Add(context.Background(), 1)
	if pt, ok := m.peers[pid]; ok {
		m.metrics.RediscoveredPeers.Add(context.Background(), 1)
		m.logger.Info("rediscover peer", zap.String("peer", pid.Pretty()))
		pt.sendCommand(newPeerCommandResetErrors())
		return
	}
	m.logger.Info("discover peer", zap.String("peer", pid.Pretty()))
	m.peers[pid] = newPeerTask(
		pid,
		m.host,
		m.opts,
		m.exporter,
		m,
		m.logger.With(zap.String("peer", pid.Pretty())),
		m.ptmetrics,
	)
}

type monitorCommandDiscover struct {
	pid peer.ID
}

func newMonitorCommandDiscover(pid peer.ID) *monitorCommandDiscover {
	return &monitorCommandDiscover{
		pid: pid,
	}
}

// execute implements monitorCommand
func (c *monitorCommandDiscover) execute(m *Monitor) {
	m.discover(c.pid)
}

type monitorCommandDiscoverWithAddr struct {
	paddr peer.AddrInfo
}

func newMonitorCommandDiscoverWithAddr(paddr peer.AddrInfo) *monitorCommandDiscoverWithAddr {
	return &monitorCommandDiscoverWithAddr{
		paddr: paddr,
	}
}

// execute implements monitorCommand
func (c *monitorCommandDiscoverWithAddr) execute(m *Monitor) {
	m.host.Peerstore().AddAddr(c.paddr.ID, c.paddr.Addrs[0], peerstore.PermanentAddrTTL)
	m.discover(c.paddr.ID)
}

type monitorCommandPeerFailed struct {
	pid peer.ID
}

func newMonitorCommandPeerFailed(pid peer.ID) *monitorCommandPeerFailed {
	return &monitorCommandPeerFailed{pid}
}

// execute implements monitorCommand
func (c *monitorCommandPeerFailed) execute(m *Monitor) {
	delete(m.peers, c.pid)
}
