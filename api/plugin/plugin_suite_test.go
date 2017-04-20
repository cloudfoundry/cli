package plugin_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	. "code.cloudfoundry.org/cli/api/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"testing"
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Suite")
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

func NewTestClient() *Client {
	client := NewClient(Config{SkipSSLValidation: true, AppName: "CF CLI API Pluting Test", AppVersion: "Unknown"})
	return client
}

func testPluginRepositoryServer(pluginRepo PluginRepository) string {
	jsonBytes, err := json.Marshal(pluginRepo)
	Expect(err).ToNot(HaveOccurred())

	server.AppendHandlers(
		VerifyRequest(http.MethodGet, "/list"),
		RespondWith(http.StatusOK, string(jsonBytes)),
	)

	return server.URL()
}
