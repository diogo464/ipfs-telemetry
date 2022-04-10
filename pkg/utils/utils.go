package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/multiformats/go-multiaddr"
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
	for _, addr := range in {
		for _, code := range []int{multiaddr.P_IP4, multiaddr.P_IP6} {
			if v, err := addr.ValueForProtocol(code); err == nil {
				ip := net.ParseIP(v)
				if ip == nil {
					continue
				}
				if ip.IsPrivate() {
					continue
				}
				return ip, nil
			}
		}
	}
	return nil, fmt.Errorf("no public address found")
}
