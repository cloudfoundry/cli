package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Spaces", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateSpace", func() {
		var (
			space         Space
			spaceToCreate Space
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			spaceToCreate = Space{Name: "some-space", Relationships: resources.Relationships{constant.RelationshipTypeOrganization: resources.Relationship{GUID: "some-org-guid"}}}
			space, warnings, executeErr = client.CreateSpace(spaceToCreate)
		})

		When("spaces exist", func() {
			BeforeEach(func() {
				response := `{
      "name": "some-space",
      "guid": "some-space-guid"
    }`

				expectedBody := `{
"name": "some-space",
"relationships": {
      "organization": {
      "data": { "guid": "some-org-guid" }
    }
  }
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/spaces"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the given space and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(space).To(Equal(Space{
					GUID: "some-space-guid",
					Name: "some-space",
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
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/spaces"),
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

	Describe("GetSpaces", func() {
		var (
			query []Query

			includes   IncludedResources
			spaces     []Space
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			spaces, includes, warnings, executeErr = client.GetSpaces(query...)
		})

		When("spaces exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/spaces?names=some-space-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "space-name-1",
      "guid": "space-guid-1",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-1" }
        }
      }
    },
    {
      "name": "space-name-2",
      "guid": "space-guid-2",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-2" }
        }
      }
    }
  ]
}`, server.URL())
				response2 := `{
  "pagination": {
    "next": null
  },
  "resources": [
    {
      "name": "space-name-3",
      "guid": "space-guid-3",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-3" }
        }
      }
    }
  ]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = []Query{{
					Key:    NameFilter,
					Values: []string{"some-space-name"},
				}}
			})

			It("returns the queried spaces and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(spaces).To(ConsistOf(
					Space{Name: "space-name-1", GUID: "space-guid-1", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-1"},
					}},
					Space{Name: "space-name-2", GUID: "space-guid-2", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-2"},
					}},
					Space{Name: "space-name-3", GUID: "space-guid-3", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-3"},
					}},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("the request uses the `include` query key", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/spaces?names=some-space-name&include=organizations&page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "space-name-1",
      "guid": "space-guid-1",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-1" }
        }
      }
    },
    {
      "name": "space-name-2",
      "guid": "space-guid-2",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-2" }
        }
      }
    }
  ],
  "included": {
  		"organizations": [
	  {
		  "guid": "org-guid-1",
		  "name": "org-name-1"
	  }
      ]
  }
}`, server.URL())
				response2 := `{
  "pagination": {
    "next": null
  },
  "resources": [
    {
      "name": "space-name-3",
      "guid": "space-guid-3",
      "relationships": {
        "organization": {
          "data": { "guid": "org-guid-3" }
        }
      }
    }
  ],
"included": {
	"organizations": [
		{
			"guid": "org-guid-2",
			"name": "org-name-2"
		}
	]
}
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name&include=organizations"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces", "names=some-space-name&page=2&per_page=2&include=organizations"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    NameFilter,
						Values: []string{"some-space-name"},
					},
					{
						Key:    Include,
						Values: []string{"organizations"},
					},
				}
			})

			It("returns the given route and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(spaces).To(ConsistOf(
					Space{Name: "space-name-1", GUID: "space-guid-1", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-1"},
					}},
					Space{Name: "space-name-2", GUID: "space-guid-2", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-2"},
					}},
					Space{Name: "space-name-3", GUID: "space-guid-3", Relationships: resources.Relationships{
						constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-3"},
					}},
				))

				Expect(includes).To(Equal(IncludedResources{
					Organizations: []Organization{
						{GUID: "org-guid-1", Name: "org-name-1"},
						{GUID: "org-guid-2", Name: "org-name-2"},
					},
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
      "detail": "Space not found",
      "title": "CF-SpaceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/spaces"),
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
							Detail: "Space not found",
							Title:  "CF-SpaceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateSpace", func() {
		var (
			spaceToUpdate Space
			updatedSpace  Space
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			updatedSpace, warnings, executeErr = client.UpdateSpace(spaceToUpdate)
		})

		When("the organization is updated successfully", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-space-guid",
					"name": "some-space-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

				expectedBody := map[string]interface{}{
					"name": "some-space-name",
					"metadata": map[string]interface{}{
						"labels": map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/spaces/some-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				spaceToUpdate = Space{
					Name: "some-space-name",
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
				Expect(len(warnings)).To(Equal(1))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(updatedSpace.Metadata.Labels).To(BeEquivalentTo(
					map[string]types.NullString{
						"k1": types.NewNullString("v1"),
						"k2": types.NewNullString("v2"),
					}))
			})

		})

	})

	Describe("DeleteSpace", func() {
		var (
			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteSpace("space-guid")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid"),
						RespondWith(http.StatusAccepted, nil, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}, "Location": []string{"job-url"}}),
					))
			})

			It("deletes the Space and returns all warnings", func() {
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
         "detail": "Space not found",
         "title": "CF-ResourceNotFound",
         "code": 10010
      }
   ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Space not found",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})
})
