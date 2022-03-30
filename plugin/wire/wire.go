package wire

import (
	"context"
	"fmt"
	"io"

	"git.d464.sh/adc/rle"
	"git.d464.sh/adc/telemetry/plugin/pb"
	"google.golang.org/protobuf/proto"
)

func ReadRequest(ctx context.Context, r io.Reader) (*pb.Request, error) {
	cerr := make(chan error)
	request := new(pb.Request)

	go func() {
		marshaled, err := rle.Read(r)
		if err != nil {
			cerr <- err
			return
		}
		if err := proto.Unmarshal(marshaled, request); err != nil {
			cerr <- err
			return
		}
		cerr <- nil
	}()

	var err error
	select {
	case err = <-cerr:
	case <-ctx.Done():
		err = ctx.Err()
	}

	if err == nil {
		return request, nil
	} else {
		return nil, err
	}
}

func WriteResponse(ctx context.Context, w io.Writer, resp *pb.Response) error {
	fmt.Println("Writting ", resp)
	cerr := make(chan error)

	go func() {
		marshaled, err := proto.Marshal(resp)
		if err != nil {
			cerr <- err
			return
		}
		if err := rle.Write(w, marshaled); err != nil {
			cerr <- err
			return
		}
		cerr <- nil
	}()

	select {
	case err := <-cerr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
