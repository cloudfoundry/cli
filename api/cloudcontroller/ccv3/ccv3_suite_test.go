package ccv3_test

import (
	"bytes"
	"fmt"
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
	RunSpecs(t, "Cloud Controller V3 Suite")
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
	SetupV3Response()
	client := NewClient("CF CLI API V3 Test", "Unknown")
	warnings, err := client.TargetCF(TargetSettings{
		SkipSSLValidation: true,
		URL:               server.URL(),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(warnings).To(BeEmpty())

	return client
}

func SetupV3Response() {
	serverURL := server.URL()
	rootResponse := fmt.Sprintf(`
{
  "links": {
    "self": {
      "href": "%s"
    },
    "cloud_controller_v2": {
      "href": "%s/v2",
      "meta": {
        "version": "2.64.0"
      }
    },
    "cloud_controller_v3": {
      "href": "%s/v3",
      "meta": {
        "version": "3.0.0-alpha.5"
      }
    }
  }
}
`, serverURL, serverURL, serverURL)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/"),
			RespondWith(http.StatusOK, rootResponse),
		),
	)

	v3Response := fmt.Sprintf(`
{
  "links": {
    "self": {
      "href": "%s/v3"
    },
    "tasks": {
      "href": "%s/v3/tasks"
    },
    "uaa": {
      "href": "https://uaa.bosh-lite.com"
    }
  }
}
`, serverURL, serverURL)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/v3"),
			RespondWith(http.StatusOK, v3Response),
		),
	)
}
