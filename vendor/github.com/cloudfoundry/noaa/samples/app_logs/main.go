package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/cloudfoundry/noaa/consumer"
)

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	appGuid        = os.Getenv("APP_GUID")
	authToken      = os.Getenv("CF_ACCESS_TOKEN")
)

func main() {
	consumer := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	consumer.SetDebugPrinter(ConsoleDebugPrinter{})

	messages, err := consumer.RecentLogs(appGuid, authToken)

	if err != nil {
		fmt.Printf("===== Error getting recent messages: %v\n", err)
	} else {
		fmt.Println("===== Recent logs")
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}

	fmt.Println("===== Streaming metrics")
	msgChan, errorChan := consumer.Stream(appGuid, authToken)

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
