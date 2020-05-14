package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/resources"
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

	Describe("GetDefaultDomain", func() {
		var (
			defaultDomain Domain
			warnings      Warnings
			executeErr    error
			orgGUID       = "some-org-guid"
		)

		JustBeforeEach(func() {
			defaultDomain, warnings, executeErr = client.GetDefaultDomain(orgGUID)
		})

		When("organizations exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
    {
      	"name": "domain-name-1",
      	"guid": "domain-guid-1",
      	"relationships": {
            "organization": {
                "data": {
                    "guid": "some-org-guid"
                }
            }
         },
"internal": false
    }`)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/organizations/%s/domains/default", orgGUID)),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried organizations and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(defaultDomain).To(Equal(
					Domain{Name: "domain-name-1", GUID: "domain-guid-1", Internal: types.NullBool{IsSet: true, Value: false},
						OrganizationGUID: "some-org-guid"},
				))
				Expect(warnings).To(ConsistOf("this is a warning"))
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
						VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/organizations/%s/domains/default", orgGUID)),
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

	Describe("GetOrganization", func() {
		var (
			organization Organization
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			organization, warnings, executeErr = client.GetOrganization("some-org-guid")
		})

		When("organization exists", func() {
			BeforeEach(func() {
				response := `{
					"name": "some-org-name",
					"guid": "some-org-guid",
					"relationships": {
						"quota": {
							"data": {
								"guid": "some-org-quota-guid"
							}
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried organization and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(organization).To(Equal(Organization{
					Name:      "some-org-name",
					GUID:      "some-org-guid",
					QuotaGUID: "some-org-quota-guid",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
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
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid"),
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

	Describe("CreateOrganization", func() {
		var (
			createdOrg Organization
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			createdOrg, warnings, executeErr = client.CreateOrganization("some-org-name")
		})

		When("the organization is created successfully", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-org-guid",
					"name": "some-org-name"
				}`

				expectedBody := map[string]interface{}{
					"name": "some-org-name",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organizations"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created org", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(createdOrg).To(Equal(Organization{
					GUID: "some-org-guid",
					Name: "some-org-name",
				}))
			})
		})

		When("an organization with the same name already exists", func() {
			BeforeEach(func() {
				response := `{
					 "errors": [
							{
								 "detail": "Organization 'some-org-name' already exists.",
								 "title": "CF-UnprocessableEntity",
								 "code": 10008
							}
					 ]
				}`

				expectedBody := map[string]interface{}{
					"name": "some-org-name",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organizations"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusUnprocessableEntity, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a meaningful organization-name-taken error", func() {
				Expect(executeErr).To(MatchError(ccerror.OrganizationNameTakenError{
					UnprocessableEntityError: ccerror.UnprocessableEntityError{
						Message: "Organization 'some-org-name' already exists.",
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("creating the org fails", func() {
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

				expectedBody := map[string]interface{}{
					"name": "some-org-name",
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/organizations"),
						VerifyJSONRepresenting(expectedBody),
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
					"guid": "some-org-guid",
					"name": "some-org-name",
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
				Expect(len(warnings)).To(Equal(1))
				Expect(updatedOrg.Metadata.Labels).To(BeEquivalentTo(
					map[string]types.NullString{
						"k1": types.NewNullString("v1"),
						"k2": types.NewNullString("v2"),
					}))
			})
		})
	})

	Describe("DeleteOrganization", func() {
		var (
			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteOrganization("org-guid")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/organizations/org-guid"),
						RespondWith(http.StatusAccepted, nil, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}, "Location": []string{"job-url"}}),
					))
			})

			It("deletes the Org and returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(jobURL).To(Equal(JobURL("job-url")))
			})
		})

		When("an error is encountered", func() {
			BeforeEach(func() {
				response := `{
   "errors": [
      {
         "detail": "Organization not found",
         "title": "CF-ResourceNotFound",
         "code": 10010
      }
   ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/organizations/org-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Organization not found",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})
})
