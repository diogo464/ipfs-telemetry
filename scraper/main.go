package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"d464.sh/scraper/preimage"
	"d464.sh/scraper/scraper"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	identify_pb "github.com/libp2p/go-libp2p/p2p/protocol/identify/pb"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/multiformats/go-multiaddr"
)

const (
	SCRAPE_WORKERS  = 96
	SCRAPE_INTERVAL = time.Minute * 5
)

type PeerData struct {
	ID        peer.ID               `json:"id"`
	Addrs     []multiaddr.Multiaddr `json:"addresses"`
	Agent     string                `json:"agent"`
	Protocols []string              `json:"protocols"`
}

type CrawlState struct {
	lock sync.Mutex
	data []PeerData
}

func NewCrawlState() *CrawlState {
	return &CrawlState{
		data: []PeerData{},
	}
}

func (c *CrawlState) GetData() []PeerData {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.data
}

func (c *CrawlState) SetData(d []PeerData) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data = d
}

func main() {
	h := createHost()
	c := createScraper(h)
	s := NewCrawlState()

	go runCrawler(h, c, s)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := s.GetData()
		if serialized, err := json.Marshal(d); err == nil {
			w.Write(serialized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createHost() host.Host {
	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		panic(err)
	}
	return h
}

func createScraper(h host.Host) *scraper.Scraper {
	p := preimage.GenerateTable()
	c, err := scraper.NewScraper(h, p)
	if err != nil {
		panic(err)
	}
	return c
}

type evHandler struct {
	h    host.Host
	lock sync.Mutex
	data []PeerData
}

func (e *evHandler) ScrapeBegin(p peer.ID) {
	ident, err := runIdentify(e.h, p)
	if err != nil {
		return
	}

	addrsb := ident.GetListenAddrs()
	addrs := make([]multiaddr.Multiaddr, 0, len(addrsb))
	for _, addrb := range addrsb {
		addr, err := multiaddr.NewMultiaddrBytes(addrb)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}

	agent := ident.GetAgentVersion()
	protocols := ident.GetProtocols()

	e.lock.Lock()
	defer e.lock.Unlock()
	e.data = append(e.data, PeerData{
		ID:        p,
		Addrs:     addrs,
		Agent:     agent,
		Protocols: protocols,
	})
}
func (e *evHandler) ScrapeFailed(p peer.ID, err error)               {}
func (e *evHandler) ScrapeFinished(p peer.ID, addrs []peer.AddrInfo) {}

func runCrawler(h host.Host, c *scraper.Scraper, s *CrawlState) {
	seeds := getRandomSeeds(h)
	for {
		handler := &evHandler{h: h, data: []PeerData{}}
		start_time := time.Now()
		c.Run(context.Background(), seeds, handler, SCRAPE_WORKERS)
		elapsed := time.Since(start_time)
		fmt.Printf("Finished scrape with %v data in %v", len(handler.data), elapsed)
		s.SetData(handler.data)
		seeds = getRandomSeeds(h)
		time.Sleep(SCRAPE_INTERVAL)
	}
}

func getRandomSeeds(h host.Host) []peer.AddrInfo {
	seeds := make([]peer.AddrInfo, 0)
	ids := h.Peerstore().PeersWithAddrs()
	if len(ids) > 0 {
		for i := 0; i < 8; i++ {
			idx := rand.Intn(len(ids))
			id := ids[idx]
			info := h.Peerstore().PeerInfo(id)
			seeds = append(seeds, info)
		}
	}
	for _, info := range dht.GetDefaultBootstrapPeerAddrInfos() {
		seeds = append(seeds, info)
	}
	return seeds
}

func runIdentify(h host.Host, p peer.ID) (*identify_pb.Identify, error) {
	const CONNECT_TIMEOUT = time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), CONNECT_TIMEOUT)
	defer cancel()

	s, err := h.NewStream(network.WithUseTransient(ctx, "Identify"), p, identify.ID)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	r := protoio.NewDelimitedReader(s, 8*1024)
	msg := new(identify_pb.Identify)
	if err := r.ReadMsg(msg); err != nil {
		return nil, err
	}

	return msg, err
}
