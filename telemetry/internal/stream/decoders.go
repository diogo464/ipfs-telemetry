package stream

import (
	"encoding/binary"
	"io"
	"math"
)

var Int64Decoder = func(b []byte) (int64, error) {
	if len(b) < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

var Float64Decoder = func(b []byte) (float64, error) {
	if len(b) < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
}

var StringDecoder = func(b []byte) (string, error) {
	return string(b), nil
}

var ByteDecoder = func(b []byte) ([]byte, error) { return b, nil }
