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
			}
			`, server.URL(), server.URL(), server.URL())

			rootRespondWith = RespondWith(
				http.StatusOK,
				rootResponse,
				http.Header{"X-Cf-Warnings": {"warning 1"}})

			v3Response := fmt.Sprintf(`{
				"links": {
					"self": {
						"href": "%s/v3"
					},
					"tasks": {
						"href": "%s/v3/tasks"
					}
				}
			}
			`, server.URL(), server.URL())

			v3RespondWith = RespondWith(
				http.StatusOK,
				v3Response,
				http.Header{"X-Cf-Warnings": {"warning 2"}})
		})

		It("returns back the CC Information", func() {
			apis, _, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(apis.UAA()).To(Equal("https://uaa.bosh-lite.com"))
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
				Expect(err).To(MatchError(cloudcontroller.APINotFoundError{URL: server.URL()}))
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
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{}))
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
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{Message: "Not found, lol"}))
				Expect(warnings).To(ConsistOf("warning 1", "this is a warning"))
			})
		})
	})
})
