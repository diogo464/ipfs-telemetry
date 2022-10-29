package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

type Number interface {
	int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64 | uintptr
}

func Max[T Number](v T, vs ...T) T {
	x := v
	for _, y := range vs {
		if y > x {
			x = y
		}
	}
	return x
}

func Min[T Number](v T, vs ...T) T {
	x := v
	for _, y := range vs {
		if y < x {
			x = y
		}
	}
	return x
}

func SliceAny[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

func ReadU32(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func WriteU32(w io.Writer, v uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	_, err := w.Write(buf)
	return err
}

func ReadU64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}

func WriteU64(w io.Writer, v uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	_, err := w.Write(buf)
	return err
}

func GetFirstPublicAddressFromMultiaddrs(in []multiaddr.Multiaddr) (net.IP, error) {
	public := GetPublicAddressesFromMultiaddrs(in)
	if len(public) == 0 {
		return nil, fmt.Errorf("no public address found")
	} else {
		return public[0], nil
	}
}

func GetPublicAddressesFromMultiaddrs(in []multiaddr.Multiaddr) []net.IP {
	public := make([]net.IP, 0)
	for _, addr := range in {
		for _, code := range []int{multiaddr.P_IP4, multiaddr.P_IP6} {
			if v, err := addr.ValueForProtocol(code); err == nil {
				ip := net.ParseIP(v)
				if ip == nil {
					continue
				}
				if ip.IsPrivate() || ip.IsLoopback() {
					continue
				}
				public = append(public, ip)
			}
		}
	}
	return public
}

func RandomCID() (cid.Cid, error) {
	buf := make([]byte, 128)
	_, err := rand.Read(buf)
	if err != nil {
		return cid.Cid{}, err
	}

	digest := sha256.Sum256(buf)
	hash, err := multihash.Encode(digest[:], multihash.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}

	return cid.NewCidV0(hash), nil
}

func RandomPeerID() peer.ID {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	var alg uint64 = multihash.SHA2_256
	hash, _ := multihash.Sum(b, alg, -1)
	return peer.ID(hash)
}
