package ccv3_test

import (
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Info", func() {
	var (
		client          *Client
		rootRespondWith http.HandlerFunc
		v3RespondWith   http.HandlerFunc
	)

	JustBeforeEach(func() {
		client = NewTestClient()

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/"),
				rootRespondWith,
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/v3"),
				v3RespondWith,
			))
	})

	Describe("when all requests are successful", func() {
		BeforeEach(func() {
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
					"network_policy_v1": {
						"href": "SERVER_URL/networking/v1/external"
					},
					"uaa": {
						"href": "https://uaa.bosh-lite.com"
					},
					"logging": {
						"href": "wss://doppler.bosh-lite.com:443"
					},
					"app_ssh": {
						"href": "ssh.bosh-lite.com:2222",
						"meta": {
							"host_key_fingerprint": "some-fingerprint",
							"oath_client": "some-client"
						}
					}
				}
			}`, "SERVER_URL", server.URL(), -1)

			rootRespondWith = RespondWith(
				http.StatusOK,
				rootResponse,
				http.Header{"X-Cf-Warnings": {"warning 1"}})

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
					}
				}
			}`, "SERVER_URL", server.URL(), -1)

			v3RespondWith = RespondWith(
				http.StatusOK,
				v3Response,
				http.Header{"X-Cf-Warnings": {"warning 2"}})
		})

		It("returns the CC Information", func() {
			apis, _, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(apis.UAA()).To(Equal("https://uaa.bosh-lite.com"))
			Expect(apis.Logging()).To(Equal("wss://doppler.bosh-lite.com:443"))
			Expect(apis.NetworkPolicyV1()).To(Equal(fmt.Sprintf("%s/networking/v1/external", server.URL())))
			Expect(apis.AppSSHHostKeyFingerprint()).To(Equal("some-fingerprint"))
			Expect(apis.AppSSHEndpoint()).To(Equal("ssh.bosh-lite.com:2222"))
			Expect(apis.OAuthClient()).To(Equal("some-client"))
		})

		It("returns back the resource links", func() {
			_, resources, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(resources[internal.AppsResource].HREF).To(Equal(server.URL() + "/v3/apps"))
			Expect(resources[internal.TasksResource].HREF).To(Equal(server.URL() + "/v3/tasks"))
		})

		It("returns all warnings", func() {
			_, _, warnings, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning 1", "warning 2"))
		})
	})

	Context("when the cloud controller encounters an error", func() {
		Context("when the root response is invalid", func() {
			BeforeEach(func() {
				rootRespondWith = RespondWith(
					http.StatusNotFound,
					`i am google, bow down`,
				)
			})

			It("returns an APINotFoundError", func() {
				_, _, _, err := client.Info()
				Expect(err).To(MatchError(ccerror.APINotFoundError{URL: server.URL()}))
			})
		})

		Context("when the error occurs making a request to '/'", func() {
			BeforeEach(func() {
				rootRespondWith = RespondWith(
					http.StatusNotFound,
					`{"errors": [{}]}`,
					http.Header{"X-Cf-Warnings": {"this is a warning"}})
			})

			It("returns the same error", func() {
				_, _, warnings, err := client.Info()
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the error occurs making a request to '/v3'", func() {
			BeforeEach(func() {
				rootResponse := fmt.Sprintf(`{
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
						},
						"uaa": {
							"href": "https://uaa.bosh-lite.com"
						},
						"logging": {
							"href": "wss://doppler.bosh-lite.com:443"
						}
					}
				}
				`, server.URL(), server.URL(), server.URL())

				rootRespondWith = RespondWith(
					http.StatusOK,
					rootResponse,
					http.Header{"X-Cf-Warnings": {"warning 1"}})
				v3RespondWith = RespondWith(
					http.StatusNotFound,
					`{"errors": [{
							"code": 10010,
							"title": "CF-ResourceNotFound",
							"detail": "Not found, lol"
						}]
					}`,
					http.Header{"X-Cf-Warnings": {"this is a warning"}})
			})

			It("returns the same error", func() {
				_, _, warnings, err := client.Info()
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "Not found, lol"}))
				Expect(warnings).To(ConsistOf("warning 1", "this is a warning"))
			})
		})
	})
})
