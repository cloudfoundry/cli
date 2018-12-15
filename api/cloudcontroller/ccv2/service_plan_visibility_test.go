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
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
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
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("CreateServicePlanVisibility", func() {
		var (
			visibility ServicePlanVisibility
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			visibility, warnings, executeErr = client.CreateServicePlanVisibility("plan-guid-1", "org-guid-1")
		})

		When("the cc returns no error", func() {
			BeforeEach(func() {
				request := map[string]string{
					"service_plan_guid": "plan-guid-1",
					"organization_guid": "org-guid-1",
				}

				response := `{
				"metadata": {
					"guid": "plan-visibility-1",
					"url": "/v2/service_plan_visibilities/f740b01a-4afe-4435-aedd-0a8308a7e7d6",
					"created_at": "2016-06-08T16:41:31Z",
					"updated_at": "2016-06-08T16:41:26Z"
				},
				"entity": {
					"service_plan_guid": "plan-guid-1",
					"organization_guid": "org-guid-1",
					"service_plan_url": "/v2/service_plans/ab5780a9-ac8e-4412-9496-4512e865011a",
					"organization_url": "/v2/organizations/55d0ff39-dac9-431f-ba6d-83f37381f1c3"
				}
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_plan_visibilities"),
						VerifyJSONRepresenting(request),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the created service plan visibility", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(visibility.GUID).To(Equal("plan-visibility-1"))
				Expect(visibility.ServicePlanGUID).To(Equal("plan-guid-1"))
				Expect(visibility.OrganizationGUID).To(Equal("org-guid-1"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_plan_visibilities"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("DeleteServicePlanVisibility", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteServicePlanVisibility("service-plan-visibility-guid")
		})

		When("the cc returns no error", func() {
			BeforeEach(func() {
				response := `{}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/service_plan_visibilities/service-plan-visibility-guid"),
						RespondWith(http.StatusNoContent, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("makes a successful request and returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/service_plan_visibilities/service-plan-visibility-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
