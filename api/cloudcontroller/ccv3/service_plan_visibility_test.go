package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Plan Visibility", func() {
	const planGUID = "fake-service-plan-guid"
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetServicePlanVisibility", func() {
		const withOrganizations = `
        {
            "type": "organization",
            "organizations": [
                {
                    "name": "org-1",
                    "guid": "org-1-guid"
                },
                {
                    "name": "org-2",
                    "guid": "org-2-guid"
                }
            ]
        }`

		const withSpace = `
        {
            "type": "space",
            "space": {
                "name": "space-1",
                "guid": "space-1-guid"
            }
        }`

		DescribeTable(
			"getting a service plan visibility",
			func(response string, expected resources.ServicePlanVisibility) {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/service_plans/%s/visibility", planGUID)),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)

				result, warnings, err := client.GetServicePlanVisibility(planGUID)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(result).To(Equal(expected))
			},
			Entry("admin", `{"type":"admin"}`, resources.ServicePlanVisibility{Type: "admin"}),
			Entry("public", `{"type":"public"}`, resources.ServicePlanVisibility{Type: "public"}),
			Entry("orgs", withOrganizations, resources.ServicePlanVisibility{
				Type: "organization",
				Organizations: []resources.ServicePlanVisibilityDetail{
					{Name: "org-1", GUID: "org-1-guid"},
					{Name: "org-2", GUID: "org-2-guid"},
				},
			}),
			Entry("space", withSpace, resources.ServicePlanVisibility{
				Type:  "space",
				Space: resources.ServicePlanVisibilityDetail{Name: "space-1", GUID: "space-1-guid"},
			}),
		)

		When("the the server responds with error", func() {
			It("returns an error", func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-2"}}),
				)

				_, warnings, err := client.GetServicePlanVisibility(planGUID)
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-2"))
			})
		})
	})

	Describe("UpdateServicePlanVisibility", func() {
		It("sets the plan visibility", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, fmt.Sprintf("/v3/service_plans/%s/visibility", planGUID)),
					VerifyBody([]byte(`{"type":"public"}`)),
					RespondWith(http.StatusOK, `{"type":"public"}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
				),
			)

			result, warnings, err := client.UpdateServicePlanVisibility(
				planGUID,
				resources.ServicePlanVisibility{
					Type: "public",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning-1"))
			Expect(result).To(Equal(resources.ServicePlanVisibility{
				Type: "public",
			}))
		})

		When("the the server responds with error", func() {
			It("returns an error", func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-2"}}),
				)

				_, warnings, err := client.UpdateServicePlanVisibility(
					planGUID,
					resources.ServicePlanVisibility{
						Type: "public",
					},
				)
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-2"))
			})
		})
	})

	Describe("DeleteServicePlanVisibility", func() {
		const orgGUID = "org-guid"

		It("deletes the org visibility", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, fmt.Sprintf("/v3/service_plans/%s/visibility/%s", planGUID, orgGUID)),
					RespondWith(http.StatusNoContent, nil, http.Header{"X-Cf-Warnings": {"warning-1"}}),
				),
			)

			warnings, err := client.DeleteServicePlanVisibility(planGUID, orgGUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning-1"))
		})

		When("the the server responds with error", func() {
			It("returns an error", func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-2"}}),
				)

				warnings, err := client.DeleteServicePlanVisibility(planGUID, orgGUID)
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-2"))
			})
		})
	})
})
