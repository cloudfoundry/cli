package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"
	"github.com/cloudfoundry/noaa"
)

const DopplerAddress = "wss://doppler.10.244.0.34.xip.io:443"
var appId = "60a13b0f-fce7-4c02-b92a-d43d583877ed"

var authToken = os.Getenv("CF_ACCESS_TOKEN")

func main() {
	connection := noaa.NewConsumer(DopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	connection.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming ContainerMetrics (will only succeed if you have admin credentials)")


	for {
		containerMetrics, err := connection.ContainerMetrics(appId, authToken)

		for _, cm := range containerMetrics {
			fmt.Printf("%v \n", cm)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}

		time.Sleep(3 * time.Second)
	}

}

type ConsoleDebugPrinter struct{}

func (c ConsoleDebugPrinter) Print(title, dump string) {
	println(title)
	println(dump)
}
