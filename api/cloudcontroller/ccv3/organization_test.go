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

var _ = Describe("Organizations", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetIsolationSegmentOrganizations", func() {
		var (
			organizations []Organization
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			organizations, warnings, executeErr = client.GetIsolationSegmentOrganizations("some-iso-guid")
		})

		When("organizations exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/isolation_segments/some-iso-guid/organizations?page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "org-name-1",
      "guid": "org-guid-1"
    },
    {
      "name": "org-name-2",
      "guid": "org-guid-2"
    }
  ]
}`, server.URL())
				response2 := `{
	"pagination": {
		"next": null
	},
	"resources": [
	  {
      "name": "org-name-3",
		  "guid": "org-guid-3"
		}
	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid/organizations"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid/organizations", "page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the queried organizations and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(organizations).To(ConsistOf(
					Organization{Name: "org-name-1", GUID: "org-guid-1"},
					Organization{Name: "org-name-2", GUID: "org-guid-2"},
					Organization{Name: "org-name-3", GUID: "org-guid-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
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
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid/organizations"),
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
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetOrganizations", func() {
		var (
			organizations []Organization
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			organizations, warnings, executeErr = client.GetOrganizations(Query{
				Key:    NameFilter,
				Values: []string{"some-org-name"},
			})
		})

		When("organizations exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/organizations?names=some-org-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "org-name-1",
      "guid": "org-guid-1"
    },
    {
      "name": "org-name-2",
      "guid": "org-guid-2"
    }
  ]
}`, server.URL())
				response2 := `{
	"pagination": {
		"next": null
	},
	"resources": [
	  {
      "name": "org-name-3",
		  "guid": "org-guid-3"
		}
	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations", "names=some-org-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations", "names=some-org-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the queried organizations and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(organizations).To(ConsistOf(
					Organization{Name: "org-name-1", GUID: "org-guid-1"},
					Organization{Name: "org-name-2", GUID: "org-guid-2"},
					Organization{Name: "org-name-3", GUID: "org-guid-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
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
      "detail": "Org not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations"),
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
							Detail: "Org not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateOrganization", func() {
		var (
			orgToUpdate Organization
			updatedOrg  Organization
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			updatedOrg, warnings, executeErr = client.UpdateOrganization(orgToUpdate)
		})

		When("the organization is updated successfully", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-app-guid",
					"name": "some-app-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

				expectedBody := map[string]interface{}{
					"name": "some-org-name",
					"metadata": map[string]interface{}{
						"labels": map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/organizations/some-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				orgToUpdate = Organization{
					Name: "some-org-name",
					GUID: "some-guid",
					Metadata: &Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					},
				}
			})

			It("should include the labels in the JSON", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(3))
				Expect(len(warnings)).To(Equal(0))
				Expect(updatedOrg.Metadata.Labels).To(BeEquivalentTo(
					map[string]types.NullString{
						"k1": types.NewNullString("v1"),
						"k2": types.NewNullString("v2"),
					}))
			})

		})

	})
})
