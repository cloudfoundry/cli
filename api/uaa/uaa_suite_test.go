package uaa_test

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestUaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UAA Suite")
}

var (
	// we create two servers in order to test that requests using different
	// resources are going to the correct server
	server    *Server
	uaaServer *Server

	TestAuthorizationResource string
	TestUAAResource           string
	TestSuiteFakeStore        *uaafakes.FakeUAAEndpointStore
)

var _ = SynchronizedBeforeSuite(func() []byte {
	return []byte{}
}, func(data []byte) {
	server = NewTLSServer()
	uaaServer = NewTLSServer()

	testAuthURL, err := url.Parse(server.URL())
	Expect(err).ToNot(HaveOccurred())
	TestAuthorizationResource = testAuthURL.Host

	testUAAURL, err := url.Parse(uaaServer.URL())
	Expect(err).ToNot(HaveOccurred())
	TestUAAResource = testUAAURL.Host

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
	SetupBootstrapResponse()

	client := NewClient(Config{
		AppName:           "CF CLI UAA API Test",
		AppVersion:        "Unknown",
		ClientID:          "client-id",
		ClientSecret:      "client-secret",
		SkipSSLValidation: true,
	})

	// the 'uaaServer' is discovered via the bootstrapping when we hit the /login
	// endpoint on 'server'
	TestSuiteFakeStore = new(uaafakes.FakeUAAEndpointStore)
	err := client.SetupResources(TestSuiteFakeStore, server.URL())
	Expect(err).ToNot(HaveOccurred())

	return client
}

func SetupBootstrapResponse() {
	response := strings.Replace(`{
				"links": {
					"uaa": "SERVER_URL"
				}
			}`, "SERVER_URL", uaaServer.URL(), -1)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/login"),
			RespondWith(http.StatusOK, response),
		),
	)
}

func verifyRequestHost(host string) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		Expect(req.Host).To(Equal(host))
	}
}
