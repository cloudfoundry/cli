package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/types"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space Quotas", func() {
	var client *Client
	var executeErr error
	var warnings Warnings
	var spaceQuota SpaceQuota

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateSpaceQuotas", func() {
		JustBeforeEach(func() {
			spaceQuota, warnings, executeErr = client.CreateSpaceQuota(spaceQuota)
		})

		When("the cloud controller returns without errors", func() {
			BeforeEach(func() {
				spaceQuota = SpaceQuota{
					GUID: "space-quota-guid",
					Name: "my-space-quota",
					Apps: AppLimit{
						TotalMemory:       types.NullInt{IsSet: true, Value: 2},
						InstanceMemory:    types.NullInt{IsSet: true, Value: 3},
						TotalAppInstances: types.NullInt{IsSet: true, Value: 4},
					},
					Services: ServiceLimit{
						PaidServicePlans:      true,
						TotalServiceInstances: types.NullInt{IsSet: true, Value: 5},
					},
					Routes: RouteLimit{
						TotalRoutes:     types.NullInt{IsSet: true, Value: 6},
						TotalRoutePorts: types.NullInt{IsSet: true, Value: 7},
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
      "data": [{
          "guid": "some-space-guid"
        }]
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/organization_quotas/9b370018-c38e-44c9-86d6-155c76801104"
    },
    "organization": {
      "href": "https://api.example.org/v3/organizations/9b370018-c38e-44c9-86d6-155c76801104"
    }
  }
}
`)
				_ = map[string]interface{}{
					"name": "production",
					"relationships": map[string]interface{}{
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"guid": "[org-guid]",
							},
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/space_quotas"),
						//VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"space-quota-warning"}}),
					),
				)
			})

			It("returns space quota and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("space-quota-warning"))
				Expect(spaceQuota).To(Equal(
					SpaceQuota{
						GUID: "space-quota-guid",
						Name: "my-space-quota",
						Apps: AppLimit{
							TotalMemory:       types.NullInt{IsSet: true, Value: 2},
							InstanceMemory:    types.NullInt{IsSet: true, Value: 3},
							TotalAppInstances: types.NullInt{IsSet: true, Value: 4},
						},
						Services: ServiceLimit{
							PaidServicePlans:      true,
							TotalServiceInstances: types.NullInt{IsSet: true, Value: 5},
						},
						Routes: RouteLimit{
							TotalRoutes:     types.NullInt{IsSet: true, Value: 6},
							TotalRoutePorts: types.NullInt{IsSet: true, Value: 7},
						},
						OrgGUID:    "some-org-guid",
						SpaceGUIDs: []string{"some-space-guid"},
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
})
