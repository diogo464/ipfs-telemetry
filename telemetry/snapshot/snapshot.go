package snapshot

import (
	"encoding/json"
	"time"
)

var gSTARTUP_TIME time.Time

func init() {
	gSTARTUP_TIME = time.Now().UTC()
}

type Snapshot struct {
	Tag string
	// utc unix nano timestamp
	Time    uint64
	Uptime  uint64
	typed   interface{}
	untyped json.RawMessage
}

func NewSnapshot(tag string, value interface{}) *Snapshot {
	return &Snapshot{
		Tag:     tag,
		Time:    uint64(time.Now().UTC().UnixNano()),
		Uptime:  uint64(time.Since(gSTARTUP_TIME).Nanoseconds()),
		typed:   value,
		untyped: nil,
	}
}

func (s *Snapshot) decodeUnwrap(output interface{}) {
	if err := json.Unmarshal(s.untyped, output); err != nil {
		panic(err)
	}
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Tag string `json:"name"`
		// utc unix nano timestamp
		Time   uint64      `json:"stamp"`
		Uptime uint64      `json:"uptime"`
		Value  interface{} `json:"value"`
	}{
		Tag:    s.Tag,
		Time:   s.Time,
		Uptime: s.Uptime,
		Value:  s.typed,
	})
}

func (s *Snapshot) UnmarshalJSON(data []byte) error {
	type WireSnapshot struct {
		Tag string `json:"name"`
		// utc unix nano timestamp
		Time   uint64          `json:"stamp"`
		Uptime uint64          `json:"uptime"`
		Value  json.RawMessage `json:"value"`
	}

	w := WireSnapshot{}
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}

	s.Tag = w.Tag
	s.Time = w.Time
	s.Uptime = w.Uptime
	s.untyped = w.Value

	return nil
}
