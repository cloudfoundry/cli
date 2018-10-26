package router_test

import (
	"bytes"
	"log"

	"code.cloudfoundry.org/cli/api/router"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"testing"
)

func TestRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Router Suite")
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

func NewTestConfig() router.Config {
	return router.Config{
		AppName:    "TestApp",
		AppVersion: "1.2.3",
	}
}

func NewTestRouterClient(config router.Config) *router.Client {
	client := router.NewClient(config, nil)
	err := client.SetupResources(server.URL()+"/routing", router.ConnectionConfig{SkipSSLValidation: true})
	Expect(err).ToNot(HaveOccurred())
	return client
}
