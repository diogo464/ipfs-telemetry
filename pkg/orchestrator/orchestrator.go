package orchestrator

import (
	"context"
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/probe"
	"github.com/diogo464/ipfs_telemetry/pkg/utils"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type namedResult struct {
	probeName string
	result    *probe.ProbeResult
}

type OrchestratorServer struct {
	h    host.Host
	kad  *dht.IpfsDHT
	opts *options

	cids   []cid.Cid
	probes []*probe.Client

	cresult chan namedResult
}

func NewOrchestratorServer(ctx context.Context, o ...Option) (*OrchestratorServer, error) {
	opts := defaults()
	err := apply(opts, o...)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return nil, err
	}

	kad, err := createDHT(ctx, h)
	if err != nil {
		return nil, err
	}

	cids, err := createRandomCIDs(opts.numCids)
	if err != nil {
		return nil, err
	}

	probes := make([]*probe.Client, 0)
	for _, addr := range opts.probeAddrs {
		conn, err := grpc.Dial(addr.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		client := probe.NewClient(conn)
		probes = append(probes, client)
	}

	return &OrchestratorServer{
		h:    h,
		kad:  kad,
		opts: opts,

		cids:   cids,
		probes: probes,

		cresult: make(chan namedResult),
	}, nil
}

func (s *OrchestratorServer) Run(ctx context.Context) error {
	logrus.Debug("providing ", len(s.cids), " cids")
	for _, c := range s.cids {
		err := s.kad.Provide(ctx, c, false)
		if err != nil {
			return err
		}
	}

	logrus.Debug("setting up probes")
	for _, client := range s.probes {
		err := client.ProbeSetCids(ctx, s.cids)
		if err != nil {
			return err
		}
		name, err := client.GetName(ctx)
		if err != nil {
			return err
		}

		go func(c *probe.Client, n string) {
			cunamedresult := make(chan *probe.ProbeResult)
			go func() {
				for r := range cunamedresult {
					s.cresult <- namedResult{
						probeName: n,
						result:    r,
					}
				}
			}()
		LOOP:
			for {
				if err := c.ProbeResults(ctx, cunamedresult); err != nil {
					logrus.Error(err)
				}
				select {
				case <-ctx.Done():
					break LOOP
				default:
				}
				time.Sleep(time.Second)
			}
		}(client, name)
	}

	logrus.Debug("starting main loop")
	for {
		select {
		case r := <-s.cresult:
			s.receivedResult(r)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *OrchestratorServer) receivedResult(r namedResult) {
	logrus.WithFields(logrus.Fields{
		"start time": r.result.RequestStart,
		"duration":   r.result.RequestDuration,
		"error":      r.result.Error,
	}).Debug("received probe result")
	s.opts.exporter.Export(r.probeName, r.result)
}

func createDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	client := dht.NewDHTClient(ctx, h, datastore.NewMapDatastore())
	if err := client.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var err error = nil
	var success bool = false
	for _, bootstrap := range dht.GetDefaultBootstrapPeerAddrInfos() {
		err = h.Connect(ctx, bootstrap)
		if err == nil {
			success = true
		}
	}

	if success {
		client.RefreshRoutingTable()
		time.Sleep(time.Second * 2)
		return client, nil
	} else {
		return nil, err
	}
}

func createRandomCIDs(n int) ([]cid.Cid, error) {
	cids := make([]cid.Cid, 0, n)
	for i := 0; i < n; i++ {
		c, err := utils.RandomCID()
		if err != nil {
			return nil, err
		}
		cids = append(cids, c)
	}
	return cids, nil
}
