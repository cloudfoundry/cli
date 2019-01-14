package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("SpaceSummary", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaceSummary", func() {
		var (
			spaceSummary      SpaceSummary
			warnings          Warnings
			spaceSummaryError error
		)

		JustBeforeEach(func() {
			spaceSummary, warnings, spaceSummaryError = client.GetSpaceSummary("some-guid")
		})

		When("there are no errors", func() {
			BeforeEach(func() {
				response := `{
				 "name": "space-name",
				 "apps": [
						{
							 "service_names": [
									"service-instance-name"
							 ],
							 "name": "app-name"
						}
				 ],
				 "services": [
						{
							 "name": "service-instance-name",
							 "last_operation": {
									"type": "create",
									"state": "succeeded",
									"description": "a description",
									"updated_at": "some time",
									"created_at": "some time"
							 },
							 "service_plan": {
									"guid": "plan-guid",
									"name": "simple-plan",
									"service": {
										 "guid": "service-guid",
										 "label": "service-label"
									}
							 }
						}
				 ]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-guid/summary"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the result and all warnings", func() {
				Expect(spaceSummaryError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(spaceSummary).To(Equal(SpaceSummary{
					Name: "space-name",
					Applications: []SpaceSummaryApplication{
						{
							Name:         "app-name",
							ServiceNames: []string{"service-instance-name"},
						},
					},
					ServiceInstances: []SpaceSummaryServiceInstance{
						{
							Name: "service-instance-name",
							ServicePlan: SpaceSummaryServicePlan{
								GUID: "plan-guid",
								Name: "simple-plan",
								Service: SpaceSummaryService{
									GUID:  "service-guid",
									Label: "service-label",
								},
							},
							LastOperation: LastOperation{
								Type:        "create",
								State:       "succeeded",
								Description: "a description",
								UpdatedAt:   "some time",
								CreatedAt:   "some time",
							},
						},
					},
				}))
			})
		})

		When("an error occurs", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-guid/summary"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				Expect(spaceSummaryError).To(MatchError(ccerror.V2UnexpectedResponseError{
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
