package telemetry

import (
	"context"
	"fmt"
	"io"
	"time"

	"git.d464.sh/adc/rle"
	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/proto"
)

const (
	ID                     = "/telemetry/telemetry/0.0.0"
	BANDWIDTH_PAYLOAD_SIZE = 32 * 1024 * 1024
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

func (c *Client) Snapshots(ctx context.Context) ([]Snapshot, error) {
	request := pb.Request{
		Body: &pb.Request_Snapshots_{
			Snapshots: &pb.Request_Snapshots{
				Session: c.Session().UUID.String(),
				Since:   c.n,
				Tags:    []string{},
			},
		},
	}

	fmt.Println("sending request")
	response, err := c.sendRequestReadResponse(ctx, &request)
	if err != nil {
		return nil, err
	}

	fmt.Println("getting snapshots")
	snapshots := response.GetSnapshots()
	if snapshots == nil {
		return nil, ERR_INVALID_RESPONSE
	}

	session, err := uuid.Parse(response.GetSession())
	if err != nil {
		return nil, ERR_INVALID_RESPONSE
	}

	c.s = uuid.NullUUID{
		UUID:  session,
		Valid: true,
	}
	c.n = snapshots.Next
	arr := make([]Snapshot, 0, len(snapshots.Snapshots))
	for _, s := range snapshots.Snapshots {
		snapshot, err := snapshotFromPB(s)
		if err != nil {
			return nil, err
		}
		arr = append(arr, snapshot)
	}

	return arr, nil
}

func (c *Client) SystemInfo(ctx context.Context) (*SystemInfo, error) {
	request := pb.Request{
		Body: &pb.Request_SystemInfo_{
			SystemInfo: &pb.Request_SystemInfo{},
		},
	}

	response, err := c.sendRequestReadResponse(ctx, &request)
	if err != nil {
		return nil, err
	}

	pbinfo := response.GetSystemInfo()
	if pbinfo == nil {
		return nil, ERR_INVALID_RESPONSE
	}

	return &SystemInfo{
		OS:     pbinfo.Os,
		Arch:   pbinfo.Arch,
		NumCPU: pbinfo.Numcpu,
	}, nil
}

func (c *Client) Download(ctx context.Context) (uint64, error) {
	stream, err := c.newStream(ctx)
	if err != nil {
		return 0, err
	}
	defer stream.Close()

	request := pb.Request{
		Body: &pb.Request_BandwidthDownload_{
			BandwidthDownload: &pb.Request_BandwidthDownload{},
		},
	}
	if err := writeRequest(ctx, stream, &request); err != nil {
		return 0, err
	}

	write_start := time.Now()
	n, err := io.Copy(stream, io.LimitReader(utils.NullReader{}, BANDWIDTH_PAYLOAD_SIZE))
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1)
	stream.Read(buf)
	elapsed := time.Since(write_start)

	rate := uint64(float64(n) / elapsed.Seconds())
	return rate, nil
}

func (c *Client) Upload(ctx context.Context) (uint64, error) {
	stream, err := c.newStream(ctx)
	if err != nil {
		return 0, err
	}
	defer stream.Close()

	request := pb.Request{
		Body: &pb.Request_BandwidthUpload_{
			BandwidthUpload: &pb.Request_BandwidthUpload{},
		},
	}
	if err := writeRequest(ctx, stream, &request); err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024)
	n := 0
	var read_start *time.Time = nil
	for {
		x, err := stream.Read(buf)
		n += x
		if read_start == nil {
			tm := time.Now()
			read_start = &tm
		}
		if err != nil && err != io.EOF {
			return 0, err
		}
		if err == io.EOF {
			break
		}
	}
	//n, err := io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(*read_start)

	rate := uint64(float64(n) / elapsed.Seconds())
	return rate, nil
}

func (c *Client) newStream(ctx context.Context) (network.Stream, error) {
	return c.h.NewStream(ctx, c.p, ID)
}

func (c *Client) sendRequestReadResponse(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	stream, err := c.newStream(ctx)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	fmt.Println("writting request")
	if err := writeRequest(ctx, stream, request); err != nil {
		return nil, err
	}
	fmt.Println("request written")

	fmt.Println("reading response")
	response, err := readResponse(ctx, stream)
	if err != nil {
		return nil, err
	}
	fmt.Println("response read")

	return response, nil
}

func readResponse(ctx context.Context, r io.Reader) (*pb.Response, error) {
	// TODO: context
	marshaled, err := rle.Read(r)
	if err != nil {
		return nil, err
	}
	response := new(pb.Response)
	if err := proto.Unmarshal(marshaled, response); err != nil {
		return nil, err
	}
	return response, nil
}

func writeRequest(ctx context.Context, w io.Writer, request *pb.Request) error {
	marshaled, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	// TODO: context
	return rle.Write(w, marshaled)
}
