package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space Quotas", func() {
	var (
		client          *Client
		executeErr      error
		warnings        Warnings
		inputSpaceQuota SpaceQuota

		createdSpaceQuota SpaceQuota
		trueValue         = true
		falseValue        = false
	)

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateSpaceQuotas", func() {
		JustBeforeEach(func() {
			createdSpaceQuota, warnings, executeErr = client.CreateSpaceQuota(inputSpaceQuota)
		})

		When("successfully creating a space quota without spaces", func() {
			BeforeEach(func() {
				inputSpaceQuota = SpaceQuota{
					Quota: Quota{
						Name: "my-space-quota",
						Apps: AppLimit{
							TotalMemory:       &types.NullInt{IsSet: true, Value: 2},
							InstanceMemory:    &types.NullInt{IsSet: true, Value: 3},
							TotalAppInstances: &types.NullInt{IsSet: true, Value: 4},
						},
						Services: ServiceLimit{
							PaidServicePlans:      &trueValue,
							TotalServiceInstances: &types.NullInt{IsSet: true, Value: 5},
						},
						Routes: RouteLimit{
							TotalRoutes:        &types.NullInt{IsSet: true, Value: 6},
							TotalReservedPorts: &types.NullInt{IsSet: true, Value: 7},
						},
					},
					OrgGUID: "some-org-guid",
				}

				response := fmt.Sprintf(`{
  "guid": "space-quota-guid",
  "created_at": "2016-05-04T17:00:41Z",
  "updated_at": "2016-05-04T17:00:41Z",
  "name": "my-space-quota",
  "apps": {
    "total_memory_in_mb": 2,
    "per_process_memory_in_mb": 3,
    "total_instances": 4,
    "per_app_tasks": 900
  },
  "services": {
    "paid_services_allowed": true,
    "total_service_instances": 5,
    "total_service_keys": 700
  },
  "routes": {
    "total_routes": 6,
    "total_reserved_ports": 7
  },
  "relationships": {
    "organization": {
      "data": {
        "guid": "some-org-guid"
      }
    },
    "spaces": {
      "data": []
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/space_quotas/space-quota-guid"
    },
    "organization": {
      "href": "https://api.example.org/v3/organizations/some-org-guid"
    }
  }
}
`)
				expectedBody := map[string]interface{}{
					"name": "my-space-quota",
					"apps": map[string]interface{}{
						"total_memory_in_mb":       2,
						"per_process_memory_in_mb": 3,
						"total_instances":          4,
					},
					"services": map[string]interface{}{
						"paid_services_allowed":   true,
						"total_service_instances": 5,
					},
					"routes": map[string]interface{}{
						"total_routes":         6,
						"total_reserved_ports": 7,
					},
					"relationships": map[string]interface{}{
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"guid": "some-org-guid",
							},
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/space_quotas"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"space-quota-warning"}}),
					),
				)
			})

			It("returns space quota and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("space-quota-warning"))
				Expect(createdSpaceQuota).To(Equal(
					SpaceQuota{
						Quota: Quota{
							GUID: "space-quota-guid",
							Name: "my-space-quota",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{IsSet: true, Value: 2},
								InstanceMemory:    &types.NullInt{IsSet: true, Value: 3},
								TotalAppInstances: &types.NullInt{IsSet: true, Value: 4},
							},
							Services: ServiceLimit{
								PaidServicePlans:      &trueValue,
								TotalServiceInstances: &types.NullInt{IsSet: true, Value: 5},
							},
							Routes: RouteLimit{
								TotalRoutes:        &types.NullInt{IsSet: true, Value: 6},
								TotalReservedPorts: &types.NullInt{IsSet: true, Value: 7},
							},
						},
						OrgGUID: "some-org-guid",
					}))
			})
		})

		When("successfully creating a space quota with spaces", func() {
			BeforeEach(func() {
				inputSpaceQuota = SpaceQuota{
					Quota: Quota{
						Name: "my-space-quota",
						Apps: AppLimit{
							TotalMemory:       &types.NullInt{IsSet: true, Value: 2},
							InstanceMemory:    &types.NullInt{IsSet: true, Value: 3},
							TotalAppInstances: &types.NullInt{IsSet: true, Value: 4},
						},
						Services: ServiceLimit{
							PaidServicePlans:      &trueValue,
							TotalServiceInstances: &types.NullInt{IsSet: true, Value: 6},
						},
						Routes: RouteLimit{
							TotalRoutes:        &types.NullInt{IsSet: true, Value: 8},
							TotalReservedPorts: &types.NullInt{IsSet: true, Value: 9},
						},
					},
					OrgGUID:    "some-org-guid",
					SpaceGUIDs: []string{"space-guid-1", "space-guid-2"},
				}

				response := fmt.Sprintf(`{
  "guid": "space-quota-guid",
  "created_at": "2016-05-04T17:00:41Z",
  "updated_at": "2016-05-04T17:00:41Z",
  "name": "my-space-quota",
  "apps": {
    "total_memory_in_mb": 2,
    "per_process_memory_in_mb": 3,
    "total_instances": 4,
    "per_app_tasks": 5
  },
  "services": {
    "paid_services_allowed": true,
    "total_service_instances": 6,
    "total_service_keys": 7
  },
  "routes": {
    "total_routes": 8,
    "total_reserved_ports": 9
  },
  "relationships": {
    "organization": {
      "data": {
        "guid": "some-org-guid"
      }
    },
    "spaces": {
      "data": [{"guid": "space-guid-1"}, {"guid": "space-guid-2"}]
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/space_quotas/9b370018-c38e-44c9-86d6-155c76801104"
    },
    "organization": {
      "href": "https://api.example.org/v3/organizations/9b370018-c38e-44c9-86d6-155c76801104"
    }
  }
}
`)

				expectedBody := map[string]interface{}{
					"name": "my-space-quota",
					"apps": map[string]interface{}{
						"total_memory_in_mb":       2,
						"per_process_memory_in_mb": 3,
						"total_instances":          4,
					},
					"services": map[string]interface{}{
						"paid_services_allowed":   true,
						"total_service_instances": 6,
					},
					"routes": map[string]interface{}{
						"total_routes":         8,
						"total_reserved_ports": 9,
					},
					"relationships": map[string]interface{}{
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"guid": "some-org-guid",
							},
						},
						"spaces": map[string]interface{}{
							"data": []map[string]interface{}{
								{"guid": "space-guid-1"},
								{"guid": "space-guid-2"},
							},
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/space_quotas"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"space-quota-warning"}}),
					),
				)
			})

			It("returns space quota and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("space-quota-warning"))
				Expect(createdSpaceQuota).To(Equal(
					SpaceQuota{
						Quota: Quota{
							GUID: "space-quota-guid",
							Name: "my-space-quota",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{IsSet: true, Value: 2},
								InstanceMemory:    &types.NullInt{IsSet: true, Value: 3},
								TotalAppInstances: &types.NullInt{IsSet: true, Value: 4},
							},
							Services: ServiceLimit{
								PaidServicePlans:      &trueValue,
								TotalServiceInstances: &types.NullInt{IsSet: true, Value: 6},
							},
							Routes: RouteLimit{
								TotalRoutes:        &types.NullInt{IsSet: true, Value: 8},
								TotalReservedPorts: &types.NullInt{IsSet: true, Value: 9},
							},
						},
						OrgGUID:    "some-org-guid",
						SpaceGUIDs: []string{"space-guid-1", "space-guid-2"},
					}))
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
						VerifyRequest(http.MethodPost, "/v3/space_quotas"),
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

	Describe("GetSpaceQuota", func() {
		var (
			spaceQuota     SpaceQuota
			spaceQuotaGUID = "space-quota-guid"
		)

		JustBeforeEach(func() {
			spaceQuota, warnings, executeErr = client.GetSpaceQuota(spaceQuotaGUID)
		})
		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {

				response := fmt.Sprintf(`{
						  "guid": "space-quota-guid",
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
							"organization": {
							  "data": null
							}
						  },
						  "links": {
							"self": { "href": "%s/v3/space_quotas/space-quota-guid" }
						  }
				}`, server.URL())

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/space_quotas/%s", spaceQuotaGUID)),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"get warning"}}),
					),
				)
			})

			It("queries the API with the given guid", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get warning"))
				Expect(spaceQuota).To(Equal(
					SpaceQuota{
						Quota: Quota{
							GUID: "space-quota-guid",
							Name: "sancho-panza",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{Value: 10240, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 8, IsSet: true},
							},
							Services: ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 8, IsSet: true},
								PaidServicePlans:      &falseValue,
							},
							Routes: RouteLimit{
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
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/space_quotas/%s", spaceQuotaGUID)),
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

	Describe("GetSpaceQuotas", func() {
		var (
			spaceQuotas []SpaceQuota
			query       Query
		)

		JustBeforeEach(func() {
			spaceQuotas, warnings, executeErr = client.GetSpaceQuotas(query)
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
				  "pagination": {
					"total_results": 1,
					"total_pages": 1,
					"first": {
					  "href": "%s/v3/space_quotas?page=1&per_page=1"
					},
					"last": {
					  "href": "%s/v3/space_quotas?page=2&per_page=1"
					},
					"next": {
					  "href": "%s/v3/space_quotas?page=2&per_page=1"
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
						"organization": {
						  "data": { "guid": "org-guid-1" }
						}
					  },
					  "links": {
						"self": { "href": "%s/v3/space_quotas/quota-guid" }
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
							"organization": {
							  "data": null
							}
						  },
						  "links": {
							"self": { "href": "%s/v3/space_quotas/quota-2-guid" }
						  }
						}
					]
				}`, server.URL())

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/space_quotas"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"page1 warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/space_quotas", "page=2&per_page=1"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"page2 warning"}}),
					),
				)
			})

			It("returns space quotas and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("page1 warning", "page2 warning"))
				Expect(spaceQuotas).To(ConsistOf(
					SpaceQuota{
						Quota: Quota{
							GUID: "quota-guid",
							Name: "don-quixote",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{Value: 5120, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 10, IsSet: true},
							},
							Services: ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 10, IsSet: true},
								PaidServicePlans:      &trueValue,
							},
							Routes: RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 8, IsSet: true},
								TotalReservedPorts: &types.NullInt{Value: 4, IsSet: true},
							},
						},
						OrgGUID: "org-guid-1",
					},
					SpaceQuota{
						Quota: Quota{
							GUID: "quota-2-guid",
							Name: "sancho-panza",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{Value: 10240, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 8, IsSet: true},
							},
							Services: ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 8, IsSet: true},
								PaidServicePlans:      &falseValue,
							},
							Routes: RouteLimit{
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
							"organization": {
							  "data": null
							}
						  },
						  "links": {
							"self": { "href": "%s/v3/space_quotas/quota-2-guid" }
						  }
						}
					]
				}`, server.URL())

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/space_quotas", "names=sancho-panza"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"page1 warning"}}),
					),
				)
			})

			It("queries the API with the given name", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("page1 warning"))
				Expect(spaceQuotas).To(ConsistOf(
					SpaceQuota{
						Quota: Quota{
							GUID: "quota-2-guid",
							Name: "sancho-panza",
							Apps: AppLimit{
								TotalMemory:       &types.NullInt{Value: 10240, IsSet: true},
								InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
								TotalAppInstances: &types.NullInt{Value: 8, IsSet: true},
							},
							Services: ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 8, IsSet: true},
								PaidServicePlans:      &falseValue,
							},
							Routes: RouteLimit{
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
						VerifyRequest(http.MethodGet, "/v3/space_quotas"),
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

	Describe("UpdateSpaceQuota", func() {
		var (
			updatedSpaceQuota SpaceQuota
			warnings          Warnings
			executeErr        error
			inputQuota        SpaceQuota
		)

		BeforeEach(func() {
			inputQuota = SpaceQuota{
				Quota: Quota{
					GUID: "elephant-trunk-guid",
					Name: "elephant-trunk",
					Apps: AppLimit{
						TotalMemory:       &types.NullInt{Value: 2048, IsSet: true},
						InstanceMemory:    &types.NullInt{Value: 1024, IsSet: true},
						TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
					},
					Services: ServiceLimit{
						TotalServiceInstances: &types.NullInt{Value: 0, IsSet: true},
						PaidServicePlans:      &trueValue,
					},
				},
			}
		})

		JustBeforeEach(func() {
			updatedSpaceQuota, warnings, executeErr = client.UpdateSpaceQuota(inputQuota)
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
						"per_app_tasks": null
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
					 "links": {
						"self": {
						   "href": "https://api.foil-venom.lite.cli.fun/v3/space_quotas/08357710-8106-4d14-b0ea-03154a36fb79"
						}
					 }
				}`

				expectedBody := map[string]interface{}{
					"name": "elephant-trunk",
					"apps": map[string]interface{}{
						"total_memory_in_mb":       2048,
						"per_process_memory_in_mb": 1024,
						"total_instances":          nil,
					},
					"services": map[string]interface{}{
						"paid_services_allowed":   true,
						"total_service_instances": 0,
					},
					"routes": map[string]interface{}{},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/space_quotas/elephant-trunk-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the updated space quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(updatedSpaceQuota).To(Equal(SpaceQuota{
					Quota: Quota{
						GUID: "elephant-trunk-guid",
						Name: "elephant-trunk",
						Apps: AppLimit{
							TotalMemory:       &types.NullInt{IsSet: true, Value: 2048},
							InstanceMemory:    &types.NullInt{IsSet: true, Value: 1024},
							TotalAppInstances: &types.NullInt{IsSet: false, Value: 0},
						},
						Services: ServiceLimit{
							TotalServiceInstances: &types.NullInt{IsSet: true, Value: 0},
							PaidServicePlans:      &trueValue,
						},
						Routes: RouteLimit{
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
						VerifyRequest(http.MethodPatch, "/v3/space_quotas/elephant-trunk-guid"),
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
})
