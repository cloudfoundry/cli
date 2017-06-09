package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

const firehoseSubscriptionId = "firehose-a"

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	authToken      = os.Getenv("CF_ACCESS_TOKEN")
)

func main() {
	filterType := flag.String("filter", "all", "filter messages by 'logs' or 'metrics' (default: all)")
	flag.Parse()

	cnsmr := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	cnsmr.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming Firehose (will only succeed if you have admin credentials)")

	var (
		msgChan   <-chan *events.Envelope
		errorChan <-chan error
	)

	switch *filterType {
	case "logs":
		msgChan, errorChan = cnsmr.FilteredFirehose(firehoseSubscriptionId, authToken, consumer.LogMessages)
	case "metrics":
		msgChan, errorChan = cnsmr.FilteredFirehose(firehoseSubscriptionId, authToken, consumer.Metrics)
	default:
		msgChan, errorChan = cnsmr.Firehose(firehoseSubscriptionId, authToken)
	}

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
