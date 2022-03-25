package wire

import (
	"bytes"
	"compress/flate"
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

func write(w io.Writer, msg message) error {
	log.Println("Writting message of type: ", msg.Type)

	var bufwriter bytes.Buffer
	fwriter, err := flate.NewWriter(&bufwriter, -1)
	if err != nil {
		return err
	}

	tyb := make([]byte, 4)
	binary.BigEndian.PutUint32(tyb, msg.Type)
	fwriter.Write(tyb)
	if err != nil {
		return err
	}

	_, err = fwriter.Write(msg.Body)
	if err != nil {
		return err
	}

	err = fwriter.Close()
	if err != nil {
		return err
	}

	data := bufwriter.Bytes()
	err = rle.Write(w, data)
	if err != nil {
		return err
	}

	return nil
}

func read(r io.Reader) (message, error) {
	compressed, err := rle.Read(r)
	if err != nil {
		return message{}, err
	}

	bufreader := bytes.NewReader(compressed)
	freader := flate.NewReader(bufreader)

	tyb := make([]byte, 4)
	_, err = freader.Read(tyb)
	if err != nil {
		return message{}, err
	}

	ty := binary.BigEndian.Uint32(tyb)
	log.Println("Reading message of type: ", ty)

	data, err := ioutil.ReadAll(freader)
	if err != nil {
		return message{}, err
	}

	return message{Type: ty, Body: data}, nil
}
