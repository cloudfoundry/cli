package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Target", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = NewClient("CF CLI API V3 Target Test", "Unknown")
	})

	Describe("TargetCF", func() {
		BeforeEach(func() {
			server.Reset()

			serverURL := server.URL()
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
					}
				}
			}`, serverURL, serverURL, serverURL)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(
						http.StatusOK,
						rootResponse,
						http.Header{"X-Cf-Warnings": {"warning 1"}}),
				),
			)

			v3Response := fmt.Sprintf(`{
				"links": {
					"self": {
						"href": "%s/v3"
					},
					"tasks": {
						"href": "%s/v3/tasks"
					}
				}
			}`, serverURL, serverURL)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3"),
					RespondWith(
						http.StatusOK,
						v3Response,
						http.Header{"X-Cf-Warnings": {"warning 2"}}),
				),
			)
		})

		Context("when passed a valid API URL", func() {
			Context("when the server has unverified SSL", func() {
				Context("when setting the skip ssl flag", func() {
					It("sets all the endpoints on the client and returns all warnings", func() {
						warnings, err := client.TargetCF(TargetSettings{
							SkipSSLValidation: true,
							URL:               server.URL(),
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning 1", "warning 2"))

						Expect(client.UAA()).To(Equal("https://uaa.bosh-lite.com"))
						Expect(client.CloudControllerAPIVersion()).To(Equal("3.0.0-alpha.5"))
					})
				})
			})
		})

		Context("when the cloud controller encounters an error", func() {
			BeforeEach(func() {
				server.SetHandler(1,
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3"),
						RespondWith(
							http.StatusNotFound,
							`{"errors": [{}]}`,
							http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the same error", func() {
				warnings, err := client.TargetCF(TargetSettings{
					SkipSSLValidation: true,
					URL:               server.URL(),
				})
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning 1", "this is a warning"))
			})
		})
	})
})
