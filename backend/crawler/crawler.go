package crawler

import (
	"time"

	"github.com/diogo464/telemetry/walker"
)

const (
	SubjectCrawler = "crawler"
	StreamCrawler  = "crawler"

	KindPeer       = "peer"
	KindCrawlBegin = "crawl_begin"
	KindCrawlEnd   = "crawl_end"
)

type NatsMessage struct {
	Kind      string       `json:"kind"`
	Timestamp time.Time    `json:"timestamp"`
	Peer      *walker.Peer `json:"peer,omitempty"`
}
