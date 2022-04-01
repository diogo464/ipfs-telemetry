package telemetry

import (
	"context"
	"fmt"
	"net"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ERR_INVALID_RESPONSE = fmt.Errorf("invalid response")

type Client struct {
	h host.Host
	p peer.ID
	// session uuid
	s uuid.NullUUID
	// sequence number for snapshots
	n uint64
}

func NewClient(h host.Host, p peer.ID) *Client {
	return &Client{h: h, p: p, s: uuid.NullUUID{}, n: 0}
}

func (c *Client) Session() uuid.NullUUID {
	return c.s
}

func (c *Client) Snapshots(ctx context.Context) ([]snapshot.Snapshot, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSnapshots(ctx, &pb.GetSnapshotsRequest{
		Session: c.s.UUID.String(),
		Since:   c.n,
	})
	if err != nil {
		return nil, err
	}

	session, err := uuid.Parse(response.Session)
	if err != nil {
		return nil, err
	}

	snapshots := make([]snapshot.Snapshot, 0)
	for _, v := range response.GetSet().GetPings() {
		s, err := snapshot.PingFromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}

	for _, v := range response.GetSet().GetNetworks() {
		s, err := snapshot.NetworkFromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}

	for _, v := range response.GetSet().GetRoutingTables() {
		s, err := snapshot.RoutingTableFromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}

	c.s.UUID = session
	c.s.Valid = true
	c.n = response.Next

	return snapshots, nil
}

func (c *Client) SystemInfo(ctx context.Context) (*SystemInfo, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSystemInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		OS:     response.Os,
		Arch:   response.Arch,
		NumCPU: response.Numcpu,
	}, nil
}

func (c *Client) Download(ctx context.Context) (uint64, error) {
	panic("todo")
	//stream, err := c.newStream(ctx)
	//if err != nil {
	//	return 0, err
	//}
	//defer stream.Close()

	//request := pb.Request{
	//	Body: &pb.Request_BandwidthDownload_{
	//		BandwidthDownload: &pb.Request_BandwidthDownload{},
	//	},
	//}
	//if err := writeRequest(ctx, stream, &request); err != nil {
	//	return 0, err
	//}

	//write_start := time.Now()
	//n, err := io.Copy(stream, io.LimitReader(utils.NullReader{}, BANDWIDTH_PAYLOAD_SIZE))
	//if err != nil {
	//	return 0, err
	//}

	//buf := make([]byte, 1)
	//stream.Read(buf)
	//elapsed := time.Since(write_start)

	//rate := uint64(float64(n) / elapsed.Seconds())
	//return rate, nil
}

func (c *Client) Upload(ctx context.Context) (uint64, error) {
	panic("todo")
	//stream, err := c.newStream(ctx)
	//if err != nil {
	//	return 0, err
	//}
	//defer stream.Close()

	//request := pb.Request{
	//	Body: &pb.Request_BandwidthUpload_{
	//		BandwidthUpload: &pb.Request_BandwidthUpload{},
	//	},
	//}
	//if err := writeRequest(ctx, stream, &request); err != nil {
	//	return 0, err
	//}

	//buf := make([]byte, 32*1024)
	//n := 0
	//var read_start *time.Time = nil
	//for {
	//	x, err := stream.Read(buf)
	//	n += x
	//	if read_start == nil {
	//		tm := time.Now()
	//		read_start = &tm
	//	}
	//	if err != nil && err != io.EOF {
	//		return 0, err
	//	}
	//	if err == io.EOF {
	//		break
	//	}
	//}
	////n, err := io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
	//if err != nil {
	//	return 0, err
	//}
	//elapsed := time.Since(*read_start)

	//rate := uint64(float64(n) / elapsed.Seconds())
	//return rate, nil
}

func (c *Client) newStream(ctx context.Context) (network.Stream, error) {
	return c.h.NewStream(ctx, c.p, ID)
}

func (c *Client) newGrpcClient() (pb.ClientClient, error) {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
			return gostream.Dial(context.Background(), c.h, c.p, ID)
		}))
	if err != nil {
		return nil, err
	}
	return pb.NewClientClient(conn), nil
}
