package ccv3_test

import (
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Info", func() {
	var (
		client          *Client
		rootRespondWith http.HandlerFunc

		info       Info
		warnings   Warnings
		executeErr error
	)

	BeforeEach(func() {
		rootRespondWith = nil
	})

	JustBeforeEach(func() {
		client, _ = NewTestClient()

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/"),
				rootRespondWith,
			),
		)

		info, warnings, executeErr = client.GetInfo()
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
		})

		It("returns the CC Information", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(info.UAA()).To(Equal("https://uaa.bosh-lite.com"))
			Expect(info.Logging()).To(Equal("wss://doppler.bosh-lite.com:443"))
			Expect(info.NetworkPolicyV1()).To(Equal(fmt.Sprintf("%s/networking/v1/external", server.URL())))
			Expect(info.AppSSHHostKeyFingerprint()).To(Equal("some-fingerprint"))
			Expect(info.AppSSHEndpoint()).To(Equal("ssh.bosh-lite.com:2222"))
			Expect(info.OAuthClient()).To(Equal("some-client"))
			Expect(info.CFOnK8s).To(BeFalse())
		})

		It("returns all warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning 1"))
		})

		When("CF-on-K8s", func() {
			BeforeEach(func() {
				rootRespondWith = RespondWith(http.StatusOK, `{ "cf_on_k8s": true }`)
			})

			It("sets the CFOnK8s", func() {
				Expect(info.CFOnK8s).To(BeTrue())
			})
		})
	})

	When("the cloud controller encounters an error", func() {
		When("the root response is invalid", func() {
			BeforeEach(func() {
				rootRespondWith = RespondWith(
					http.StatusNotFound,
					`i am google, bow down`,
					http.Header{"X-Cf-Warnings": {"warning 2"}},
				)
			})

			It("returns an APINotFoundError and no warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.APINotFoundError{URL: server.URL()}))
				Expect(warnings).To(BeNil())
			})
		})

		When("the error occurs making a request to '/'", func() {
			BeforeEach(func() {
				rootRespondWith = RespondWith(
					http.StatusNotFound,
					`{"errors": [{}]}`,
					http.Header{"X-Cf-Warnings": {"this is a warning"}})
			})

			It("returns the same error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
