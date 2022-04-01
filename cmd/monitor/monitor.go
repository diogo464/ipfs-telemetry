package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/monitor"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	MAX_FAILED_ATTEMPTS = 3
	COLLECT_PERIOD      = time.Second * 1
)

type PeerState struct {
	failedAttemps int
	client        *telemetry.Client
}

type Monitor struct {
	pb.UnimplementedMonitorServer
	h  host.Host
	db *sql.DB

	// this peer was discovered
	cdiscovered chan peer.ID
	// collect telemetry from this peer
	ccollect chan peer.ID

	// how many consecutive attemps at collecting telemetry have failed
	peers map[peer.ID]*PeerState
}

func NewMonitor(h host.Host, db *sql.DB) (*Monitor, error) {
	return &Monitor{
		h:           h,
		db:          db,
		cdiscovered: make(chan peer.ID),
		ccollect:    make(chan peer.ID),
		peers:       make(map[peer.ID]*PeerState),
	}, nil
}

func (s *Monitor) Discover(ctx context.Context, req *pb.DiscoverRequest) (*emptypb.Empty, error) {
	p, err := peer.Decode(req.Peer)
	if err != nil {
		return nil, err
	}
	s.PeerDiscovered(p)
	return &emptypb.Empty{}, nil
}

func (s *Monitor) PeerDiscovered(p peer.ID) {
	s.cdiscovered <- p
}

func (s *Monitor) StartMonitoring(ctx context.Context) {
LOOP:
	for {
		select {
		case discovered := <-s.cdiscovered:
			fmt.Println("Discovered:", discovered)
			s.setupPeer(discovered)
			s.collectTelemetry(discovered)
		case collect := <-s.ccollect:
			fmt.Println("Collect:", collect)
			s.collectTelemetry(collect)
		case <-ctx.Done():
			break LOOP
		}
	}
}

func (s *Monitor) setupPeer(p peer.ID) {
	if _, ok := s.peers[p]; !ok {
		s.peers[p] = &PeerState{
			failedAttemps: 0,
			client:        telemetry.NewClient(s.h, p),
		}
	}
}

func (s *Monitor) collectTelemetry(p peer.ID) {
	state := s.peers[p]
	if err := s.tryCollectTelemetry(p); err != nil {
		state.failedAttemps += 1
		fmt.Println("Failed to collect telemetry for", p)
		fmt.Println(err)
	}

	if state.failedAttemps >= MAX_FAILED_ATTEMPTS {
		fmt.Println("Peer", p, "reached max failed attemps")
		delete(s.peers, p)
	} else {
		go func() {
			time.Sleep(COLLECT_PERIOD)
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
	for _, snapshot := range snapshots {
		if err := s.handleSnapshot(p, snapshot); err != nil {
			return err
		}
	}
	return nil
}
