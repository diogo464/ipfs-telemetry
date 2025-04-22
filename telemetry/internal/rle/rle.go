package rle

import (
	"encoding/binary"
	"io"
)

func Read(r io.Reader) ([]byte, error) {
	lb := make([]byte, 4)
	if _, err := io.ReadFull(r, lb); err != nil {
		return nil, err
	}

	l := int(binary.BigEndian.Uint32(lb))
	msg := make([]byte, l)
	if _, err := io.ReadFull(r, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func Write(w io.Writer, msg []byte) error {
	lb := make([]byte, 4)
	binary.BigEndian.PutUint32(lb, uint32(len(msg)))

	if _, err := w.Write(lb); err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	return nil
}
