package telemetry

import (
	"net"

	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var _ (net.Listener) = (*serviceListener)(nil)

type serviceListener struct {
	listener   net.Listener
	serviceAcl *serviceAccessControl
}

func newServiceListener(h host.Host, tag protocol.ID, serviceAcl *serviceAccessControl) (*serviceListener, error) {
	listener, err := gostream.Listen(h, tag)
	if err != nil {
		return nil, err
	}

	return &serviceListener{
		listener:   listener,
		serviceAcl: serviceAcl,
	}, nil
}

// Accept implements net.Listener
func (l *serviceListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return nil, err
		}

		idStr := conn.RemoteAddr().String()
		id, err := peer.Decode(idStr)
		if err != nil {
			return nil, err
		}

		if !l.serviceAcl.isAllowed(id) {
			conn.Close()
			continue
		}

		return conn, nil
	}
}

// Addr implements net.Listener
func (l *serviceListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close implements net.Listener
func (l *serviceListener) Close() error {
	return l.listener.Close()
}
