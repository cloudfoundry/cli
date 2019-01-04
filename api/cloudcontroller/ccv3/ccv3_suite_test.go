package ccv3_test

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
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

var _ = BeforeEach(func() {
	server = NewTLSServer()

	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
})

var _ = AfterEach(func() {
	server.Close()
})

func NewTestClient(config ...Config) (*Client, *ccv3fakes.FakeClock) {
	SetupV3Response()
	var client *Client
	fakeClock := new(ccv3fakes.FakeClock)

	if config != nil {
		client = TestClient(config[0], fakeClock)
	} else {
		client = TestClient(Config{AppName: "CF CLI API V3 Test", AppVersion: "Unknown"}, fakeClock)
	}
	warnings, err := client.TargetCF(TargetSettings{
		SkipSSLValidation: true,
		URL:               server.URL(),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(warnings).To(BeEmpty())

	return client, fakeClock
}

func SetupV3Response() {
	serverURL := server.URL()
	rootResponse := strings.Replace(`{
		"links": {
			"self": {
				"href": "SERVER_URL"
			},
			"cloud_controller_v2": {
				"href": "SERVER_URL/v2",
				"meta": {
					"version": "2.64.0"
				}
			},
			"cloud_controller_v3": {
				"href": "SERVER_URL/v3",
				"meta": {
					"version": "3.0.0-alpha.5"
				}
			},
			"uaa": {
				"href": "https://uaa.bosh-lite.com"
			}
		}
	}`, "SERVER_URL", serverURL, -1)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/"),
			RespondWith(http.StatusOK, rootResponse),
		),
	)

	v3Response := strings.Replace(`{
		"links": {
			"self": {
				"href": "SERVER_URL/v3"
			},
			"apps": {
				"href": "SERVER_URL/v3/apps"
			},
			"tasks": {
				"href": "SERVER_URL/v3/tasks"
			},
			"isolation_segments": {
				"href": "SERVER_URL/v3/isolation_segments"
			},
			"builds": {
				"href": "SERVER_URL/v3/builds"
			},
			"organizations": {
				"href": "SERVER_URL/v3/organizations"
			},
			"service_instances": {
				"href": "SERVER_URL/v3/service_instances"
			},
			"spaces": {
				"href": "SERVER_URL/v3/spaces"
			},
			"packages": {
				"href": "SERVER_URL/v3/packages"
			},
			"processes": {
				"href": "SERVER_URL/v3/processes"
			},
			"droplets": {
				"href": "SERVER_URL/v3/droplets"
			},
			"deployments": {
				"href": "SERVER_URL/v3/deployments"
			},
			"stacks": {
				"href": "SERVER_URL/v3/stacks"
			},
			"buildpacks": {
				"href": "SERVER_URL/v3/buildpacks"
			}
		}
	}`, "SERVER_URL", serverURL, -1)

	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/v3"),
			RespondWith(http.StatusOK, v3Response),
		),
	)
}
