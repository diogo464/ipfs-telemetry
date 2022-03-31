package main

import (
	"fmt"

	"git.d464.sh/adc/telemetry/plugin/traceroute"
)

func main() {
	result, err := traceroute.Trace("google.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Provider)
	fmt.Println(string(result.Output))
}
