package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/diogo464/telemetry/pkg/actionqueue"
	pb "github.com/diogo464/telemetry/pkg/proto/monitor"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
)

const (
	ActionDiscover = iota
	ActionTelemetry
	ActionBandwidth
	ActionRemovePeer
)

type action struct {
	kind int
	pid  peer.ID
}

type actionFn = func(*peerState) error

type peerState struct {
	mu sync.Mutex
	id peer.ID
	// consecutive failed attempts
	failedAttemps int
	lastSession   telemetry.Session
	nextSeqN      uint64
}

type Monitor struct {
	pb.UnimplementedMonitorServer
	h        host.Host
	opts     *options
	peers    map[peer.ID]*peerState
	exporter Exporter

	actions *actionqueue.Queue[action]
	caction chan actionqueue.Action[action]
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
		h:        opts.Host,
		opts:     opts,
		peers:    make(map[peer.ID]*peerState),
		exporter: exporter,

		actions: actionqueue.NewQueue[action](),
		caction: make(chan actionqueue.Action[action], 8),
	}, nil
}

func (s *Monitor) Close() {
	s.h.Close()
}

func (s *Monitor) PeerDiscovered(p peer.ID) {
	s.caction <- actionqueue.Now(&action{
		kind: ActionDiscover,
		pid:  p,
	})
}

func (s *Monitor) Run(ctx context.Context) {
LOOP:
	for {
		logrus.Debug("Monitor main loop")
		actionTimer := s.actions.TimerUntilAction()
		select {
		case <-actionTimer.C:
			action := s.actions.Pop()
			switch action.kind {
			case ActionDiscover:
				logrus.WithField("peer", action.pid).Debug("action discovery")
				s.onActionDiscover(action.pid)
			case ActionTelemetry:
				logrus.WithField("peer", action.pid).Debug("action telemetry")
				s.executePeerAction(action.pid, ActionTelemetry, s.opts.CollectPeriod, s.collectTelemetry)
			case ActionBandwidth:
				logrus.WithField("peer", action.pid).Debug("action bandwidth")
				s.executePeerAction(action.pid, ActionBandwidth, s.opts.BandwidthPeriod, s.collectBandwidth)
			case ActionRemovePeer:
				logrus.WithField("peer", action.pid).Debug("action remove peer")
				delete(s.peers, action.pid)
			}
		case action := <-s.caction:
			s.actions.Push(action)
		case <-ctx.Done():
			break LOOP
		}
	}
}

func (s *Monitor) executePeerAction(p peer.ID, a int, t time.Duration, fn actionFn) {
	if state, ok := s.peers[p]; ok {
		go func() {
			state.mu.Lock()
			defer state.mu.Unlock()

			var delay time.Duration
			if err := fn(state); err != nil {
				delay = s.opts.RetryInterval
				s.peerError(state, err)
			} else {
				state.failedAttemps = 0
				delay = t
			}

			s.caction <- actionqueue.After(&action{
				kind: a,
				pid:  p,
			}, delay)
		}()
	}
}

func (s *Monitor) onActionDiscover(p peer.ID) {
	if err := s.setupPeer(p); err == nil {
		logrus.WithField("peer", p).Debug("peer setup")
		s.caction <- actionqueue.Now(&action{
			kind: ActionTelemetry,
			pid:  p,
		})
		s.caction <- actionqueue.Now(&action{
			kind: ActionBandwidth,
			pid:  p,
		})
	} else {
		logrus.WithField("peer", p).Error("failed to setup peer: ", err)
	}
}

func (s *Monitor) setupPeer(p peer.ID) error {
	if _, ok := s.peers[p]; ok {
		return nil
	}
	s.peers[p] = &peerState{
		mu:            sync.Mutex{},
		id:            p,
		failedAttemps: 0,
		lastSession:   telemetry.InvalidSession,
		nextSeqN:      0,
	}
	return nil
}

// must be holding the state's lock
func (s *Monitor) collectTelemetry(state *peerState) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.CollectTimeout)
	defer cancel()

	logrus.WithField("peer", state.id).Debug("creating client")
	client, err := telemetry.NewClient(s.h, state.id)
	if err != nil {
		return err
	}
	defer client.Close()

	logrus.WithField("peer", state.id).Debug("getting session info")
	session, err := client.SessionInfo(ctx)
	if err != nil {
		return err
	}

	since := state.nextSeqN
	if session.Session != state.lastSession {
		since = 0
		state.lastSession = session.Session
	}

	stream := make(chan telemetry.DatapointStreamItem)
	go func() {
		for item := range stream {
			state.nextSeqN = item.NextSeqN
			logrus.WithField("peer", state.id).Debug("exporting ", len(item.Datapoints), " datapoint.")
			s.exporter.ExportDatapoints(state.id, session.Session, item.Datapoints)
		}
	}()

	err = client.Datapoints(ctx, since, stream)
	if err != nil {
		return err
	}
	close(stream)

	return nil
}

// must be holding the state's lock
func (s *Monitor) collectBandwidth(state *peerState) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.BandwidthTimeout)
	defer cancel()

	client, err := telemetry.NewClient(s.h, state.id)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.SessionInfo(ctx)
	if err != nil {
		return err
	}

	bandwidth, err := client.Bandwidth(ctx, telemetry.DEFAULT_PAYLOAD_SIZE)
	if err != nil {
		return err
	}

	s.exporter.ExportBandwidth(state.id, session.Session, bandwidth)

	return nil
}

// must be holding the state's lock
func (s *Monitor) peerError(state *peerState, err error) {
	logrus.WithField("peer", state.id).Debug("peer error: ", err)
	state.failedAttemps += 1
	if state.failedAttemps > s.opts.MaxFailedAttemps {
		s.caction <- actionqueue.After(&action{
			kind: ActionRemovePeer,
			pid:  state.id,
		}, s.opts.RetryInterval)
	}
}
