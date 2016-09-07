package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/noaa/consumer"
)

var (
	dopplerAddress = os.Getenv("DOPPLER_ADDR")
	appGuid        = os.Getenv("APP_GUID")
	clientName     = os.Getenv("UAA_CLIENT_NAME")
	clientSecret   = os.Getenv("UAA_CLIENT_SECRET")
	uaaEndpoint    = os.Getenv("UAA_ENDPOINT")
)

func main() {

	uaa, err := uaago.NewClient(uaaEndpoint)
	if err != nil {
		fmt.Printf("Error from uaaClient %s\n", err)
		os.Exit(1)
	}

	refresher := tokenRefresher{uaaClient: uaa}

	consumer := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	consumer.RefreshTokenFrom(&refresher)
	consumer.SetDebugPrinter(ConsoleDebugPrinter{})

	fmt.Println("===== Streaming metrics")
	msgChan, errorChan := consumer.Firehose(appGuid, "")

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

type tokenRefresher struct {
	uaaClient *uaago.Client
}

func (t *tokenRefresher) RefreshAuthToken() (string, error) {
	token, err := t.uaaClient.GetAuthToken(clientName, clientSecret, true)
	if err != nil {
		return "", err
	}
	return token, nil
}
