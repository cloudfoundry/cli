package ccv3_test

import (
	"code.cloudfoundry.org/cli/types"
	"fmt"
	"net/http"

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

	Describe("CreateDomain for Unscoped Domains", func() {
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
	})

	Describe("CreateDomain for Scoped Domains", func() {
		var (
			domain     Domain
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			domain, warnings, executeErr = client.CreateDomain(Domain{Name: "some-name", Internal: types.NullBool{IsSet: false, Value: true}, OrganizationGuid: "organization-guid"})
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
					OrganizationGuid: "organization-guid",
					Internal:         types.NullBool{IsSet: true, Value: false},
				}))
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
					Domain{Name: "domain-name-1", GUID: "domain-guid-1", OrganizationGuid: "owning-org-1"},
					Domain{Name: "domain-name-2", GUID: "domain-guid-2", OrganizationGuid: "owning-org-2"},
					Domain{Name: "domain-name-3", GUID: "domain-guid-3", OrganizationGuid: "owning-org-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})
	})
})
