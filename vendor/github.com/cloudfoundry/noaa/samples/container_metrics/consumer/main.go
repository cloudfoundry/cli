package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
)

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	appId          = os.Getenv("APP_GUID")
	authToken      = os.Getenv("CF_ACCESS_TOKEN")
)

func main() {
	consumer := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	consumer.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming ContainerMetrics (will only succeed if you have admin credentials)")

	for {
		containerMetrics, err := consumer.ContainerMetrics(appId, authToken)

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
