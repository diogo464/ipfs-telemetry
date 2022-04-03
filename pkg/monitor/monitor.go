package monitor

import (
	"context"
	"fmt"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/monitor"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerState struct {
	failedAttemps int
	client        *telemetry.Client
}

type Monitor struct {
	pb.UnimplementedMonitorServer
	h    host.Host
	opts *options
	// this peer was discovered
	cdiscovered chan peer.ID
	// collect telemetry from this peer
	ccollect chan peer.ID
	peers    map[peer.ID]*PeerState
	exporter Exporter
}

func NewMonitor(ctx context.Context, exporter Exporter, o ...Option) (*Monitor, error) {
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

	return &Monitor{
		h:           opts.Host,
		opts:        opts,
		cdiscovered: make(chan peer.ID),
		ccollect:    make(chan peer.ID),
		peers:       make(map[peer.ID]*PeerState),
		exporter:    exporter,
	}, nil
}

func (s *Monitor) Close() {
	s.h.Close()
}

func (s *Monitor) PeerDiscovered(p peer.ID) {
	s.cdiscovered <- p
}

func (s *Monitor) Run(ctx context.Context) {
LOOP:
	for {
		select {
		case discovered := <-s.cdiscovered:
			fmt.Println("Discovered:", discovered)
			if err := s.setupPeer(discovered); err != nil {
				fmt.Println("Failed to setup discovered peer", discovered, ":", err)
			} else {
				s.collectTelemetry(discovered)
			}
		case collect := <-s.ccollect:
			fmt.Println("Collect:", collect)
			s.collectTelemetry(collect)
		case <-ctx.Done():
			break LOOP
		}
	}
}

func (s *Monitor) setupPeer(p peer.ID) error {
	if _, ok := s.peers[p]; ok {
		return nil
	}

	client, err := telemetry.NewClient(s.h, p)
	if err != nil {
		return err
	}

	s.peers[p] = &PeerState{
		failedAttemps: 0,
		client:        client,
	}

	return nil
}

func (s *Monitor) collectTelemetry(p peer.ID) {
	state := s.peers[p]
	if err := s.tryCollectTelemetry(p); err != nil {
		state.failedAttemps += 1
		fmt.Println("Failed to collect telemetry for", p)
		fmt.Println(err)
	}

	if state.failedAttemps >= s.opts.MaxFailedAttemps {
		fmt.Println("Peer", p, "reached max failed attemps")
		delete(s.peers, p)
	} else {
		go func() {
			time.Sleep(s.opts.CollectPeriod)
			s.ccollect <- p
		}()
	}
}

func (s *Monitor) tryCollectTelemetry(p peer.ID) error {
	state := s.peers[p]
	snapshots, err := state.client.Snapshots(context.Background())
	if err != nil {
		return err
	}
	s.exporter.Export(p, snapshots)
	return nil
}
