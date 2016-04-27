package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/cloudfoundry/noaa/consumer"
)

const firehoseSubscriptionId = "firehose-a"

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	authToken      = os.Getenv("CF_ACCESS_TOKEN")
)

func main() {
	consumer := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	consumer.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming Firehose (will only succeed if you have admin credentials)")

	msgChan, errorChan := consumer.Firehose(firehoseSubscriptionId, authToken)
	go func() {
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
