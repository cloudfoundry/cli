package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Domain", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSharedDomain", func() {
		Context("when the shared domain exists", func() {
			BeforeEach(func() {
				response := `{
						"metadata": {
							"guid": "shared-domain-guid",
							"updated_at": null
						},
						"entity": {
							"name": "shared-domain-1.com",
							"router_group_guid": "some-router-group-guid",
							"router_group_type": "http"
						}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains/shared-domain-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the shared domain and all warnings", func() {
				domain, warnings, err := client.GetSharedDomain("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{
					Name:            "shared-domain-1.com",
					GUID:            "shared-domain-guid",
					RouterGroupGUID: "some-router-group-guid",
					RouterGroupType: constant.HTTPRouterGroup,
					Type:            constant.SharedDomain,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the shared domain does not exist", func() {
			BeforeEach(func() {
				response := `{
					"code": 130002,
					"description": "The domain could not be found: shared-domain-guid",
					"error_code": "CF-DomainNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains/shared-domain-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				domain, _, err := client.GetSharedDomain("shared-domain-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The domain could not be found: shared-domain-guid",
				}))
				Expect(domain).To(Equal(Domain{}))
			})
		})
	})

	Describe("GetPrivateDomain", func() {
		Context("when the private domain exists", func() {
			BeforeEach(func() {
				response := `{
						"metadata": {
							"guid": "private-domain-guid",
							"updated_at": null
						},
						"entity": {
							"name": "private-domain-1.com"
						}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains/private-domain-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the private domain and all warnings", func() {
				domain, warnings, err := client.GetPrivateDomain("private-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{
					Name: "private-domain-1.com",
					GUID: "private-domain-guid",
					Type: constant.PrivateDomain,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the private domain does not exist", func() {
			BeforeEach(func() {
				response := `{
					"code": 130002,
					"description": "The domain could not be found: private-domain-guid",
					"error_code": "CF-DomainNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains/private-domain-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				domain, _, err := client.GetPrivateDomain("private-domain-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The domain could not be found: private-domain-guid",
				}))
				Expect(domain).To(Equal(Domain{}))
			})
		})
	})

	Describe("GetSharedDomains", func() {
		Context("when the cloud controller does not return an error", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/shared_domains?q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "domain-guid-1"
							},
							"entity": {
								"name": "domain-name-1",
								"router_group_guid": "some-router-group-guid-1",
								"router_group_type": "http"
							}
						},
						{
							"metadata": {
								"guid": "domain-guid-2"
							},
							"entity": {
								"name": "domain-name-2",
								"router_group_guid": "some-router-group-guid-2",
								"router_group_type": "http"
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "domain-guid-3"
							},
							"entity": {
								"name": "domain-name-3",
								"router_group_guid": "some-router-group-guid-3",
								"router_group_type": "http"
							}
						},
						{
							"metadata": {
								"guid": "domain-guid-4"
							},
							"entity": {
								"name": "domain-name-4",
								"router_group_guid": "some-router-group-guid-4",
								"router_group_type": "http"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains", "q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains", "q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the shared domain and warnings", func() {
				domains, warnings, err := client.GetSharedDomains(QQuery{
					Filter:   NameFilter,
					Operator: InOperator,
					Values:   []string{"domain-name-1", "domain-name-2", "domain-name-3", "domain-name-4"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(Equal([]Domain{
					{
						GUID:            "domain-guid-1",
						Name:            "domain-name-1",
						RouterGroupGUID: "some-router-group-guid-1",
						RouterGroupType: constant.HTTPRouterGroup,
						Type:            constant.SharedDomain,
					},
					{
						GUID:            "domain-guid-2",
						Name:            "domain-name-2",
						RouterGroupGUID: "some-router-group-guid-2",
						RouterGroupType: constant.HTTPRouterGroup,
						Type:            constant.SharedDomain,
					},
					{
						GUID:            "domain-guid-3",
						Name:            "domain-name-3",
						RouterGroupGUID: "some-router-group-guid-3",
						RouterGroupType: constant.HTTPRouterGroup,
						Type:            constant.SharedDomain,
					},
					{
						GUID:            "domain-guid-4",
						Name:            "domain-name-4",
						RouterGroupGUID: "some-router-group-guid-4",
						RouterGroupType: constant.HTTPRouterGroup,
						Type:            constant.SharedDomain,
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when the cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the warnings and error", func() {
				domains, warnings, err := client.GetSharedDomains()
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        1,
						Description: "some error description",
						ErrorCode:   "CF-SomeError",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(domains).To(Equal([]Domain{}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetOrganizationPrivateDomains", func() {
		Context("when the cloud controller does not return an error", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/organizations/some-org-guid/private_domains?page=2",
					"resources": [
						{
							"metadata": {
								"guid": "private-domain-guid-1"
							},
							"entity": {
								"name": "private-domain-name-1"
							}
						},
						{
							"metadata": {
								"guid": "private-domain-guid-2"
							},
							"entity": {
								"name": "private-domain-name-2"
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "private-domain-guid-3"
							},
							"entity": {
								"name": "private-domain-name-3"
							}
						},
						{
							"metadata": {
								"guid": "private-domain-guid-4"
							},
							"entity": {
								"name": "private-domain-name-4"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/private_domains"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/private_domains", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the domains and warnings", func() {
				domains, warnings, err := client.GetOrganizationPrivateDomains("some-org-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(Equal([]Domain{
					{
						Name: "private-domain-name-1",
						GUID: "private-domain-guid-1",
						Type: constant.PrivateDomain,
					},
					{
						Name: "private-domain-name-2",
						GUID: "private-domain-guid-2",
						Type: constant.PrivateDomain,
					},
					{
						Name: "private-domain-name-3",
						GUID: "private-domain-guid-3",
						Type: constant.PrivateDomain,
					},
					{
						Name: "private-domain-name-4",
						GUID: "private-domain-guid-4",
						Type: constant.PrivateDomain,
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when the client includes includes query parameters for name", func() {
			It("it includes the query parameters in the request", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/private_domains", "q=name:private-domain-name"),
						RespondWith(http.StatusOK, ""),
					),
				)

				client.GetOrganizationPrivateDomains("some-org-guid", QQuery{
					Filter:   NameFilter,
					Operator: EqualOperator,
					Values:   []string{"private-domain-name"},
				})
			})
		})

		Context("when the cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					   "description": "The organization could not be found: glah",
					   "error_code": "CF-OrganizationNotFound",
					   "code": 30003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/private_domains"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the warnings and error", func() {
				domains, warnings, err := client.GetOrganizationPrivateDomains("some-org-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The organization could not be found: glah",
				}))
				Expect(domains).To(Equal([]Domain{}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
