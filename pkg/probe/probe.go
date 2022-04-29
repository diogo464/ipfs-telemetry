package probe

import (
	"context"
	"sync"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/probe"
	"github.com/diogo464/telemetry/pkg/walker"
	"github.com/gammazero/deque"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dht_pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type probeSessionID uint64

type discoveredPeer struct {
	session  probeSessionID
	addrInfo peer.AddrInfo
}

type completedPeerProbe struct {
	session   probeSessionID
	result    *ProbeResult
	providers int
}

type sessionData struct {
	startTime      time.Time
	sessionCid     cid.Cid
	providersFound int
	probedPeers    map[peer.ID]struct{}
}

type ProbeServer struct {
	pb.UnimplementedProbeServer

	h    host.Host
	pm   *dht_pb.ProtocolMessenger
	opts *options

	cids       []cid.Cid
	cids_index int

	cdiscovered   chan discoveredPeer
	ccompleted    chan *completedPeerProbe
	nextSessionID probeSessionID
	sessions      map[probeSessionID]*sessionData
	ongoingProbes int
	pendingProbes *deque.Deque

	csetcids     chan []cid.Cid
	observers_mu sync.Mutex
	observers    map[chan<- *ProbeResult]struct{}
}

func NewProbeServer(o ...Option) (*ProbeServer, error) {
	opts := defaults()
	if err := apply(opts, o...); err != nil {
		return nil, err
	}

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return nil, err
	}

	sender := walker.NewMessageSender(h)
	pm, err := dht_pb.NewProtocolMessenger(sender)
	if err != nil {
		return nil, err
	}

	return &ProbeServer{
		h:    h,
		pm:   pm,
		opts: opts,

		cids:       make([]cid.Cid, 0),
		cids_index: 0,

		cdiscovered:   make(chan discoveredPeer),
		ccompleted:    make(chan *completedPeerProbe),
		nextSessionID: 0,
		sessions:      make(map[probeSessionID]*sessionData),
		ongoingProbes: 0,
		pendingProbes: deque.New(),

		csetcids:     make(chan []cid.Cid),
		observers_mu: sync.Mutex{},
		observers:    make(map[chan<- *ProbeResult]struct{}),
	}, nil
}

func (s *ProbeServer) Run(ctx context.Context) error {
	defer func() {
		s.observers_mu.Lock()
		for observer := range s.observers {
			close(observer)
		}
		s.observers_mu.Unlock()
	}()

	newSessionTimer := time.NewTicker(s.opts.probeNewSessionInterval)
	newProbeTimer := time.NewTicker(s.opts.probeNewProbeInterval)

	logrus.Debug("starting main loop")
	for {
		select {
		case a := <-s.cdiscovered:
			s.discoveredPeer(a.session, a.addrInfo)
		case p := <-s.ccompleted:
			s.completedProbe(ctx, p)
		case c := <-s.csetcids:
			logrus.Debug("setting ", len(c), " cids")
			startNewSession := len(s.cids) != 0
			s.setCids(c)
			if startNewSession {
				newSessionTimer.Reset(s.opts.probeNewSessionInterval)
				s.startNewProbeSession()
			}
		case <-newProbeTimer.C:
			s.newProbe(ctx)
		case <-newSessionTimer.C:
			if s.ongoingProbes < s.opts.probeMaxOngoing {
				s.startNewProbeSession()
			}
			s.deleteExpiredSessions()
		case <-ctx.Done():
			return ctx.Err()
		}

		OngoingProbes.Set(float64(s.ongoingProbes))
		Sessions.Set(float64(len(s.sessions)))
	}
}

func (s *ProbeServer) discoveredPeer(sessionID probeSessionID, addrInfo peer.AddrInfo) {
	s.pendingProbes.PushBack(&discoveredPeer{
		session:  sessionID,
		addrInfo: addrInfo,
	})
}

func (s *ProbeServer) newProbe(ctx context.Context) {
	for s.pendingProbes.Len() > 0 && s.ongoingProbes < s.opts.probeMaxOngoing {
		pending := s.pendingProbes.PopFront().(*discoveredPeer)
		if data, ok := s.sessions[pending.session]; ok {
			if _, ok := data.probedPeers[pending.addrInfo.ID]; !ok {
				TotalProbes.Inc()

				data.probedPeers[pending.addrInfo.ID] = struct{}{}
				s.ongoingProbes += 1
				go s.probePeerTask(ctx, pending.session, pending.addrInfo, data.sessionCid)
				break
			}
		}
	}
}

