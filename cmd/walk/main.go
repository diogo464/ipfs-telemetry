package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"

	"github.com/diogo464/telemetry/walker"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var _ (walker.Observer) = (*observer)(nil)

type observer struct {
	initialized bool
	out         io.Writer
	q           chan interface{}
	done        chan struct{}
}

func newObserver(out io.Writer) *observer {
	o := &observer{
		out:  out,
		q:    make(chan interface{}, 100),
		done: make(chan struct{}),
	}
	go o.writeLoop()
	return o
}

func (o *observer) ObserveError(e *walker.Error) {
	o.queue(e)
}

func (o *observer) ObservePeer(p *walker.Peer) {
	o.queue(p)
}

func (o *observer) Close() {
	close(o.q)
	<-o.done
	if o.initialized {
		o.out.Write([]byte("]"))
	}
}

func (o *observer) queue(v interface{}) {
	o.q <- v
}

func (o *observer) writeLoop() {
	comma := []byte(",")
	writeComma := false
	for v := range o.q {
		if writeComma {
			_, err := o.out.Write(comma)
			if err != nil {
				panic(err)
			}
		}
		o.write(v)
		writeComma = true
	}
	o.done <- struct{}{}
}

func (o *observer) write(v interface{}) {
	if !o.initialized {
		o.initialized = true
		o.out.Write([]byte("["))
	}

	marshaled, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	o.out.Write(marshaled)
}

func main() {
	app := &cli.App{
		Name:        "walk",
		Description: "walk the DHT and dump all data",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_CONCURRENCY,
			FLAG_CONNECT_TIMEOUT,
			FLAG_REQUEST_TIMEOUT,
			FLAG_INTERVAL,
			FLAG_TCP,
			FLAG_UDP,
			FLAG_IPV4,
			FLAG_IPV6,
			FLAG_OUTPUT,
			FLAG_COMPRESS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	opts := make([]walker.Option, 0)

	if c.IsSet(FLAG_CONCURRENCY.Name) {
		opts = append(opts, walker.WithConcurrency(uint(c.Int(FLAG_CONCURRENCY.Name))))
	}

	if c.IsSet(FLAG_CONNECT_TIMEOUT.Name) {
		opts = append(opts, walker.WithConnectTimeout(c.Duration(FLAG_CONNECT_TIMEOUT.Name)))
	}

	if c.IsSet(FLAG_REQUEST_TIMEOUT.Name) {
		opts = append(opts, walker.WithRequestTimeout(c.Duration(FLAG_REQUEST_TIMEOUT.Name)))
	}

	if c.IsSet(FLAG_INTERVAL.Name) {
		opts = append(opts, walker.WithInterval(c.Duration(FLAG_INTERVAL.Name)))
	}

	allowTcp := c.Bool(FLAG_TCP.Name) || (!c.Bool(FLAG_TCP.Name) && !c.Bool(FLAG_UDP.Name))
	allowUdp := c.Bool(FLAG_UDP.Name) || (!c.Bool(FLAG_TCP.Name) && !c.Bool(FLAG_UDP.Name))
	allowIpv4 := c.Bool(FLAG_IPV4.Name) || (!c.Bool(FLAG_IPV4.Name) && !c.Bool(FLAG_IPV6.Name))
	allowIpv6 := c.Bool(FLAG_IPV6.Name) || (!c.Bool(FLAG_IPV4.Name) && !c.Bool(FLAG_IPV6.Name))
	opts = append(opts, walker.WithAddressFilter(func(m multiaddr.Multiaddr) bool {
		return ((walker.AddressFilterTcp(m) && allowTcp) ||
			(walker.AddressFilterUdp(m) && allowUdp) ||
			(walker.AddressFilterIpv4(m) && allowIpv4) ||
			(walker.AddressFilterIpv6(m) && allowIpv6)) && walker.AddressFilterPublic(m)
	}))

	var output io.WriteCloser = os.Stdout
	if c.IsSet(FLAG_OUTPUT.Name) {
		f, err := os.Create(c.String(FLAG_OUTPUT.Name))
		if err != nil {
			return err
		}
		defer f.Close()
		output = f
	}
	if c.Bool(FLAG_COMPRESS.Name) {
		output = gzip.NewWriter(output)
	}
	observer := newObserver(output)
	opts = append(opts, walker.WithObserver(observer))

	w, err := walker.New(opts...)
	if err != nil {
		return err
	}

	err = w.Walk(c.Context)
	if err != nil {
		return err
	}

	observer.Close()

	err = output.Close()
	if err != nil {
		return err
	}

	return nil
}
