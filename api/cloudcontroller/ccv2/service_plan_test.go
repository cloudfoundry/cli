package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
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
		When("the service plan exists", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-service-plan-guid"
					},
					"entity": {
						"name": "some-service-plan",
						"public": true,
						"service_guid": "some-service-guid",
						"description": "some-description",
						"free": true
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
					Public:      true,
					ServiceGUID: "some-service-guid",
					Description: "some-description",
					Free:        true,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("the service plan does not exist (testing general error case)", func() {
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

	Describe("GetServicePlans", func() {
		var (
			services   []ServicePlan
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			services, warnings, executeErr = client.GetServicePlans(Filter{
				Type:     constant.ServiceGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-service-guid"},
			})
		})

		When("the cc returns back service plans", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/service_plans?q=service_guid:some-service-guid&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "some-service-plan-guid-1"
							},
							"entity": {
								"name": "some-service-plan",
								"service_guid": "some-service-guid",
								"free": false,
								"description": "some-description"
							}
						},
						{
							"metadata": {
								"guid": "some-service-plan-guid-2"
							},
							"entity": {
								"name": "other-service-plan",
								"service_guid": "some-service-guid",
								"free": true,
								"description": "other-description"
							}
						}
					]
				}`

				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-service-plan-guid-3"
							},
							"entity": {
								"name": "some-service-plan",
								"service_guid": "some-service-guid",
								"free": false,
								"description": "some-description"
							}
						},
						{
							"metadata": {
								"guid": "some-service-plan-guid-4"
							},
							"entity": {
								"name": "other-service-plan",
								"service_guid": "some-service-guid",
								"free": true,
								"description": "other-description"
							}
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plans", "q=service_guid:some-service-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plans", "q=service_guid:some-service-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried services", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(services).To(ConsistOf([]ServicePlan{
					{GUID: "some-service-plan-guid-1", ServiceGUID: "some-service-guid", Free: false, Description: "some-description", Name: "some-service-plan"},
					{GUID: "some-service-plan-guid-2", ServiceGUID: "some-service-guid", Free: true, Description: "other-description", Name: "other-service-plan"},
					{GUID: "some-service-plan-guid-3", ServiceGUID: "some-service-guid", Free: false, Description: "some-description", Name: "some-service-plan"},
					{GUID: "some-service-plan-guid-4", ServiceGUID: "some-service-guid", Free: true, Description: "other-description", Name: "other-service-plan"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "Some description.",
					"error_code": "CF-Error",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plans"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "Some description.",
						ErrorCode:   "CF-Error",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
