package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Plan", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServicePlan", func() {
		Context("when the service plan exists", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-service-plan-guid"
					},
					"entity": {
						"name": "some-service-plan",
						"service_guid": "some-service-guid"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plans/some-service-plan-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the service plan and warnings", func() {
				servicePlan, warnings, err := client.GetServicePlan("some-service-plan-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(servicePlan).To(Equal(ServicePlan{
					GUID:        "some-service-plan-guid",
					Name:        "some-service-plan",
					ServiceGUID: "some-service-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the service plan does not exist (testing general error case)", func() {
			BeforeEach(func() {
				response := `{
					"description": "The service plan could not be found: non-existant-service-plan-guid",
					"error_code": "CF-ServicePlanNotFound",
					"code": 110003
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plans/non-existant-service-plan-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					))
			})

			It("returns an error and warnings", func() {
				_, warnings, err := client.GetServicePlan("non-existant-service-plan-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The service plan could not be found: non-existant-service-plan-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
