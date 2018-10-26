package ccv2_test

import (
	"fmt"
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

	Describe("CreateSharedDomain", func() {
		var (
			domain          string
			routerGroupGUID string
		)

		When("no errors are encountered", func() {
			BeforeEach(func() {
				response := `{
											"metadata": {
												"guid": "43436c2d-2b4f-45c2-9f50-e530e1cedba6",
												"url": "/v2/shared_domains/43436c2d-2b4f-45c2-9f50-e530e1cedba6",
												"created_at": "2016-06-08T16:41:37Z",
												"updated_at": "2016-06-08T16:41:26Z"
											},
											"entity": {
												"name": "example.com",
												"internal": false,
												"router_group_guid": "some-guid",
												"router_group_type": "tcp"
											}
										}
										`
				domain = "some-domain-name.com"
				routerGroupGUID = "some-guid"
				body := fmt.Sprintf(`{"name":"%s","router_group_guid":"%s"}`, domain, routerGroupGUID)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/shared_domains"),
						VerifyBody([]byte(body)),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1,warning-2"}}),
					))
			})

			It("should call the API and return all warnings", func() {
				warnings, err := client.CreateSharedDomain(domain, routerGroupGUID)
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the API returns an unauthorized error", func() {
			BeforeEach(func() {
				response := `{
											"description": "You are not authorized to perform the requested action",
											"error_code": "CF-NotAuthorized",
											"code": 10003
										}`
				domain = "some-domain-name.com"
				body := fmt.Sprintf(`{"name":"%s"}`, domain)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/shared_domains"),
						VerifyBody([]byte(body)),
						RespondWith(http.StatusForbidden, response, http.Header{"X-Cf-Warnings": {"this is your final warning"}}),
					))
			})

			It("should return the error and all warnings", func() {
				warnings, err := client.CreateSharedDomain(domain, "")
				Expect(warnings).To(ConsistOf("this is your final warning"))
				Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "You are not authorized to perform the requested action"}))
			})
		})
	})

	Describe("GetSharedDomain", func() {
		When("the shared domain exists", func() {
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

		When("the shared domain does not exist", func() {
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
		When("the private domain exists", func() {
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

		When("the private domain does not exist", func() {
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

	Describe("GetPrivateDomains", func() {
		When("the cloud controller does not return an error", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/private_domains?q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "domain-guid-1"
							},
							"entity": {
								"name": "domain-name-1"
							}
						},
						{
							"metadata": {
								"guid": "domain-guid-2"
							},
							"entity": {
								"name": "domain-name-2"
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
								"name": "domain-name-3"
							}
						},
						{
							"metadata": {
								"guid": "domain-guid-4"
							},
							"entity": {
								"name": "domain-name-4"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains", "q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains", "q=name%20IN%20domain-name-1,domain-name-2,domain-name-3,domain-name-4&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the private domains and warnings", func() {
				domains, warnings, err := client.GetPrivateDomains(Filter{
					Type:     constant.NameFilter,
					Operator: constant.InOperator,
					Values:   []string{"domain-name-1", "domain-name-2", "domain-name-3", "domain-name-4"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(Equal([]Domain{
					{
						GUID: "domain-guid-1",
						Name: "domain-name-1",
						Type: constant.PrivateDomain,
					},
					{
						GUID: "domain-guid-2",
						Name: "domain-name-2",
						Type: constant.PrivateDomain,
					},
					{
						GUID: "domain-guid-3",
						Name: "domain-name-3",
						Type: constant.PrivateDomain,
					},
					{
						GUID: "domain-guid-4",
						Name: "domain-name-4",
						Type: constant.PrivateDomain,
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the warnings and error", func() {
				domains, warnings, err := client.GetPrivateDomains()
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

	Describe("GetSharedDomains", func() {
		When("the cloud controller does not return an error", func() {
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
				domains, warnings, err := client.GetSharedDomains(Filter{
					Type:     constant.NameFilter,
					Operator: constant.InOperator,
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

		When("the cloud controller returns an error", func() {
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
		When("the cloud controller does not return an error", func() {
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

		When("the client includes includes query parameters for name", func() {
			It("it includes the query parameters in the request", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/organizations/some-org-guid/private_domains", "q=name:private-domain-name"),
						RespondWith(http.StatusOK, ""),
					),
				)

				client.GetOrganizationPrivateDomains("some-org-guid", Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"private-domain-name"},
				})
			})
		})

		When("the cloud controller returns an error", func() {
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
