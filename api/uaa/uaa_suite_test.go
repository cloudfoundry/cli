package uaa_test

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"

	. "code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestUaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UAA Suite")
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

func NewTestUAAClientAndStore() *Client {
	SetupAuthResponse()

	client := NewClient(Config{
		AppName:           "CF CLI UAA API Test",
		AppVersion:        "Unknown",
		ClientID:          "client-id",
		ClientSecret:      "client-secret",
		SkipSSLValidation: true,
	})

	err := client.SetupResources(server.URL())

	Expect(err).ToNot(HaveOccurred())

	return client
}

func SetupAuthResponse() {
	serverURL := server.URL()

	response := strings.Replace(`{
				"links": {
					"uaa": "SERVER_URL",
					"login": "SERVER_URL"
				}
			}`, "SERVER_URL", serverURL, -1)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/login"),
			RespondWith(http.StatusOK, response),
		),
	)
}
