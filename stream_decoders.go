package telemetry

import (
	"encoding/binary"
	"io"
)

var Int64StreamDecoder = func(b []byte) (int64, error) {
	if len(b) < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}
