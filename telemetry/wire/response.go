package wire

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"github.com/google/uuid"
)

type ResponseType uint32

const (
	RESPONSE_SNAPSHOT ResponseType = iota
	RESPONSE_SYSTEM_INFO
)

type Response struct {
	Type ResponseType `json:"type"`
	Body interface{}  `json:"body"`
}

func newResponse(t ResponseType, b interface{}) *Response {
	return &Response{Type: t, Body: b}
}

type ResponseSnapshot struct {
	Session   uuid.UUID            `json:"session"`
	Snapshots []*snapshot.Snapshot `json:"snapshots"`
	NextSeqN  uint64               `json:"nextseqn"`
}

func NewResponseSnapshot(session uuid.UUID, snapshots []*snapshot.Snapshot, nextseqn uint64) *Response {
	return newResponse(RESPONSE_SNAPSHOT, &ResponseSnapshot{Session: session, Snapshots: snapshots, NextSeqN: nextseqn})
}

func (r *Response) GetSnapshot() (*ResponseSnapshot, error) {
	if v, ok := r.Body.(*ResponseSnapshot); ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("invalid body for request Since")
	}
}

type ResponseSystemInfo struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	NumCPU int    `json:"numcpus"`
}

func NewResponseSystemInfo() *Response {
	return newResponse(RESPONSE_SYSTEM_INFO, &ResponseSystemInfo{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		NumCPU: runtime.NumCPU(),
	})
}

func (r *Response) GetSystemInfo() (*ResponseSystemInfo, error) {
	if v, ok := r.Body.(*ResponseSystemInfo); ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("invalid body for request")
	}
}

func ReadResponse(ctx context.Context, r io.Reader) (*Response, error) {
	msg, err := read(ctx, r)
	if err != nil {
		return nil, err
	}

	response := &Response{Type: ResponseType(msg.Type), Body: nil}
	switch response.Type {
	case RESPONSE_SNAPSHOT:
		response.Body = new(ResponseSnapshot)
	case RESPONSE_SYSTEM_INFO:
		response.Body = new(ResponseSystemInfo)
	default:
		return nil, fmt.Errorf("invalid response type: %v", msg.Type)
	}

	if err := json.Unmarshal(msg.Body, response.Body); err != nil {
		return nil, err
	}

	return response, nil
}

func WriteResponse(ctx context.Context, w io.Writer, resp *Response) error {
	if data, err := json.Marshal(resp.Body); err == nil {
		return write(ctx, w, message{Type: uint32(resp.Type), Body: data})
	} else {
		return err
	}
}