func (s *ProbeServer) probePeerTask(ctx context.Context, sessionID probeSessionID, addrInfo peer.AddrInfo, c cid.Cid) {
	s.h.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.TempAddrTTL)

	probe := &completedPeerProbe{
		session: sessionID,
		result: &ProbeResult{
			RequestStart:    time.Now(),
			RequestDuration: 0,
			Peer:            addrInfo.ID,
			Error:           nil,
		},
		providers: 0,
	}
	providers, closest, err := s.pm.GetProviders(ctx, addrInfo.ID, c.Hash())
	probe.result.RequestDuration = time.Since(probe.result.RequestStart)
	probe.providers = len(providers)
	if err != nil {
		probe.result.Error = err
		s.ccompleted <- probe
		return
	}

	for _, caddrInfo := range closest {
		s.cdiscovered <- discoveredPeer{
			session:  sessionID,
			addrInfo: *caddrInfo,
		}
	}

	s.ccompleted <- probe
}

func (s *ProbeServer) completedProbe(ctx context.Context, p *completedPeerProbe) {
	OngoingProbes.Dec()
	if p.result.Error == nil {
		SuccessfulProbes.Inc()
	} else {
		FailedProbes.Inc()
	}

	s.ongoingProbes -= 1
	if data, ok := s.sessions[p.session]; ok {
		data.providersFound += p.providers
		if data.providersFound >= s.opts.probeSessionProvidersLimit {
			logrus.WithFields(logrus.Fields{
				"session": p.session,
			}).Debug("session hit provider limit")
			delete(s.sessions, p.session)
		}
	}

	logrus.WithFields(logrus.Fields{
		"start time": p.result.RequestStart,
		"duration":   p.result.RequestDuration,
		"providers":  p.providers,
		"error":      p.result.Error,
	}).Debug("completed probe")

	s.observers_mu.Lock()
	for observer := range s.observers {
		observer <- p.result
	}
	s.observers_mu.Unlock()
}

func (s *ProbeServer) deleteExpiredSessions() {
	for k, v := range s.sessions {
		if time.Since(v.startTime) > s.opts.probeSessionLifetimeLimit {
			logrus.WithFields(logrus.Fields{
				"session": k,
			}).Debug("session expired")
			delete(s.sessions, k)
		}
	}
}

func (s *ProbeServer) setCids(cids []cid.Cid) {
	s.cids = cids
	s.cids_index = 0
}

func (s *ProbeServer) startNewProbeSession() {
	if len(s.cids) == 0 {
		return
	}
	probeCid := s.cids[s.cids_index]
	s.cids_index = (s.cids_index + 1) % len(s.cids)
	sessionID := s.nextSessionID
	s.nextSessionID += 1
	logrus.Debug("starting new probe session(", sessionID, ") with cid ", probeCid)

	s.sessions[sessionID] = &sessionData{
		startTime:      time.Now(),
		sessionCid:     probeCid,
		providersFound: 0,
		probedPeers:    make(map[peer.ID]struct{}),
	}

	go func() {
		for _, addrInfo := range dht.GetDefaultBootstrapPeerAddrInfos() {
			s.cdiscovered <- discoveredPeer{
				session:  sessionID,
				addrInfo: addrInfo,
			}
		}
	}()
}

// grpc

func (s *ProbeServer) ProbeGetName(context.Context, *emptypb.Empty) (*pb.ProbeGetNameResponse, error) {
	return &pb.ProbeGetNameResponse{Name: s.opts.probeName}, nil
}

func (s *ProbeServer) ProbeSetCids(ctx context.Context, req *pb.ProbeSetCidsRequest) (*emptypb.Empty, error) {
	cids := make([]cid.Cid, 0)
	for _, cidStr := range req.Cids {
		c, err := cid.Decode(cidStr)
		if err != nil {
			return nil, err
		}
		cids = append(cids, c)
	}
	select {
	case s.csetcids <- cids:
		return &emptypb.Empty{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *ProbeServer) ProbeResults(req *emptypb.Empty, stream pb.Probe_ProbeResultsServer) error {
	observer := make(chan *ProbeResult, 128)
	s.observers_mu.Lock()
	s.observers[observer] = struct{}{}
	s.observers_mu.Unlock()

	var err error = nil
	for result := range observer {
		errorStr := ""
		if result.Error != nil {
			errorStr = result.Error.Error()
		}
		err = stream.Send(&pb.ProbeResultResponse{
			RequestStart:    timestamppb.New(result.RequestStart),
			RequestDuration: durationpb.New(result.RequestDuration),
			Peer:            result.Peer.String(),
			Error:           errorStr,
		})

		if err != nil {
			break
		}
	}

	s.observers_mu.Lock()
	delete(s.observers, observer)
	s.observers_mu.Unlock()

	return err
}
