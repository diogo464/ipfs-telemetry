package snapshot

import "time"

var gSTARTUP_TIME time.Time

type Snapshot struct {
	Tag string `json:"name"`
	// utc unix nano timestamp
	Time   uint64      `json:"stamp"`
	Uptime uint64      `json:"uptime"`
	Value  interface{} `json:"value"`
}

func NewSnapshot(tag string, value interface{}) *Snapshot {
	return &Snapshot{
		Tag:    tag,
		Time:   uint64(time.Now().UTC().UnixNano()),
		Uptime: uint64(time.Since(gSTARTUP_TIME).Nanoseconds()),
		Value:  value,
	}
}

func init() {
	gSTARTUP_TIME = time.Now().UTC()
}
