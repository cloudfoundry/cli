package ccv3_test

import (
	"bytes"
	"log"
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"testing"
)

func TestCcv3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ccv3 Suite")
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

func NewTestClient() *CloudControllerClient {
	SetupV3Response()
	client := NewCloudControllerClient()
	warnings, err := client.TargetCF(server.URL(), true)
	Expect(err).ToNot(HaveOccurred())
	Expect(warnings).To(BeEmpty())
	return client
}

func SetupV3Response() {
	response := `{}`
	// response = strings.Replace(response, "APISERVER", serverAPIURL, -1)
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/v3/"),
			RespondWith(http.StatusOK, response),
		),
	)
}
