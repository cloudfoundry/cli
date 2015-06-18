package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

var dopplerAddress = os.Getenv("DOPPLER_ADDR")
var appGuid = os.Getenv("APP_GUID")
var authToken = os.Getenv("CF_ACCESS_TOKEN")

func main() {
	connection := noaa.NewConsumer(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	connection.SetDebugPrinter(ConsoleDebugPrinter{})

	messages, err := connection.RecentLogs(appGuid, authToken)

	if err != nil {
		fmt.Printf("===== Error getting recent messages: %v\n", err)
	} else {
		fmt.Println("===== Recent logs")
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}

	fmt.Println("===== Streaming metrics")
	msgChan := make(chan *events.Envelope)
	go func() {
		defer close(msgChan)
		errorChan := make(chan error)
		go connection.Stream(appGuid, authToken, msgChan, errorChan)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	for msg := range msgChan {
		fmt.Printf("%v \n", msg)
	}
}

type ConsoleDebugPrinter struct{}

func (c ConsoleDebugPrinter) Print(title, dump string) {
	println(title)
	println(dump)
}
