package walker

import (
	"context"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/protoio"
)

type MessageSender struct {
	h host.Host
}

func NewMessageSender(h host.Host) *MessageSender {
	return &MessageSender{h}
}

func (ms *MessageSender) SendRequest(ctx context.Context, p peer.ID, pmes *pb.Message) (*pb.Message, error) {
	stream, err := ms.h.NewStream(ctx, p, dht.DefaultProtocols...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	writer := protoio.NewDelimitedWriter(stream)
	if err := writer.WriteMsg(pmes); err != nil {
		return nil, err
	}

	msg := new(pb.Message)
	reader := protoio.NewDelimitedReader(stream, network.MessageSizeMax)
	if err := reader.ReadMsg(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ms *MessageSender) SendMessage(ctx context.Context, p peer.ID, pmes *pb.Message) error {
	stream, err := ms.h.NewStream(ctx, p, dht.DefaultProtocols...)
	if err != nil {
		return err
	}
	defer stream.Close()

	writer := protoio.NewDelimitedWriter(stream)
	if err := writer.WriteMsg(pmes); err != nil {
		return err
	}

	return nil
}
