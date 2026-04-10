package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	. "code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Access Rule Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("GetAccessRulesForSpace", func() {
		var (
			spaceGUID     string
			domainName    string
			hostname      string
			path          string
			labelSelector string

			results    []AccessRuleWithRoute
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			spaceGUID = "space-guid-1"
			domainName = ""
			hostname = ""
			path = ""
			labelSelector = ""
		})

		JustBeforeEach(func() {
			results, warnings, executeErr = actor.GetAccessRulesForSpace(
				spaceGUID,
				domainName,
				hostname,
				path,
				labelSelector,
			)
		})

		When("getting access rules succeeds with multiple rules", func() {
			BeforeEach(func() {
				// Mock GetAccessRules call with included routes
				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{
							GUID:      "rule-guid-1",
							Name:      "rule-1",
							Selector:  "cf:app:app-guid-1",
							RouteGUID: "route-guid-1",
						},
						{
							GUID:      "rule-guid-2",
							Name:      "rule-2",
							Selector:  "cf:any",
							RouteGUID: "route-guid-2",
						},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{
								GUID:       "route-guid-1",
								SpaceGUID:  "space-guid-1",
								DomainGUID: "domain-guid-1",
								Host:       "app1",
								Path:       "/path1",
							},
							{
								GUID:       "route-guid-2",
								SpaceGUID:  "space-guid-1",
								DomainGUID: "domain-guid-2",
								Host:       "app2",
								Path:       "",
							},
						},
					},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				// Mock GetDomain calls for domain name resolution
				fakeCloudControllerClient.GetDomainStub = func(guid string) (resources.Domain, ccv3.Warnings, error) {
					switch guid {
					case "domain-guid-1":
						return resources.Domain{GUID: "domain-guid-1", Name: "example.com"}, ccv3.Warnings{"get-domain-warning-1"}, nil
					case "domain-guid-2":
						return resources.Domain{GUID: "domain-guid-2", Name: "test.com"}, ccv3.Warnings{"get-domain-warning-2"}, nil
					default:
						return resources.Domain{}, nil, errors.New("domain not found")
					}
				}

				// Mock GetApplications for app name resolution
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "app-guid-1", Name: "my-app"},
					},
					ccv3.Warnings{"get-app-warning"},
					nil,
				)
			})

			It("returns access rules with route and domain information", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"get-access-rules-warning",
					"get-domain-warning-1",
					"get-domain-warning-2",
					"get-app-warning",
				))

				Expect(results).To(HaveLen(2))

				// First rule
			Expect(results[0].GUID).To(Equal("rule-guid-1"))
			Expect(results[0].Name).To(Equal("rule-1"))
			Expect(results[0].Selector).To(Equal("cf:app:app-guid-1"))
			Expect(results[0].Route.GUID).To(Equal("route-guid-1"))
			Expect(results[0].Route.Host).To(Equal("app1"))
			Expect(results[0].Route.Path).To(Equal("/path1"))
			Expect(results[0].DomainName).To(Equal("example.com"))
			Expect(results[0].ScopeType).To(Equal("app"))
			Expect(results[0].SourceName).To(Equal("my-app"))

			// Second rule
			Expect(results[1].GUID).To(Equal("rule-guid-2"))
			Expect(results[1].Name).To(Equal("rule-2"))
			Expect(results[1].Selector).To(Equal("cf:any"))
			Expect(results[1].Route.GUID).To(Equal("route-guid-2"))
			Expect(results[1].Route.Host).To(Equal("app2"))
			Expect(results[1].DomainName).To(Equal("test.com"))
			Expect(results[1].ScopeType).To(Equal("any"))
			Expect(results[1].SourceName).To(Equal(""))
			})

			It("calls GetAccessRules with space GUID and include route filters", func() {
				Expect(fakeCloudControllerClient.GetAccessRulesCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetAccessRulesArgsForCall(0)
				Expect(queries).To(ContainElement(ccv3.Query{
					Key:    ccv3.SpaceGUIDFilter,
					Values: []string{"space-guid-1"},
				}))
				Expect(queries).To(ContainElement(ccv3.Query{
					Key:    ccv3.Include,
					Values: []string{"route"},
				}))
			})

			It("does not call GetRoutes separately", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
			})
		})

		When("domain name filter is provided", func() {
			BeforeEach(func() {
				domainName = "example.com"

				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{
							GUID:      "rule-guid-1",
							Name:      "rule-1",
							Selector:  "cf:any",
							RouteGUID: "route-guid-1",
						},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{
								GUID:       "route-guid-1",
								SpaceGUID:  "space-guid-1",
								DomainGUID: "domain-guid-1",
								Host:       "app1",
							},
						},
					},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{GUID: "domain-guid-1", Name: "example.com"},
					ccv3.Warnings{"get-domain-warning"},
					nil,
				)
			})

			It("filters routes by domain GUID", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				// Routes are filtered in-memory from included resources
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
			})
		})

		When("hostname filter is provided", func() {
			BeforeEach(func() {
				hostname = "myapp"

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("adds hostname filter to route query", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				// GetRoutes should not be called since routes come from included resources
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
			})
		})

		When("path filter is provided", func() {
			BeforeEach(func() {
				path = "/api"

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("adds path filter to route query", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				// GetRoutes should not be called if no access rules are found
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
			})
		})

		When("label selector filter is provided", func() {
			BeforeEach(func() {
				labelSelector = "env=production"

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("adds label selector filter to access rules query", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.GetAccessRulesCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetAccessRulesArgsForCall(0)

				Expect(queries).To(ContainElement(ccv3.Query{
					Key:    ccv3.LabelSelectorFilter,
					Values: []string{"env=production"},
				}))
			})
		})

		When("no access rules are found in the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("returns an empty list without error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(warnings).To(ConsistOf("get-access-rules-warning"))
			})

			It("does not call GetRoutes", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
			})
		})

		When("getting access rules fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetAccessRulesReturns(
					nil,
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					errors.New("api error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("api error"))
				Expect(warnings).To(ConsistOf("get-access-rules-warning"))
				Expect(results).To(BeNil())
			})
		})

		When("getting domain by name fails", func() {
			BeforeEach(func() {
				domainName = "invalid-domain.com"

				// Mock GetAccessRules to return at least one access rule
				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{GUID: "access-rule-guid-1", RouteGUID: "route-guid-1"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "route-guid-1", DomainGUID: "domain-guid-1"},
						},
					},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"get-domains-warning"},
					actionerror.DomainNotFoundError{Name: "invalid-domain.com"},
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "invalid-domain.com"}))
				Expect(warnings).To(ConsistOf("get-access-rules-warning", "get-domains-warning"))
				Expect(results).To(BeNil())
			})
		})

		When("resolving domain name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{
							GUID:      "rule-guid-1",
							Name:      "rule-1",
							Selector:  "cf:any",
							RouteGUID: "route-guid-1",
						},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{
								GUID:       "route-guid-1",
								SpaceGUID:  "space-guid-1",
								DomainGUID: "domain-guid-1",
								Host:       "app1",
							},
						},
					},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				// Domain lookup fails
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{},
					ccv3.Warnings{"get-domain-warning"},
					errors.New("domain lookup error"),
				)
			})

			It("uses the domain GUID as fallback and continues", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].DomainName).To(Equal("domain-guid-1"))
				Expect(warnings).To(ContainElement("get-domain-warning"))
			})
		})

		When("resolving target name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{
							GUID:      "rule-guid-1",
							Name:      "rule-1",
							Selector:  "cf:app:app-guid-1",
							RouteGUID: "route-guid-1",
						},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{
								GUID:       "route-guid-1",
								SpaceGUID:  "space-guid-1",
								DomainGUID: "domain-guid-1",
								Host:       "app1",
							},
						},
					},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{GUID: "domain-guid-1", Name: "example.com"},
					ccv3.Warnings{"get-domain-warning"},
					nil,
				)

				// App lookup fails
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv3.Warnings{"get-app-warning"},
					errors.New("app lookup error"),
				)
			})

		It("leaves source name blank and populates scope type", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].ScopeType).To(Equal("app"))
			Expect(results[0].SourceName).To(Equal(""))
			Expect(warnings).To(ContainElement("get-app-warning"))
		})
		})
	})

	// Note: resolveAccessRuleTarget and splitSelector are unexported methods
	// and are tested indirectly through GetAccessRulesForSpace above.

	Describe("GetAccessRulesByRoute", func() {
		var (
			domainName string
			hostname   string
			path       string

			rules      []resources.AccessRule
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			domainName = "example.com"
			hostname = "myapp"
			path = ""
		})

		JustBeforeEach(func() {
			rules, warnings, executeErr = actor.GetAccessRulesByRoute(domainName, hostname, path)
		})

		When("the route exists with access rules", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "space-guid-1",
							DomainGUID: "domain-guid-1",
							Host:       "myapp",
						},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{GUID: "rule-guid-1", Name: "rule-1", Selector: "cf:any"},
						{GUID: "rule-guid-2", Name: "rule-2", Selector: "cf:app:app-guid-1"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("returns the access rules", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(rules).To(HaveLen(2))
				Expect(rules[0].Name).To(Equal("rule-1"))
				Expect(rules[1].Name).To(Equal("rule-2"))
				Expect(warnings).To(ConsistOf(
					"get-domains-warning",
					"get-routes-warning",
					"get-access-rules-warning",
				))
			})
		})

		When("the route does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
			})

			It("returns a RouteNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					Host:       "myapp",
					DomainName: "example.com",
					Path:       "",
				}))
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning"))
			})
		})
	})

	Describe("AddAccessRule", func() {
		var (
			ruleName   string
			domainName string
			selector   string
			hostname   string
			path       string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			ruleName = "my-rule"
			domainName = "example.com"
			selector = "cf:app:app-guid-1"
			hostname = "myapp"
			path = ""
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.AddAccessRule(ruleName, domainName, selector, hostname, path)
		})

		When("creating the access rule succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "space-guid-1",
							DomainGUID: "domain-guid-1",
							Host:       "myapp",
						},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)

				fakeCloudControllerClient.CreateAccessRuleReturns(
					resources.AccessRule{GUID: "rule-guid-1", Name: "my-rule"},
					ccv3.Warnings{"create-rule-warning"},
					nil,
				)
			})

			It("creates the access rule and returns warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"get-domains-warning",
					"get-routes-warning",
					"create-rule-warning",
				))

				Expect(fakeCloudControllerClient.CreateAccessRuleCallCount()).To(Equal(1))
				rule := fakeCloudControllerClient.CreateAccessRuleArgsForCall(0)
				Expect(rule.Name).To(Equal("my-rule"))
				Expect(rule.Selector).To(Equal("cf:app:app-guid-1"))
				Expect(rule.RouteGUID).To(Equal("route-guid-1"))
			})
		})

		When("the route does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
			})

			It("returns a RouteNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					Host:       "myapp",
					DomainName: "example.com",
					Path:       "",
				}))
			})
		})
	})

	Describe("DeleteAccessRule", func() {
		var (
			ruleName   string
			domainName string
			hostname   string
			path       string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			ruleName = "my-rule"
			domainName = "example.com"
			hostname = "myapp"
			path = ""
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteAccessRule(ruleName, domainName, hostname, path)
		})

		When("the access rule exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "space-guid-1",
							DomainGUID: "domain-guid-1",
							Host:       "myapp",
						},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{GUID: "rule-guid-1", Name: "my-rule", Selector: "cf:any"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteAccessRuleReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"delete-rule-warning"},
					nil,
				)
			})

			It("deletes the access rule and returns warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"get-domains-warning",
					"get-routes-warning",
					"get-access-rules-warning",
					"delete-rule-warning",
				))

				Expect(fakeCloudControllerClient.DeleteAccessRuleCallCount()).To(Equal(1))
				guid := fakeCloudControllerClient.DeleteAccessRuleArgsForCall(0)
				Expect(guid).To(Equal("rule-guid-1"))
			})
		})

		When("the access rule does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid-1", Name: "example.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "space-guid-1",
							DomainGUID: "domain-guid-1",
							Host:       "myapp",
						},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)

				fakeCloudControllerClient.GetAccessRulesReturns(
					[]resources.AccessRule{
						{GUID: "rule-guid-other", Name: "other-rule", Selector: "cf:any"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-access-rules-warning"},
					nil,
				)
			})

			It("returns an AccessRuleNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.AccessRuleNotFoundError{Name: "my-rule"}))
				Expect(warnings).To(ConsistOf(
					"get-domains-warning",
					"get-routes-warning",
					"get-access-rules-warning",
				))
			})
		})
	})
})
