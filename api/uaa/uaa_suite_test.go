package uaa_test

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"testing"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo/v2"
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

func NewTestConfig() *uaafakes.FakeConfig {
	config := new(uaafakes.FakeConfig)
	config.BinaryNameReturns("CF CLI UAA API Test")
	config.BinaryVersionReturns("Unknown")
	config.UAAOAuthClientReturns("client-id")
	config.UAAOAuthClientSecretReturns("client-secret")
	config.SkipSSLValidationReturns(true)
	return config
}

func NewTestUAAClientAndStore(config Config) *Client {
	client := NewClient(config)

	// the 'uaaServer' is discovered via the bootstrapping when we hit the /login
	// endpoint on 'server'
	err := client.SetupResources(uaaServer.URL(), server.URL())
	Expect(err).ToNot(HaveOccurred())

	return client
}

func verifyRequestHost(host string) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		Expect(req.Host).To(Equal(host))
	}
}
