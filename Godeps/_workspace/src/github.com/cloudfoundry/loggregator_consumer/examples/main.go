package main

import (
	"crypto/tls"
	"fmt"
	consumer "github.com/cloudfoundry/loggregator_consumer"
)

var LoggregatorAddress = "loggregator.10.244.0.34.xip.io:443"
var appGuid = "<get your app guid and paste it here>"
var authToken = "<get your auth token and paste it here>"

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
