package telemetry

import "io"

var Int64PropertyDecoder = func(r io.Reader) (int64, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, io.ErrUnexpectedEOF
	}
	return Int64StreamDecoder(buf)
}

var StrPropertyDecoder = func(r io.Reader) (string, error) {
	v, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(v), nil
}
