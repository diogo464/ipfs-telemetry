package cli

import (
	"encoding/json"
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/urfave/cli/v2"
)

// Display this stream segment
type StreamDecoder func([]byte) (interface{}, error)

var FLAGS = []cli.Flag{
	fLAG_CONN_TYPE,
	fLAG_HOST,
}

var COMMANDS = []*cli.Command{
	commandUpload,
	commandDownload,
	commandDebug,
	commandSessionInfo,
	commandSystemInfo,
	commandStream,
	commandStreams,
	commandProperty,
	commandProperties,
	commandWalk,
	//CommandWatch,
}

var registeredStreamDecoders map[string]StreamDecoder = make(map[string]StreamDecoder)

func RegisterStreamDecoder[T any](streamName string, decoder telemetry.StreamDecoder[T]) {
	registeredStreamDecoders[streamName] = func(b []byte) (interface{}, error) {
		return decoder(b)
	}
}

func clientFromContext(c *cli.Context) (*telemetry.Client, error) {
	switch c.String(fLAG_CONN_TYPE.Name) {
	case "libp2p":

		info, err := peer.AddrInfoFromString(c.String(fLAG_HOST.Name))
		if err != nil {
			return nil, err
		}
		h, err := libp2p.New(libp2p.NoListenAddrs)
		if err != nil {
			return nil, err
		}
		fmt.Println(info)
		h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		return telemetry.NewClient(h, info.ID)
	case "tcp":
		return telemetry.NewClient2(c.String(fLAG_HOST.Name))
	default:
		return nil, fmt.Errorf("Unknown connection type: %s", c.String(fLAG_CONN_TYPE.Name))
	}
}

func printAsJson(v interface{}) {
	marshaled, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshaled))
}
