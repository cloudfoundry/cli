package ccv2_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space Quotas", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaceQuotaDefinition", func() {
		When("no errors are encountered", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "space-quota-guid",
						"url": "/v2/space_quota_definitions/space-quota-guid",
						"updated_at": null
					},
					"entity": {
						"name": "space-quota"
					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/space_quota_definitions/space-quota-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the Space Quota", func() {
				spaceQuota, warnings, err := client.GetSpaceQuotaDefinition("space-quota-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(spaceQuota).To(Equal(SpaceQuota{
					Name: "space-quota",
					GUID: "space-quota-guid",
				}))
			})
		})

		When("the request returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 210002,
					"description": "The space quota could not be found: some-space-quota-guid",
					"error_code": "CF-AppNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/space_quota_definitions/some-space-quota-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.GetSpaceQuotaDefinition("some-space-quota-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The space quota could not be found: some-space-quota-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetSpaceQuotas", func() {
		When("the organization is not found", func() {
			var badOrgGuid = "bad-org-guid"
			BeforeEach(func() {
				response := fmt.Sprintf(`{
					"description": "The organization could not be found: %s",
					"error_code": "CF-OrganizationNotFound",
					"code": 30003
			 }`, badOrgGuid)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v2/organizations/%s/space_quota_definitions", badOrgGuid)),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					))
			})

			It("returns an error indicating that the org was not found", func() {
				_, warnings, err := client.GetSpaceQuotas(badOrgGuid)
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: fmt.Sprintf("The organization could not be found: %s", badOrgGuid)}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("the organization is found", func() {
			BeforeEach(func() {
				response := `{
				  "resources": [
  					  {
  					    "metadata": {
  					      "guid": "some-space-quota-guid-1"
  					    },
  					    "entity": {
  					      "name": "some-quota-1",
  					      "organization_guid": "some-org-guid"
  					    }
  					  },
  					  {
  					    "metadata": {
  					      "guid": "some-space-quota-guid-2"
  					    },
  					    "entity": {
  					      "name": "some-quota-2",
  					      "organization_guid": "some-org-guid"
  					    }
  					  }
  					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/space_quota_definitions"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all the space quotas for the org guid", func() {
				spaceQuotas, warnings, err := client.GetSpaceQuotas("some-org-guid")
				Expect(spaceQuotas).To(ConsistOf([]SpaceQuota{
					SpaceQuota{GUID: "some-space-quota-guid-1", Name: "some-quota-1"},
					SpaceQuota{GUID: "some-space-quota-guid-2", Name: "some-quota-2"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("SetSpaceQuotas", func() {
		When("the calls succeeds", func() {
			BeforeEach(func() {
				response := `{
					"description": "Could not find VCAP::CloudController::Space with guid: some-space-guid",
					"error_code": "CF-InvalidRelation",
					"code": 1002
				 }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/space_quota_definitions/some-space-quota-guid/spaces/some-space-guid"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("sets the space quota for the space and retruns warnings", func() {
				warnings, err := client.SetSpaceQuota("some-space-guid", "some-space-quota-guid")
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(err).To(MatchError(ccerror.InvalidRelationError{Message: "Could not find VCAP::CloudController::Space with guid: some-space-guid"}))
			})
		})

		When("the calls fails", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "space-quota-guid"
					  },
					  "entity": {
						"name": "some-space-quota",
						"organization_guid": "some-org-guid"
					  }
					
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/space_quota_definitions/some-space-quota-guid/spaces/some-space-guid"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("sets the space quota for the space and retruns warnings", func() {
				warnings, err := client.SetSpaceQuota("some-space-guid", "some-space-quota-guid")
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
})
