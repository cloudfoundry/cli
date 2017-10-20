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
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-service-guid"
					},
					"entity": {
						"label": "some-service",
						"description": "some-description",
						"documentation_url": "some-url"
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
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
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
