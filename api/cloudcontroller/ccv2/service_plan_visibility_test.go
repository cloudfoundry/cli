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

var _ = Describe("Service Plan Visibility", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServicePlanVisibilities", func() {
		var (
			visibilities []ServicePlanVisibility
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			visibilities, warnings, executeErr = client.GetServicePlanVisibilities(Filter{
				Type:     constant.OrganizationGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-org-guid"},
			})
		})

		When("the cc returns back service brokers", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/service_plan_visibilities?q=organization_guid:some-org-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-visibility-guid-1"
						},
						"entity": {
							"service_plan_guid": "some-service-plan-guid-1",
							"organization_guid": "some-org-guid"
						}
					},
					{
						"metadata": {
							"guid": "some-visibility-guid-2"
						},
						"entity": {
							"service_plan_guid": "some-service-plan-guid-2",
							"organization_guid": "some-org-guid"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "some-visibility-guid-3"
						},
						"entity": {
							"service_plan_guid": "some-service-plan-guid-3",
							"organization_guid": "some-org-guid"
						}
					},
					{
						"metadata": {
							"guid": "some-visibility-guid-4"
						},
						"entity": {
							"service_plan_guid": "some-service-plan-guid-4",
							"organization_guid": "some-org-guid"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plan_visibilities", "q=organization_guid:some-org-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plan_visibilities", "q=organization_guid:some-org-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried service brokers", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(visibilities).To(ConsistOf([]ServicePlanVisibility{
					{GUID: "some-visibility-guid-1", OrganizationGUID: "some-org-guid", ServicePlanGUID: "some-service-plan-guid-1"},
					{GUID: "some-visibility-guid-2", OrganizationGUID: "some-org-guid", ServicePlanGUID: "some-service-plan-guid-2"},
					{GUID: "some-visibility-guid-3", OrganizationGUID: "some-org-guid", ServicePlanGUID: "some-service-plan-guid-3"},
					{GUID: "some-visibility-guid-4", OrganizationGUID: "some-org-guid", ServicePlanGUID: "some-service-plan-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "The broker is broken.",
					"error_code": "CF-BrokenBroker",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_plan_visibilities"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "The broker is broken.",
						ErrorCode:   "CF-BrokenBroker",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
