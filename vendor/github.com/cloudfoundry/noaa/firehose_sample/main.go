package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

var dopplerAddress = os.Getenv("DOPPLER_ADDR")
var authToken = os.Getenv("CF_ACCESS_TOKEN")

const firehoseSubscriptionId = "firehose-a"

func main() {
	connection := noaa.NewConsumer(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	connection.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming Firehose (will only succeed if you have admin credentials)")

	msgChan := make(chan *events.Envelope)
	go func() {
		defer close(msgChan)
		errorChan := make(chan error)
		go connection.Firehose(firehoseSubscriptionId, authToken, msgChan, errorChan)

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
