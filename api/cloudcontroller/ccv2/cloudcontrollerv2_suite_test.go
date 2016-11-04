package ccv2_test

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"testing"
)

func TestCloudcontrollerv2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloud Controller V2 Suite")
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
	SetupV2InfoResponse()
	client := NewClient("CF CLI API V2 Test", "Unknown")
	warnings, err := client.TargetCF(TargetSettings{
		SkipSSLValidation: true,
		URL:               server.URL(),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(warnings).To(BeEmpty())
	return client
}

func SetupV2InfoResponse() {
	serverAPIURL := server.URL()[8:]
	response := `{
		"name":"",
		"build":"",
		"support":"http://support.cloudfoundry.com",
		"version":0,
		"description":"",
		"authorization_endpoint":"https://login.APISERVER",
		"token_endpoint":"https://uaa.APISERVER",
		"min_cli_version":null,
		"min_recommended_cli_version":null,
		"api_version":"2.59.0",
		"app_ssh_endpoint":"ssh.APISERVER",
		"app_ssh_host_key_fingerprint":"a6:d1:08:0b:b0:cb:9b:5f:c4:ba:44:2a:97:26:19:8a",
		"routing_endpoint": "https://APISERVER/routing",
		"app_ssh_oauth_client":"ssh-proxy",
		"logging_endpoint":"wss://loggregator.APISERVER",
		"doppler_logging_endpoint":"wss://doppler.APISERVER"
	}`
	response = strings.Replace(response, "APISERVER", serverAPIURL, -1)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/v2/info"),
			RespondWith(http.StatusOK, response),
		),
	)
}
