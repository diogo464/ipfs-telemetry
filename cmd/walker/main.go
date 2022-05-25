package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/diogo464/telemetry/pkg/walker"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var FLAG_OUTPUT = &cli.StringFlag{
	Name:    "output",
	Usage:   "output file",
	EnvVars: []string{"WALKER_OUTPUT"},
	Value:   "dht.json",
}

var FLAG_CONNECT_TIMEOUT = &cli.IntFlag{
	Name:    "connect-timeout",
	Usage:   "timeout, in seconds, to establish a connection",
	EnvVars: []string{"WALKER_CONNECT_TIMEOUT"},
}

var FLAG_REQUEST_TIMEOUT = &cli.IntFlag{
	Name:    "request-timeout",
	Usage:   "timeout, in seconds, for a peer request",
	EnvVars: []string{"WALKER_REQUEST_TIMEOUT"},
}

var FLAG_INTERVAL = &cli.IntFlag{
	Name:    "interval",
	Usage:   "interval, in milliseconds, between each new connection",
	EnvVars: []string{"WALKER_INTERVAL"},
}

var FLAG_CONCURRENCY = &cli.IntFlag{
	Name:    "concurrency",
	Usage:   "maximum number of parallel requests",
	EnvVars: []string{"WALKER_CONCURRENCY"},
}

var _ walker.Observer = (*statusCollector)(nil)

type statusCollector struct {
	mu      sync.Mutex
	success int
	error   int
}

// ObserveError implements walker.Observer
func (c *statusCollector) ObserveError(*walker.Error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.error += 1
	c.printStatus()
}

// ObservePeer implements walker.Observer
func (c *statusCollector) ObservePeer(*walker.Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.success += 1
	c.printStatus()
}

func (c *statusCollector) printStatus() {
	fmt.Print("success/error : ", c.success, "/", c.error, "\r")
}

var _ walker.Observer = (*fileCollector)(nil)

type fileCollector struct {
	mu   sync.Mutex
	file *os.File
}

// ObserveError implements walker.Observer
func (c *fileCollector) ObserveError(e *walker.Error) {
	c.writeData(struct {
		ID        peer.ID               `json:"id"`
		Addresses []multiaddr.Multiaddr `json:"addresses"`
		Time      time.Time             `json:"time"`
		Err       string                `json:"error"`
	}{
		ID:        e.ID,
		Addresses: e.Addresses,
		Time:      e.Time,
		Err:       e.Err.Error(),
	})
}

// ObservePeer implements walker.Observer
func (c *fileCollector) ObservePeer(p *walker.Peer) {
	c.writeData(p)
}

func (c *fileCollector) writeData(d any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, err := json.Marshal(d); err == nil {
		data = append(data, '\n')
		if _, err := c.file.Write(data); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func main() {
	app := &cli.App{
		Name:        "walker",
		Description: "walk the dht",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_OUTPUT,
			FLAG_CONNECT_TIMEOUT,
			FLAG_REQUEST_TIMEOUT,
			FLAG_INTERVAL,
			FLAG_CONCURRENCY,
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}

	opts := make([]walker.Option, 0)

	file, err := os.Create(c.String(FLAG_OUTPUT.Name))
	if err != nil {
		return err
	}
	defer file.Close()

	fileCollector := &fileCollector{file: file}
	statusCollector := &statusCollector{}
	opts = append(opts,
		walker.WithObserver(
			walker.NewMultiObserver(
				fileCollector,
				statusCollector,
			),
		),
	)

	if c.IsSet(FLAG_CONNECT_TIMEOUT.Name) {
		dur := time.Duration(int64(time.Second) * int64(c.Int(FLAG_CONNECT_TIMEOUT.Name)))
		opts = append(opts, walker.WithConnectTimeout(dur))
	}

	if c.IsSet(FLAG_REQUEST_TIMEOUT.Name) {
		dur := time.Duration(int64(time.Second) * int64(c.Int(FLAG_REQUEST_TIMEOUT.Name)))
		opts = append(opts, walker.WithRequestTimeout(dur))
	}

	if c.IsSet(FLAG_INTERVAL.Name) {
		interval := time.Duration(int64(time.Millisecond) * int64(c.Int(FLAG_INTERVAL.Name)))
		opts = append(opts, walker.WithInterval(interval))
	}

	if c.IsSet(FLAG_CONCURRENCY.Name) {
		opts = append(opts, walker.WithConcurrency(uint(c.Int(FLAG_CONCURRENCY.Name))))
	}

	walker, err := walker.New(h, opts...)
	if err != nil {
		return err
	}

	err = walker.Walk(c.Context)
	if err != nil {
		return err
	}

	fmt.Println("success: ", statusCollector.success, ", error: ", statusCollector.error)

	return nil
}
