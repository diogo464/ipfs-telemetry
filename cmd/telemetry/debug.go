package main

import (
	"fmt"
	"sort"

	"git.d464.sh/uni/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandDebug = &cli.Command{
	Name:   "debug",
	Action: actionDebug,
}

var _ sort.Interface = (*debugStreamsSort)(nil)

type debugStreamsSort struct {
	streams []telemetry.DebugStream
	sortBy  func(s1 telemetry.DebugStream, s2 telemetry.DebugStream) bool
}

// Len implements sort.Interface
func (s *debugStreamsSort) Len() int {
	return len(s.streams)
}

// Less implements sort.Interface
func (s *debugStreamsSort) Less(i int, j int) bool {
	return s.sortBy(s.streams[i], s.streams[j])
}

// Swap implements sort.Interface
func (s *debugStreamsSort) Swap(i int, j int) {
	s.streams[i], s.streams[j] = s.streams[j], s.streams[i]
}

func actionDebug(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	debug, err := client.Debug(c.Context)
	if err != nil {
		return err
	}

	sort.Sort(&debugStreamsSort{
		streams: debug.Streams,
		sortBy: func(s1, s2 telemetry.DebugStream) bool {
			return s1.Name < s2.Name
		},
	})

	var streamsUsed uint32 = 0
	var streamsTotal uint32 = 0

	for _, stream := range debug.Streams {
		fmt.Println(stream.Name)
		fmt.Println("\tUsed:", stream.UsedSize/1024, " KiB")
		fmt.Println("\tTotal:", stream.TotalSize/1024, " KiB")
		fmt.Printf("\tRatio: %.2f%%\n", 100.0*float64(stream.UsedSize)/float64(stream.TotalSize))
		streamsUsed += stream.UsedSize
		streamsTotal += stream.TotalSize
	}

	fmt.Println("Used:", streamsUsed/1024, " KiB")
	fmt.Println("Total:", streamsTotal/1024, " KiB")
	fmt.Printf("Ratio: %.2f%%\n", 100.0*float64(streamsUsed)/float64(streamsTotal))

	return nil
}
