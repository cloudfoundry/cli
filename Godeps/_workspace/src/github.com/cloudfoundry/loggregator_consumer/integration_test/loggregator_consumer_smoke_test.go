package integration_test

import (
	"crypto/tls"
	"encoding/json"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"

	"io/ioutil"
	"strings"
)

var _ = Describe("LoggregatorConsumer:", func() {
	var appGuid, authToken string
	var connection consumer.LoggregatorConsumer

	BeforeEach(func() {
		var err error
		appGuid = os.Getenv("TEST_APP_GUID")
		loggregatorEndpoint := os.Getenv("LOGGREGATOR_ENDPOINT")

		connection = consumer.New(loggregatorEndpoint, &tls.Config{InsecureSkipVerify: true}, nil)
		authToken, err = getAuthToken()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		connection.Close()
	})

	It("should return data for recent", func() {
		messages, err := connection.Recent(appGuid, authToken)
		Expect(err).NotTo(HaveOccurred())
		Expect(messages).To(ContainElement(ContainSubstring("Tick")))
	})

	It("should return data for tail", func(done Done) {
		messagesChan, err := connection.Tail(appGuid, authToken)
		Expect(err).NotTo(HaveOccurred())

		for m := range messagesChan {
			if strings.Contains(string(m.GetMessage()), "Tick") {
				break
			}
		}

		close(done)
	}, 2)

})

type Config struct {
	AccessToken string
}

func getAuthToken() (string, error) {
	bytes, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.cf/config.json"))
	if err != nil {
		return "", err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return "", err
	}

	return config.AccessToken, nil
}
