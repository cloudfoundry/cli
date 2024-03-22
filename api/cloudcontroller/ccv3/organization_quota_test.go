package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Organization Quotas", func() {
	var (
		client     *Client
		executeErr error
		warnings   Warnings
		orgQuotas  []resources.OrganizationQuota
		query      Query
		trueValue  = true
		falseValue = false
	)

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetOrganizationQuotas", func() {
		JustBeforeEach(func() {
			orgQuotas, warnings, executeErr = client.GetOrganizationQuotas(query)
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
						"per_app_tasks": 5,
						"log_rate_limit_in_bytes_per_second": 8
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
							"per_app_tasks": 5,
							"log_rate_limit_in_bytes_per_second": 16
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
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-guid",
							Name: "don-quixote",
							Apps: resources.AppLimit{
								TotalMemory:       &types.NullInt{Value: 5120, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 10, IsSet: true},
								TotalLogVolume:    &types.NullInt{Value: 8, IsSet: true},
							},
							Services: resources.ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 10, IsSet: true},
								PaidServicePlans:      &trueValue,
							},
							Routes: resources.RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 8, IsSet: true},
								TotalReservedPorts: &types.NullInt{Value: 4, IsSet: true},
							},
						},
					},
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-2-guid",
							Name: "sancho-panza",
							Apps: resources.AppLimit{
								TotalMemory:       &types.NullInt{Value: 10240, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 8, IsSet: true},
								TotalLogVolume:    &types.NullInt{Value: 16, IsSet: true},
							},
							Services: resources.ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 8, IsSet: true},
								PaidServicePlans:      &falseValue,
							},
							Routes: resources.RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 10, IsSet: true},
								TotalReservedPorts: &types.NullInt{Value: 5, IsSet: true},
							},
						},
					},
				))
			})
		})

		When("requesting quotas by name", func() {
			BeforeEach(func() {
				query = Query{
					Key:    NameFilter,
					Values: []string{"sancho-panza"},
				}

				response := fmt.Sprintf(`{
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
							"per_app_tasks": 5,
							"log_rate_limit_in_bytes_per_second": 8
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
						VerifyRequest(http.MethodGet, "/v3/organization_quotas", "names=sancho-panza"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"page1 warning"}}),
					),
				)
			})

			It("queries the API with the given name", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("page1 warning"))
				Expect(orgQuotas).To(ConsistOf(
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-2-guid",
							Name: "sancho-panza",
							Apps: resources.AppLimit{
								TotalMemory:       &types.NullInt{Value: 10240, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 8, IsSet: true},
								TotalLogVolume:    &types.NullInt{Value: 8, IsSet: true},
							},
							Services: resources.ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 8, IsSet: true},
								PaidServicePlans:      &falseValue,
							},
							Routes: resources.RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 10, IsSet: true},
								TotalReservedPorts: &types.NullInt{Value: 5, IsSet: true},
							},
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

	Describe("GetOrganizationQuota", func() {
		var (
			returnedOrgQuota resources.OrganizationQuota
			orgQuotaGUID     = "quota_guid"
		)
		JustBeforeEach(func() {
			returnedOrgQuota, warnings, executeErr = client.GetOrganizationQuota(orgQuotaGUID)
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				response := fmt.Sprintf(`{
				  "guid": "quota-guid",
				  "created_at": "2016-05-04T17:00:41Z",
				  "updated_at": "2016-05-04T17:00:41Z",
				  "name": "don-quixote",
				  "apps": {
					"total_memory_in_mb": 5120,
					"per_process_memory_in_mb": 1024,
					"total_instances": 10,
					"per_app_tasks": 5,
					"log_rate_limit_in_bytes_per_second": 8
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
				}`, server.URL())

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/organization_quotas/%s", orgQuotaGUID)),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"show warning"}}),
					),
				)
			})

			It("returns org quotas and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("show warning"))
				Expect(returnedOrgQuota).To(Equal(
					resources.OrganizationQuota{
						Quota: resources.Quota{
							GUID: "quota-guid",
							Name: "don-quixote",
							Apps: resources.AppLimit{
								TotalMemory:       &types.NullInt{Value: 5120, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 10, IsSet: true},
								TotalLogVolume:    &types.NullInt{Value: 8, IsSet: true},
							},
							Services: resources.ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 10, IsSet: true},
								PaidServicePlans:      &trueValue,
							},
							Routes: resources.RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 8, IsSet: true},
								TotalReservedPorts: &types.NullInt{Value: 4, IsSet: true},
							},
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
							"detail": "Quota not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/organization_quotas/%s", orgQuotaGUID)),
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
							Detail: "Quota not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateOrganizationQuota", func() {
		var (
			createdOrgQuota resources.OrganizationQuota
			warnings        Warnings
			executeErr      error
			inputQuota      resources.OrganizationQuota
		)

		BeforeEach(func() {
			inputQuota = resources.OrganizationQuota{
				Quota: resources.Quota{
					Name: "elephant-trunk",
					Apps: resources.AppLimit{
						TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
						InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
						TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
						TotalLogVolume:    &types.NullInt{Value: 0, IsSet: false},
					},
					Services: resources.ServiceLimit{
						TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
						PaidServicePlans:      &trueValue,
					},
					Routes: resources.RouteLimit{
						TotalRoutes:        &types.NullInt{Value: 6, IsSet: true},
						TotalReservedPorts: &types.NullInt{Value: 5, IsSet: true},
					},
				},
			}
		})

		JustBeforeEach(func() {
			createdOrgQuota, warnings, executeErr = client.CreateOrganizationQuota(inputQuota)
		})

		When("the organization quota is created successfully", func() {
			BeforeEach(func() {
				response := `{
					 "guid": "elephant-trunk-guid",
					 "created_at": "2020-01-16T19:44:47Z",
					 "updated_at": "2020-01-16T19:44:47Z",
					 "name": "elephant-trunk",
					 "apps": {
						"total_memory_in_mb": 2048,
						"per_process_memory_in_mb": 1024,
						"total_instances": null,
						"per_app_tasks": null,
						"log_rate_limit_in_bytes_per_second": null
					 },
					 "services": {
						"paid_services_allowed": true,
						"total_service_instances": 0,
						"total_service_keys": null
					 },
					 "routes": {
						"total_routes": 6,
						"total_reserved_ports": 5
					 },
					 "domains": {
						"total_domains": null
					 },
					 "links": {
						"self": {
						   "href": "https://api.foil-venom.lite.cli.fun/v3/organization_quotas/08357710-8106-4d14-b0ea-03154a36fb79"
						}
					 }
				}`

				expectedBody := map[string]interface{}{
					"name": "elephant-trunk",
					"apps": map[string]interface{}{
						"total_memory_in_mb":                 2048,
						"per_process_memory_in_mb":           1024,
						"total_instances":                    nil,
						"log_rate_limit_in_bytes_per_second": nil,
					},
					"services": map[string]interface{}{
						"paid_services_allowed":   true,
						"total_service_instances": 0,
					},
					"routes": map[string]interface{}{
						"total_routes":         6,
						"total_reserved_ports": 5,
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organization_quotas"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created org", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				expectedOrgQuota := inputQuota
				expectedOrgQuota.GUID = "elephant-trunk-guid"
				Expect(createdOrgQuota).To(Equal(expectedOrgQuota))
			})
		})

		When("an organization quota with the same name already exists", func() {
			BeforeEach(func() {
				response := `{
					 "errors": [
							{
								 "detail": "Organization Quota 'anteater-snout' already exists.",
								 "title": "CF-UnprocessableEntity",
								 "code": 10008
							}
					 ]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organization_quotas"),
						RespondWith(http.StatusUnprocessableEntity, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a meaningful organization quota-name-taken error", func() {
				Expect(executeErr).To(MatchError(ccerror.QuotaAlreadyExists{
					Message: "Organization Quota 'anteater-snout' already exists.",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("creating the quota fails", func() {
			BeforeEach(func() {
				response := `{
					 "errors": [
							{
								 "detail": "Fail",
								 "title": "CF-SomeError",
								 "code": 10002
							},
							{
								 "detail": "Something went terribly wrong",
								 "title": "CF-UnknownError",
								 "code": 10001
							}
					 ]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organization_quotas"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10002,
							Detail: "Fail",
							Title:  "CF-SomeError",
						},
						{
							Code:   10001,
							Detail: "Something went terribly wrong",
							Title:  "CF-UnknownError",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("DeleteOrganizationQuota", func() {
		var (
			jobURL       JobURL
			orgQuotaGUID = "quota_guid"
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteOrganizationQuota(orgQuotaGUID)
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, fmt.Sprintf("/v3/organization_quotas/%s", orgQuotaGUID)),
						RespondWith(http.StatusAccepted, nil, http.Header{
							"X-Cf-Warnings": {"delete warning"},
							"Location":      {"/v3/jobs/some-job"},
						}),
					),
				)
			})

			It("returns org quotas and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("delete warning"))
				Expect(jobURL).To(Equal(JobURL("/v3/jobs/some-job")))
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
							"detail": "Quota not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, fmt.Sprintf("/v3/organization_quotas/%s", orgQuotaGUID)),
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
							Detail: "Quota not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateOrganizationQuota", func() {
		var (
			updatedOrgQuota resources.OrganizationQuota
			warnings        Warnings
			executeErr      error
			inputQuota      resources.OrganizationQuota
		)

		BeforeEach(func() {
			inputQuota = resources.OrganizationQuota{
				Quota: resources.Quota{
					GUID: "elephant-trunk-guid",
					Name: "elephant-trunk",
					Apps: resources.AppLimit{
						TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
						InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
						TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
						TotalLogVolume:    &types.NullInt{Value: 8, IsSet: true},
					},
					Services: resources.ServiceLimit{
						TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
						PaidServicePlans:      &trueValue,
					},
				},
			}
		})

		JustBeforeEach(func() {
			updatedOrgQuota, warnings, executeErr = client.UpdateOrganizationQuota(inputQuota)
		})

		When("updating the quota succeeds", func() {
			BeforeEach(func() {
				response := `{
					 "guid": "elephant-trunk-guid",
					 "created_at": "2020-01-16T19:44:47Z",
					 "updated_at": "2020-01-16T19:44:47Z",
					 "name": "elephant-trunk",
					 "apps": {
						"total_memory_in_mb": 2048,
						"per_process_memory_in_mb": 1024,
						"total_instances": null,
						"per_app_tasks": null,
						"log_rate_limit_in_bytes_per_second": 8
					 },
					 "services": {
						"paid_services_allowed": true,
						"total_service_instances": 0,
						"total_service_keys": null
					 },
					 "routes": {
						"total_routes": null,
						"total_reserved_ports": null
					 },
					 "domains": {
						"total_domains": null
					 },
					 "links": {
						"self": {
						   "href": "https://api.foil-venom.lite.cli.fun/v3/organization_quotas/08357710-8106-4d14-b0ea-03154a36fb79"
						}
					 }
				}`

				expectedBody := map[string]interface{}{
					"name": "elephant-trunk",
					"apps": map[string]interface{}{
						"total_memory_in_mb":                 2048,
						"per_process_memory_in_mb":           1024,
						"total_instances":                    nil,
						"log_rate_limit_in_bytes_per_second": 8,
					},
					"services": map[string]interface{}{
						"paid_services_allowed":   true,
						"total_service_instances": 0,
					},
					"routes": map[string]interface{}{},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organization_quotas/elephant-trunk-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the updated org quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(updatedOrgQuota).To(Equal(resources.OrganizationQuota{
					Quota: resources.Quota{
						GUID: "elephant-trunk-guid",
						Name: "elephant-trunk",
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{IsSet: true, Value: 2048},
							InstanceMemory:    &types.NullInt{IsSet: true, Value: 1024},
							TotalAppInstances: &types.NullInt{IsSet: false, Value: 0},
							TotalLogVolume:    &types.NullInt{Value: 8, IsSet: true},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{IsSet: true, Value: 0},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{IsSet: false, Value: 0},
							TotalReservedPorts: &types.NullInt{IsSet: false, Value: 0},
						},
					},
				}))
			})
		})

		When("updating the quota fails", func() {
			BeforeEach(func() {
				response := `{
					 "errors": [
							{
								 "detail": "Fail",
								 "title": "CF-SomeError",
								 "code": 10002
							},
							{
								 "detail": "Something went terribly wrong",
								 "title": "CF-UnknownError",
								 "code": 10001
							}
					 ]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organization_quotas/elephant-trunk-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10002,
							Detail: "Fail",
							Title:  "CF-SomeError",
						},
						{
							Code:   10001,
							Detail: "Something went terribly wrong",
							Title:  "CF-UnknownError",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("ApplyOrganizationQuota", func() {
		var (
			warnings             Warnings
			executeErr           error
			AppliedOrgQuotaGUIDS resources.RelationshipList
			quotaGuid            = "quotaGuid"
			orgGuid              = "orgGuid"
		)

		JustBeforeEach(func() {
			AppliedOrgQuotaGUIDS, warnings, executeErr = client.ApplyOrganizationQuota(quotaGuid, orgGuid)
		})

		When("the organization quota is applied successfully", func() {
			BeforeEach(func() {
				response := `{
					"data": [
						{
							"guid": "orgGuid"
						}
					]
				}`

				expectedBody := map[string][]map[string]string{
					"data": {{"guid": "orgGuid"}},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, fmt.Sprintf("/v3/organization_quotas/%s/relationships/organizations", quotaGuid)),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created org", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(AppliedOrgQuotaGUIDS).To(Equal(resources.RelationshipList{GUIDs: []string{orgGuid}}))
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
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, fmt.Sprintf("/v3/organization_quotas/%s/relationships/organizations", quotaGuid)),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

	})
})
