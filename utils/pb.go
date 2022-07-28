package utils

import (
	"io"
	"time"

	"github.com/diogo464/telemetry/rle"
	"github.com/gogo/protobuf/types"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
)

func TimeFromPB(in *types.Timestamp) time.Time {
	return time.Unix(in.Seconds, int64(in.Nanos))
}

func TimeToPB(in *time.Time) *types.Timestamp {
	return &types.Timestamp{
		Seconds: in.Unix(),
		Nanos:   int32(in.Nanosecond()),
	}
}

func DurationFromPB(in *types.Duration) time.Duration {
	return time.Duration(in.Seconds*1_000_000_000 + int64(in.Nanos))
}

func DurationToPB(in *time.Duration) *types.Duration {
	return &types.Duration{
		Seconds: in.Nanoseconds() / 1_000_000_000,
		Nanos:   int32(in.Nanoseconds() % 1_000_000_000),
	}
}

func DurationArrayToPB(in []time.Duration) []*types.Duration {
	out := make([]*types.Duration, 0, len(in))
	for _, dur := range in {
		out = append(out, DurationToPB(&dur))
	}
	return out
}

func MultiAddrsToPB(in []multiaddr.Multiaddr) []string {
	addrs := make([]string, 0, len(in))
	for _, a := range in {
		addrs = append(addrs, a.String())
	}
	return addrs
}

func ReadRle(r io.Reader, v proto.Message) error {
	data, err := rle.Read(r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, v)
}
