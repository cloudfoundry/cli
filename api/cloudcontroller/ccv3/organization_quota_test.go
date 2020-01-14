package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Organization Quotas", func() {
	var client *Client
	var executeErr error
	var warnings Warnings
	var orgQuotas []OrgQuota

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetOrganizationQuotas", func() {
		JustBeforeEach(func() {
			orgQuotas, warnings, executeErr = client.GetOrganizationQuotas()
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
				  "pagination": {
					"total_results": 1,
					"total_pages": 1,
					"first": {
					  "href": "%s/v3/organization_quotas?page=1&per_page=1"
					},
					"last": {
					  "href": "%s/v3/organization_quotas?page=2&per_page=1"
					},
					"next": {
					  "href": "%s/v3/organization_quotas?page=2&per_page=1"
					},
					"previous": null
				  },
				  "resources": [
					{
					  "guid": "quota-guid",
					  "created_at": "2016-05-04T17:00:41Z",
					  "updated_at": "2016-05-04T17:00:41Z",
					  "name": "don-quixote",
					  "apps": {
						"total_memory_in_mb": 5120,
						"per_process_memory_in_mb": 1024,
						"total_instances": 10,
						"per_app_tasks": 5
					  },
					  "services": {
						"paid_services_allowed": true,
						"total_service_instances": 10,
						"total_service_keys": 20
					  },
					  "routes": {
						"total_routes": 8,
						"total_reserved_ports": 4
					  },
					  "domains": {
						"total_private_domains": 7
					  },
					  "relationships": {
						"organizations": {
						  "data": [
							{ "guid": "org-guid1" },
							{ "guid": "org-guid2" }
						  ]
						}
					  },
					  "links": {
						"self": { "href": "%s/v3/organization_quotas/quota-guid" }
					  }
					}
				  ]
				}`, server.URL(), server.URL(), server.URL(), server.URL())

				response2 := fmt.Sprintf(`{
					"pagination": {
						"next": null
					},
					"resources": [
						{
						  "guid": "quota-2-guid",
						  "created_at": "2017-05-04T17:00:41Z",
						  "updated_at": "2017-05-04T17:00:41Z",
						  "name": "sancho-panza",
						  "apps": {
							"total_memory_in_mb": 10240,
							"per_process_memory_in_mb": 1024,
							"total_instances": 8,
							"per_app_tasks": 5
						  },
						  "services": {
							"paid_services_allowed": false,
							"total_service_instances": 8,
							"total_service_keys": 20
						  },
						  "routes": {
							"total_routes": 10,
							"total_reserved_ports": 5
						  },
						  "domains": {
							"total_private_domains": 7
						  },
						  "relationships": {
							"organizations": {
							  "data": []
							}
						  },
						  "links": {
							"self": { "href": "%s/v3/organization_quotas/quota-2-guid" }
						  }
						}
					]
				}`, server.URL())

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organization_quotas"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"page1 warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organization_quotas", "page=2&per_page=1"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"page2 warning"}}),
					),
				)
			})

			It("returns org quotas and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("page1 warning", "page2 warning"))
				Expect(orgQuotas).To(ConsistOf(
					OrgQuota{
						GUID: "quota-guid",
						Name: "don-quixote",
						Apps: AppLimit{
							TotalMemory:       types.NullInt{Value: 5120, IsSet: true},
							InstanceMemory:    types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: types.NullInt{Value: 10, IsSet: true},
						},
						Services: ServiceLimit{
							TotalServiceInstances: types.NullInt{Value: 10, IsSet: true},
							PaidServicePlans:      true,
						},
						Routes: RouteLimit{
							TotalRoutes:     types.NullInt{Value: 8, IsSet: true},
							TotalRoutePorts: types.NullInt{Value: 4, IsSet: true},
						},
					},
					OrgQuota{
						GUID: "quota-2-guid",
						Name: "sancho-panza",
						Apps: AppLimit{
							TotalMemory:       types.NullInt{Value: 10240, IsSet: true},
							InstanceMemory:    types.NullInt{Value: 1024, IsSet: true},
							TotalAppInstances: types.NullInt{Value: 8, IsSet: true},
						},
						Services: ServiceLimit{
							TotalServiceInstances: types.NullInt{Value: 8, IsSet: true},
							PaidServicePlans:      false,
						},
						Routes: RouteLimit{
							TotalRoutes:     types.NullInt{Value: 10, IsSet: true},
							TotalRoutePorts: types.NullInt{Value: 5, IsSet: true},
						},
					},
				))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organization_quotas"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
