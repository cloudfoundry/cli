package uaa_test

import (
	"bytes"
	"log"
	"testing"

	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	"code.cloudfoundry.org/cli/api/uaa/wrapper"
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

func NewTestUAAClientAndStore() (*Client, *uaafakes.FakeAuthenticationStore) {
	fakeStore := new(uaafakes.FakeAuthenticationStore)
	client := NewClient(Config{
		URL:               server.URL(),
		SkipSSLValidation: true,
		Store:             fakeStore,
		AppName:           "CF CLI UAA API Test",
		AppVersion:        "Unknown",
	})
	client.WrapConnection(wrapper.NewErrorWrapper())
	return client, fakeStore
}
