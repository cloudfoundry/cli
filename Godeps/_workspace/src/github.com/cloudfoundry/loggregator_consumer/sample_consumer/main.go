package main

import (
	"crypto/tls"
	"fmt"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"os"
)

var LoggregatorAddress = "wss://loggregator.10.244.0.34.xip.io:443"
var appGuid = os.Getenv("APP_GUID")
var authToken = os.Getenv("CF_ACCESS_TOKEN")

func main() {
	connection := consumer.New(LoggregatorAddress, &tls.Config{InsecureSkipVerify: true}, nil)

	messages, err := connection.Recent(appGuid, authToken)

	if err != nil {
		fmt.Printf("===== Error getting recent messages: %v\n", err)
	} else {
		fmt.Println("===== Recent messages")
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}

	fmt.Println("===== Tailing messages")
	msgChan, err := connection.Tail(appGuid, authToken)

	if err != nil {
		fmt.Printf("===== Error tailing: %v\n", err)
	} else {
		for msg := range msgChan {
			fmt.Printf("%v \n", msg)
		}
	}
}
