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

func NewFakeRequesterTestClient(requester Requester, config ...Config) (*Client, *ccv3fakes.FakeClock) {
	SetupV3Response()
	var client *Client
	fakeClock := new(ccv3fakes.FakeClock)

	if config != nil {
		client = TestClient(config[0], fakeClock, requester)
	} else {
		client = TestClient(
			Config{AppName: "CF CLI API V3 Test", AppVersion: "Unknown"},
			fakeClock,
			requester,
		)
	}
	//settings := TargetSettings{
	//	SkipSSLValidation: true,
	//	URL:               server.URL(),
	//}
	//client.CloudControllerURL = settings.URL
	//client.InitializeConnection(settings)
	//client.Info = Info{
	//	Links:	InfoLinks{
	//		AppSSH:          APILink{HREF: "ssh.environment.cf-app.com:2222"},
	//		CCV3:            APILink{HREF: "https://api.environment.cf-app.com/v3"},
	//		Logging:         APILink{HREF: "wss://doppler.environment.cf-app.com:443"},
	//		LogCache:        APILink{HREF: "https://log-cache.environment.cf-app.com"},
	//		NetworkPolicyV1: APILink{HREF: "https://api.environment.cf-app.com/networking/v1/external"},
	//		Routing:         APILink{HREF: "https://api.environment.cf-app.com/routing"},
	//		UAA:             APILink{HREF: "https://uaa.environment.cf-app.com"},
	//	},
	//}
	//
	//resourceLinks := ResourceLinks{
	//	"self": APILink{HREF: "https://api.environment.cf-app.com/v3"},
	//	"apps": APILink{HREF: "https://api.environment.cf-app.com/v3/apps"},
	//	"audit_events": APILink{HREF: "https://api.environment.cf-app.com/v3/audit_events"},
	//	"buildpacks": APILink{HREF: "https://api.environment.cf-app.com/v3/buildpacks"},
	//	"builds": APILink{HREF: "https://api.environment.cf-app.com/v3/builds"},
	//	"deployments": APILink{HREF: "https://api.environment.cf-app.com/v3/deployments"},
	//	"domains": APILink{HREF: "https://api.environment.cf-app.com/v3/domains"},
	//	"droplets": APILink{HREF: "https://api.environment.cf-app.com/v3/droplets"},
	//	"environment_variable_groups": APILink{HREF: "https://api.environment.cf-app.com/v3/environment_variable_groups"},
	//	"feature_flags": APILink{HREF: "https://api.environment.cf-app.com/v3/feature_flags"},
	//	"info": APILink{HREF: "https://api.environment.cf-app.com/v3/info"},
	//	"isolation_segments": APILink{HREF: "https://api.environment.cf-app.com/v3/isolation_segments"},
	//	"organizations": APILink{HREF: "https://api.environment.cf-app.com/v3/organizations"},
	//	"organization_quotas": APILink{HREF: "https://api.environment.cf-app.com/v3/organization_quotas"},
	//	"packages": APILink{HREF: "https://api.environment.cf-app.com/v3/packages"},
	//	"processes": APILink{HREF: "https://api.environment.cf-app.com/v3/processes"},
	//	"resource_matches": APILink{HREF: "https://api.environment.cf-app.com/v3/resource_matches"},
	//	"roles": APILink{HREF: "https://api.environment.cf-app.com/v3/roles"},
	//	"routes": APILink{HREF: "https://api.environment.cf-app.com/v3/routes"},
	//	"service_brokers": APILink{HREF: "https://api.environment.cf-app.com/v3/service_brokers"},
	//	"service_instances": APILink{HREF: "https://api.environment.cf-app.com/v3/service_instances"},
	//	"service_offerings": APILink{HREF: "https://api.environment.cf-app.com/v3/service_offerings"},
	//	"service_plans": APILink{HREF: "https://api.environment.cf-app.com/v3/service_plans"},
	//	"spaces": APILink{HREF: "https://api.environment.cf-app.com/v3/spaces"},
	//	"space_quotas": APILink{HREF: "https://api.environment.cf-app.com/v3/space_quotas"},
	//	"stacks": APILink{HREF: "https://api.environment.cf-app.com/v3/stacks"},
	//	"tasks": APILink{HREF: "https://api.environment.cf-app.com/v3/tasks"},
	//	"users": APILink{HREF: "https://api.environment.cf-app.com/v3/users"},
	//}
	//resources := map[string]string{}
	//for resource, link := range resourceLinks {
	//	resources[resource] = link.HREF
	//}
	//client.InitializeRouter(resources)

	return client, fakeClock
}

func NewTestClient(config ...Config) (*Client, *ccv3fakes.FakeClock) {
	SetupV3Response()
	var client *Client
	fakeClock := new(ccv3fakes.FakeClock)

	if config != nil {
		client = TestClient(config[0], fakeClock, NewRequester(config[0]))
	} else {
		singleConfig := Config{AppName: "CF CLI API V3 Test", AppVersion: "Unknown"}
		client = TestClient(
			singleConfig,
			fakeClock,
			NewRequester(singleConfig),
		)
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
			"organization_quotas": {
				"href": "SERVER_URL/v3/organization_quotas"
			},
			"security_groups": {
				"href": "SERVER_URL/v3/security_groups"
			},
			"service_brokers": {
				"href": "SERVER_URL/v3/service_brokers"
			},
			"service_instances": {
				"href": "SERVER_URL/v3/service_instances"
			},
			"service_offerings": {
				"href": "SERVER_URL/v3/service_offerings"
			},
			"service_plans": {
				"href": "SERVER_URL/v3/service_plans"
			},
			"spaces": {
				"href": "SERVER_URL/v3/spaces"
			},
			"space_quotas": {
				"href": "SERVER_URL/v3/space_quotas"
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
			"audit_events": {
				"href": "SERVER_URL/v3/audit_events"
			},
            "domains": {
              "href": "SERVER_URL/v3/domains"
            },
			"deployments": {
				"href": "SERVER_URL/v3/deployments"
			},
			"stacks": {
				"href": "SERVER_URL/v3/stacks"
			},
			"buildpacks": {
				"href": "SERVER_URL/v3/buildpacks"
			},
			"feature_flags": {
				"href": "SERVER_URL/v3/feature_flags"
			},
			"resource_matches": {
				"href": "SERVER_URL/v3/resource_matches"
			},
            "roles": {
                "href": "SERVER_URL/v3/roles"
            },
            "routes": {
                "href": "SERVER_URL/v3/routes"
            },
            "users": {
                "href": "SERVER_URL/v3/users"
            },
            "environment_variable_groups": {
                "href": "SERVER_URL/v3/environment_variable_groups"
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
