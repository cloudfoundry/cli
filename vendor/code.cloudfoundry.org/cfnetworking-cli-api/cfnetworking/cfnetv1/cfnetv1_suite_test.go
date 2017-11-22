package cfnetv1_test

import (
	"bytes"
	"log"

	. "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"testing"
)

func TestCFNetV1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CF Networking V1 Client Suite")
}

var server *Server

var _ = SynchronizedBeforeSuite(func() []byte {
	return []byte{}
}, func(data []byte) {
	server = NewTLSServer()

	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
})

var _ = SynchronizedAfterSuite(func() {
	server.Close()
}, func() {})

var _ = BeforeEach(func() {
	server.Reset()
})

func NewTestClient(passed ...Config) *Client {
	var config Config
	if len(passed) > 0 {
		config = passed[0]
	} else {
		config = Config{}
	}
	config.AppName = "CF Networking V1 Test"
	config.AppVersion = "Unknown"
	config.SkipSSLValidation = true

	if config.URL == "" {
		config.URL = server.URL()
	}

	return NewClient(config)
}
