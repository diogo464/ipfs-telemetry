package wire

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type RequestType uint32

const (
	REQUEST_SNAPSHOT RequestType = iota
	REQUEST_SYSTEM_INFO
	REQUEST_BANDWDITH_DOWNLOAD
	REQUEST_BANDWDITH_UPLOAD
)

type Request struct {
	Type RequestType `json:"type"`
	Body interface{} `json:"body"`
}

func newRequest(t RequestType, b interface{}) *Request {
	return &Request{Type: t, Body: b}
}

type RequestSnapshot struct {
	Session uuid.UUID `json:"session"`
	Since   uint64    `json:"since"`
}

func NewRequestSnapshot(session uuid.UUID, since uint64) *Request {
	return newRequest(REQUEST_SNAPSHOT, &RequestSnapshot{
		Session: session,
		Since:   since,
	})
}

func NewRequestBandwdithDownload() *Request {
	return newRequest(REQUEST_BANDWDITH_DOWNLOAD, nil)
}

func NewRequestBandwdithUpload() *Request {
	return newRequest(REQUEST_BANDWDITH_UPLOAD, nil)
}

func NewRequestSystemInfo() *Request {
	return newRequest(REQUEST_SYSTEM_INFO, nil)
}

func (r *Request) GetSince() *RequestSnapshot {
	return r.Body.(*RequestSnapshot)
}

func ReadRequest(ctx context.Context, r io.Reader) (*Request, error) {
	msg, err := read(ctx, r)
	if err != nil {
		return nil, err
	}

	request := &Request{Type: RequestType(msg.Type), Body: nil}
	switch request.Type {
	case REQUEST_SNAPSHOT:
		request.Body = new(RequestSnapshot)
	case REQUEST_SYSTEM_INFO:
		request.Body = nil
	case REQUEST_BANDWDITH_DOWNLOAD:
		request.Body = nil
	case REQUEST_BANDWDITH_UPLOAD:
		request.Body = nil
	default:
		return nil, fmt.Errorf("invalid request type: %v", msg.Type)
	}

	if request.Body != nil {
		if err := json.Unmarshal(msg.Body, request.Body); err != nil {
			return nil, err
		}
	}

	return request, nil
}

func WriteRequest(ctx context.Context, w io.Writer, req *Request) error {
	if data, err := json.Marshal(req.Body); err == nil {
		return write(ctx, w, message{
			Type: uint32(req.Type),
			Body: data,
		})
	} else {
		return err
	}
}
