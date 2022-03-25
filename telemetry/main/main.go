package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"d464.sh/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <remote addr> <command> <args>\n", os.Args[0])
		return
	}

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		panic(err)
	}
	addr, err := peer.AddrInfoFromString(os.Args[1])
	if err != nil {
		panic(err)
	}
	h.Peerstore().AddAddrs(addr.ID, addr.Addrs, time.Hour)

	client := telemetry.NewTelemetryClient(h, addr.ID)

	cmd := os.Args[2]
	args := os.Args[3:]

	switch cmd {
	case "since":
		command_since(client, args)
	case "download":
		command_download(client, args)
	case "upload":
		command_upload(client, args)
	case "systeminfo":
		command_system_info(client, args)
	default:
		fmt.Println("invalid command: ", cmd)
	}
}

func command_since(client *telemetry.TelemetryClient, args []string) {
	since, err := strconv.ParseUint(args[0], 10, 8)
	if err != nil {
		panic(err)
	}

	resp, err := client.Snapshots(context.Background(), since)
	if err != nil {
		panic(err)
	}

	for _, s := range resp.Snapshots {
		fmt.Println("Name = ", s.Name)
		fmt.Println("Value = ", s.Value)
	}

	fmt.Println(resp)
}

func command_download(client *telemetry.TelemetryClient, args []string) {
	rate, err := client.Download(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Peer Download Rate = %v byte/s\n", rate)
	fmt.Printf("Peer Download Rate = %v MB/s\n", float64(rate)/(1024.0*1024.0))
}

func command_upload(client *telemetry.TelemetryClient, args []string) {
	rate, err := client.Upload(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Peer Upload Rate = %v byte/s\n", rate)
	fmt.Printf("Peer Upload Rate = %v MB/s\n", float64(rate)/(1024.0*1024.0))
}

func command_system_info(client *telemetry.TelemetryClient, args []string) {
	info, err := client.SystemInfo(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}
