package wire

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"

	"git.d464.sh/adc/telemetry/telemetry/rle"
)

type message struct {
	Type uint32
	Body []byte
}

func write(ctx context.Context, w io.Writer, msg message) error {
	cerr := make(chan error)

	go func() {
		var bufwriter bytes.Buffer
		fwriter, err := flate.NewWriter(&bufwriter, -1)
		if err != nil {
			cerr <- err
			return
		}

		tyb := make([]byte, 4)
		binary.BigEndian.PutUint32(tyb, msg.Type)
		fwriter.Write(tyb)
		if err != nil {
			cerr <- err
			return
		}

		_, err = fwriter.Write(msg.Body)
		if err != nil {
			cerr <- err
			return
		}

		err = fwriter.Close()
		if err != nil {
			cerr <- err
			return
		}

		data := bufwriter.Bytes()
		err = rle.Write(w, data)
		if err != nil {
			cerr <- err
			return
		}
	}()

	select {
	case err := <-cerr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func read(ctx context.Context, r io.Reader) (message, error) {
	cmsg := make(chan message)
	cerr := make(chan error)

	go func() {
		compressed, err := rle.Read(r)
		if err != nil {
			cerr <- err
			return
		}

		bufreader := bytes.NewReader(compressed)
		freader := flate.NewReader(bufreader)

		tyb := make([]byte, 4)
		_, err = freader.Read(tyb)
		if err != nil {
			cerr <- err
			return
		}

		ty := binary.BigEndian.Uint32(tyb)
		log.Println("Reading message of type: ", ty)

		data, err := ioutil.ReadAll(freader)
		if err != nil {
			cerr <- err
			return
		}
		cmsg <- message{Type: ty, Body: data}
	}()

	select {
	case msg := <-cmsg:
		return msg, nil
	case err := <-cerr:
		return message{}, err
	case <-ctx.Done():
		return message{}, ctx.Err()
	}
}
