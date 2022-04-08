package utils

import (
	"encoding/binary"
	"io"
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
