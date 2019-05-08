package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/types"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Domain", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateDomain for Shared Domains", func() {
		var (
			domain     Domain
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			domain, warnings, executeErr = client.CreateDomain(Domain{Name: "some-name", Internal: types.NullBool{IsSet: true, Value: true}})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-guid",
					"name": "some-name",
					"internal": true
				}`

				expectedBody := `{
					"name": "some-name",
					"internal": true
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/domains"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given domain and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(domain).To(Equal(Domain{
					GUID:     "some-guid",
					Name:     "some-name",
					Internal: types.NullBool{IsSet: true, Value: true},
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
						VerifyRequest(http.MethodPost, "/v3/domains"),
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

	Describe("CreateDomain for Private Domains", func() {
		var (
			domain     Domain
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			domain, warnings, executeErr = client.CreateDomain(Domain{Name: "some-name", Internal: types.NullBool{IsSet: false, Value: true}, OrganizationGUID: "organization-guid"})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-guid",
					"name": "some-name",
					"relationships": { 
						"organization": { 
							"data" : { 
								"guid" : "organization-guid"
							}
						}
					},
					"internal": false
				}`

				expectedRequestBody := `{
					"name": "some-name",
					"relationships": {
						"organization": {
							"data" : {
								"guid" : "organization-guid"
							}
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/domains"),
						VerifyJSON(expectedRequestBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given domain and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(domain).To(Equal(Domain{
					GUID:             "some-guid",
					Name:             "some-name",
					OrganizationGUID: "organization-guid",
					Internal:         types.NullBool{IsSet: true, Value: false},
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
						VerifyRequest(http.MethodPost, "/v3/domains"),
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

	Describe("DeleteDomain", func() {
		var (
			domainGUID   string
			jobURLString string
			jobURL       JobURL
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteDomain(domainGUID)
		})

		When("domain exists", func() {
			domainGUID = "domain-guid"
			jobURLString = "https://api.test.com/v3/jobs/job-guid"

			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/domains/domain-guid"),
						RespondWith(http.StatusAccepted, nil, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {jobURLString},
						}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(jobURL).To(Equal(JobURL(jobURLString)))
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
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/domains/domain-guid"),
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

	Describe("GetDomains", func() {
		var (
			query      Query
			domains    []Domain
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			domains, warnings, executeErr = client.GetDomains(query)
		})

		When("domains exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
		"pagination": {
			"next": {
				"href": "%s/v3/domains?page=2&per_page=2"
			}
		},
	  "resources": [
	    {
	      	"name": "domain-name-1",
	      	"guid": "domain-guid-1",
	      	"relationships": {
	            "organization": {
	                "data": {
	                    "guid": "owning-org-1"
	                }
	            }
	         }
	    },
	    {
	      	"name": "domain-name-2",
	      	"guid": "domain-guid-2",
			"relationships": {
	            "organization": {
	                "data": {
	                    "guid": "owning-org-2"
	                }
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
			"name": "domain-name-3",
	         "guid": "domain-guid-3",
			"relationships": {
	            "organization": {
	                "data": {
	                    "guid": "owning-org-3"
	                }
	            }
	         }
			}
		]
	}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/domains"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/domains", "page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{}
			})

			It("returns the queried domain and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(domains).To(ConsistOf(
					Domain{Name: "domain-name-1", GUID: "domain-guid-1", OrganizationGUID: "owning-org-1"},
					Domain{Name: "domain-name-2", GUID: "domain-guid-2", OrganizationGUID: "owning-org-2"},
					Domain{Name: "domain-name-3", GUID: "domain-guid-3", OrganizationGUID: "owning-org-3"},
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
						VerifyRequest(http.MethodGet, "/v3/domains"),
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

	Describe("GetOrganizationDomains", func() {
		var (
			orgGUID    string
			query      Query
			domains    []Domain
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			domains, warnings, executeErr = client.GetOrganizationDomains(orgGUID, query)
		})

		When("domains exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/organizations/some-org-guid/domains?organization_guids=some-org-guid&page=2&per_page=2"
		}
	},
  "resources": [
    {
      	"name": "domain-name-1",
      	"guid": "domain-guid-1",
      	"relationships": {
            "organization": {
                "data": {
                    "guid": "some-org-guid"
                }
            }
         }
    },
    {
      	"name": "domain-name-2",
      	"guid": "domain-guid-2",
		"relationships": {
            "organization": {
                "data": {
                    "guid": "some-org-guid"
                }
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
			"name": "domain-name-3",
			"guid": "domain-guid-3",
			"relationships": {
            	"organization": {
                	"data": {
                    	"guid": "some-org-guid"
                	}
            	}
         	}
		}
	]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid/domains", "organization_guids=some-org-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid/domains", "organization_guids=some-org-guid&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    OrganizationGUIDFilter,
					Values: []string{orgGUID},
				}
			})

			It("returns the queried domain and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(domains).To(ConsistOf(
					Domain{Name: "domain-name-1", GUID: "domain-guid-1", OrganizationGUID: orgGUID},
					Domain{Name: "domain-name-2", GUID: "domain-guid-2", OrganizationGUID: orgGUID},
					Domain{Name: "domain-name-3", GUID: "domain-guid-3", OrganizationGUID: orgGUID},
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
						VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid/domains"),
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

	Describe("SharePrivateDomainToOrgs", func() {
		var (
			orgGUID    = "some-org-guid"
			domainGUID = "some-domain-guid"
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.SharePrivateDomainToOrgs(
				domainGUID,
				SharedOrgs{GUIDs: []string{orgGUID}},
			)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{"data": 
								[{
									"guid": "some-org-guid"
								}]
						}`

				expectedBody := `{
					"data": [
						{"guid": "some-org-guid"}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/domains/some-domain-guid/relationships/shared_organizations"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(BeNil())
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
      		"detail": "Organization not found",
      		"title": "CF-ResourceNotFound"
    	}
  	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/domains/some-domain-guid/relationships/shared_organizations"),
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
							Detail: "Organization not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UnsharePrivateDomainFromOrg", func() {
		var (
			orgGUID    = "some-org-guid"
			domainGUID = "some-domain-guid"
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.UnsharePrivateDomainFromOrg(
				domainGUID,
				orgGUID,
			)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/domains/some-domain-guid/relationships/shared_organizations/some-org-guid"),
						RespondWith(http.StatusNoContent, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(BeNil())
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
      		"detail": "Organization not found",
      		"title": "CF-ResourceNotFound"
    	}
  	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/domains/some-domain-guid/relationships/shared_organizations/some-org-guid"),
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
							Detail: "Organization not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
