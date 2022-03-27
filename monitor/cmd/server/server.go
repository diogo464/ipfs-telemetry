package server

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"git.d464.sh/adc/telemetry/monitor/dal"
	"git.d464.sh/adc/telemetry/telemetry"
	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	MAX_FAILED_ATTEMPTS = 3
	COLLECT_PERIOD      = time.Second * 5
)

type PeerState struct {
	failedAttemps int
	client        *telemetry.TelemetryClient
}

type Server struct {
	h  host.Host
	db *sql.DB

	// this peer was discovered
	cdiscovered chan peer.ID
	// collect telemetry from this peer
	ccollect chan peer.ID

	// how many consecutive attemps at collecting telemetry have failed
	peers map[peer.ID]*PeerState
}

func NewServer(h host.Host, db *sql.DB) (*Server, error) {
	return &Server{
		h:           h,
		db:          db,
		cdiscovered: make(chan peer.ID),
		ccollect:    make(chan peer.ID),
		peers:       make(map[peer.ID]*PeerState),
	}, nil
}

func (s *Server) PeerDiscovered(p peer.ID) {
	s.cdiscovered <- p
}

func (s *Server) StartMonitoring(ctx context.Context) {
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
			break
		}
	}
}

func (s *Server) setupPeer(p peer.ID) {
	s.peers[p] = &PeerState{
		failedAttemps: 0,
		client:        telemetry.NewTelemetryClient(s.h, p),
	}
}

func (s *Server) collectTelemetry(p peer.ID) {
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

func (s *Server) tryCollectTelemetry(p peer.ID) error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	state := s.peers[p]
	client := state.client
	snapshots, err := client.Snapshots(ctx)
	if err != nil {
		return err
	}

	sessionID, err := dal.Session(context.TODO(), s.db, p, client.Session().UUID)
	if err != nil {
		return err
	}

	s.processSnapshots(p, sessionID, snapshots.Snapshots)
	return nil
}

func (s *Server) processSnapshots(p peer.ID, sessionID int, snapshots []*snapshot.Snapshot) {
	fmt.Println("Processing", len(snapshots), "snapshots")
	for _, ss := range snapshots {
		switch ss.Tag {
		case snapshot.TAG_ROUTING_TABLE:
			s.processSnapshotRoutingTable(p, sessionID, ss)
		case snapshot.TAG_NETWORK:
			s.processSnapshotNetwork(p, sessionID, ss)
		case snapshot.TAG_PING:
			s.processSnapshotPing(p, sessionID, ss)
		}
	}
}

func (s *Server) processSnapshotRoutingTable(p peer.ID, sessionID int, snapshot *snapshot.Snapshot) {
	if err := dal.RoutingTable(context.TODO(), s.db, sessionID, snapshot); err != nil {
		fmt.Println(err)
	}
}

func (s *Server) processSnapshotNetwork(p peer.ID, sessionID int, snapshot *snapshot.Snapshot) {
}

func (s *Server) processSnapshotPing(p peer.ID, sessionID int, snapshot *snapshot.Snapshot) {
}
