package snapshot

import "time"

type Snapshot struct {
	Name string `json:"name"`
	// utc unix nano timestamp
	Time  uint64      `json:"stamp"`
	Value interface{} `json:"value"`
}

func NewSnapshot(name string, value interface{}) *Snapshot {
	return &Snapshot{
		Name:  name,
		Time:  uint64(time.Now().UTC().UnixNano()),
		Value: value,
	}
}
