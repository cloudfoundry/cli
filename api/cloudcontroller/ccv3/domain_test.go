package ccv3_test

import (
	"code.cloudfoundry.org/cli/types"
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

})
