package main

import (
	"crypto/tls"
	"fmt"
	consumer "github.com/cloudfoundry/loggregator_consumer/dropsonde_consumer"
	"os"
)

var DopplerAddress = "wss://doppler.10.244.0.34.xip.io:443"
var appGuid = os.Getenv("APP_GUID")
var authToken = os.Getenv("CF_ACCESS_TOKEN")

func main() {
	connection := consumer.NewDropsondeConsumer(DopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)

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
	msgChan, err := connection.Stream(appGuid, authToken)

	if err != nil {
		fmt.Printf("===== Error streaming: %v\n", err)
	} else {
		for msg := range msgChan {
			fmt.Printf("%v \n", msg)
		}
	}
}
