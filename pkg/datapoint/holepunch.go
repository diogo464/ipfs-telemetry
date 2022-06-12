package datapoint

import (
	"time"
)

const HolePunchName = "holepunch"

type HolePunch struct {
	Timestamp time.Time `json:"timestamp"`
	Success   uint32    `json:"success"`
	Failure   uint32    `json:"failure"`
}
