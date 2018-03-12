package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetService", func() {
		Context("when the service exists", func() {
			Context("when the value of the 'extra' json key is non-empty", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"label": "some-service",
							"description": "some-description",
							"documentation_url": "some-url",
							"extra": "{\"provider\":{\"name\":\"The name\"},\"listing\":{\"imageUrl\":\"http://catgifpage.com/cat.gif\",\"blurb\":\"fake broker that is fake\",\"longDescription\":\"A long time ago, in a galaxy far far away...\"},\"displayName\":\"The Fake Broker\",\"shareable\":true}"
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the service and warnings", func() {
					service, warnings, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:             "some-service-guid",
						Label:            "some-service",
						Description:      "some-description",
						DocumentationURL: "some-url",
						Extra: ServiceExtra{
							Shareable: true,
						},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when the value of the 'extra' json key is null", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"extra": null
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})

			Context("when the value of the 'extra' json key is the empty string", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"extra": ""
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})

			Context("when the key 'extra' is not in the json response", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})
		})

		Context("when the service does not exist (testing general error case)", func() {
			BeforeEach(func() {
				response := `{
					"description": "The service could not be found: non-existant-service-guid",
					"error_code": "CF-ServiceNotFound",
					"code": 120003
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/services/non-existant-service-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					))
			})

			It("returns an error and warnings", func() {
				_, warnings, err := client.GetService("non-existant-service-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The service could not be found: non-existant-service-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
