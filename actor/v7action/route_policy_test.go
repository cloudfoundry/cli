package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	. "code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Policy Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("AddRoutePolicy", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.AddRoutePolicy("apps.example.com", "cf:any", "myapp", "")
		})

		When("the API calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid", EnforceRoutePolicies: types.NullBool{IsSet: true, Value: true}}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid", Host: "myapp"}},
					ccv3.Warnings{"routes-warning"},
					nil,
				)
				fakeCloudControllerClient.CreateRoutePolicyReturns(
					resources.RoutePolicy{GUID: "policy-guid"},
					ccv3.Warnings{"create-warning"},
					nil,
				)
			})

			It("creates the route policy and returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning", "create-warning"))

				Expect(fakeCloudControllerClient.CreateRoutePolicyCallCount()).To(Equal(1))
				passedPolicy := fakeCloudControllerClient.CreateRoutePolicyArgsForCall(0)
				Expect(passedPolicy.Source).To(Equal("cf:any"))
				Expect(passedPolicy.RouteGUID).To(Equal("route-guid"))
			})
		})

		When("the domain does not enforce route policies", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid", EnforceRoutePolicies: types.NullBool{IsSet: false}}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
			})

			It("returns a DomainNotEnforcingRoutePoliciesError and the domain warning", func() {
				Expect(executeErr).To(MatchError(actionerror.DomainNotEnforcingRoutePoliciesError{Name: "apps.example.com"}))
				Expect(warnings).To(ConsistOf("domain-warning"))
				Expect(fakeCloudControllerClient.CreateRoutePolicyCallCount()).To(Equal(0))
			})
		})

		When("getting the domain fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"domain-warning"},
					errors.New("domain-error"),
				)
			})

			It("returns the error and domain warning", func() {
				Expect(executeErr).To(MatchError("domain-error"))
				Expect(warnings).To(ConsistOf("domain-warning"))
			})
		})

		When("getting routes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid", EnforceRoutePolicies: types.NullBool{IsSet: true, Value: true}}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(nil, ccv3.Warnings{"routes-warning"}, errors.New("routes-error"))
			})

			It("returns the error and all collected warnings", func() {
				Expect(executeErr).To(MatchError("routes-error"))
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning"))
			})
		})

		When("the route is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid", EnforceRoutePolicies: types.NullBool{IsSet: true, Value: true}}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{},
					ccv3.Warnings{"routes-warning"},
					nil,
				)
			})

			It("returns a RouteNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					Host:       "myapp",
					DomainName: "apps.example.com",
					Path:       "",
				}))
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning"))
			})
		})

		When("creating the route policy fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid", EnforceRoutePolicies: types.NullBool{IsSet: true, Value: true}}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid", Host: "myapp"}},
					ccv3.Warnings{"routes-warning"},
					nil,
				)
				fakeCloudControllerClient.CreateRoutePolicyReturns(
					resources.RoutePolicy{},
					ccv3.Warnings{"create-warning"},
					errors.New("create-error"),
				)
			})

			It("returns the error and all collected warnings", func() {
				Expect(executeErr).To(MatchError("create-error"))
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning", "create-warning"))
			})
		})
	})

	Describe("DeleteRoutePolicyBySource", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteRoutePolicyBySource("apps.example.com", "cf:any", "myapp", "")
		})

		When("the API calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid", Host: "myapp"}},
					ccv3.Warnings{"routes-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-guid", Source: "cf:any", RouteGUID: "route-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-policies-warning"},
					nil,
				)
				fakeCloudControllerClient.DeleteRoutePolicyReturns(
					"",
					ccv3.Warnings{"delete-warning"},
					nil,
				)
			})

			It("deletes the route policy and returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning", "get-policies-warning", "delete-warning"))

				Expect(fakeCloudControllerClient.DeleteRoutePolicyCallCount()).To(Equal(1))
				deletedGUID := fakeCloudControllerClient.DeleteRoutePolicyArgsForCall(0)
				Expect(deletedGUID).To(Equal("policy-guid"))
			})
		})

		When("getting the domain fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"domain-warning"},
					errors.New("domain-error"),
				)
			})

			It("returns the error and domain warning", func() {
				Expect(executeErr).To(MatchError("domain-error"))
				Expect(warnings).To(ConsistOf("domain-warning"))
			})
		})

		When("getting routes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(nil, ccv3.Warnings{"routes-warning"}, errors.New("routes-error"))
			})

			It("returns the error and all collected warnings", func() {
				Expect(executeErr).To(MatchError("routes-error"))
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning"))
			})
		})

		When("the route is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{"domain-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{},
					ccv3.Warnings{"routes-warning"},
					nil,
				)
			})

			It("returns a RouteNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					Host:       "myapp",
					DomainName: "apps.example.com",
					Path:       "",
				}))
				Expect(warnings).To(ConsistOf("domain-warning", "routes-warning"))
			})
		})

		When("getting route policies fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid", Host: "myapp"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					nil,
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-policies-warning"},
					errors.New("get-policies-error"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-policies-error"))
				Expect(warnings).To(ConsistOf("get-policies-warning"))
			})
		})

		When("no matching policy is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid", Host: "myapp"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "other-policy-guid", Source: "cf:app:some-app-guid", RouteGUID: "route-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns a RoutePolicyNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RoutePolicyNotFoundError{Source: "cf:any"}))
			})
		})
	})

	Describe("GetRoutePoliciesForSpace", func() {
		var (
			policiesWithRoutes []RoutePolicyWithRoute
			warnings           Warnings
			executeErr         error
			domainFilter       string
			hostnameFilter     string
			pathFilter         string
		)

		BeforeEach(func() {
			domainFilter = ""
			hostnameFilter = ""
			pathFilter = ""
		})

		JustBeforeEach(func() {
			policiesWithRoutes, warnings, executeErr = actor.GetRoutePoliciesForSpace(
				"space-guid", domainFilter, hostnameFilter, pathFilter, "",
			)
		})

		When("the API calls are successful (cf:any source)", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "p-guid", Source: "cf:any", RouteGUID: "r-guid"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "r-guid", Host: "backend", DomainGUID: "domain-guid"},
						},
					},
					ccv3.Warnings{"get-policies-warning"},
					nil,
				)
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{Name: "apps.example.com", GUID: "domain-guid"},
					ccv3.Warnings{},
					nil,
				)
			})

			It("sends ?include=route,source and returns policies with route info", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-policies-warning"))
				Expect(policiesWithRoutes).To(HaveLen(1))
				Expect(policiesWithRoutes[0].GUID).To(Equal("p-guid"))
				Expect(policiesWithRoutes[0].Source).To(Equal("cf:any"))
				Expect(policiesWithRoutes[0].Route.Host).To(Equal("backend"))
				Expect(policiesWithRoutes[0].DomainName).To(Equal("apps.example.com"))
				Expect(policiesWithRoutes[0].ScopeType).To(Equal("any"))

				queries := fakeCloudControllerClient.GetRoutePoliciesArgsForCall(0)
				Expect(queries).To(ContainElement(ccv3.Query{Key: ccv3.Include, Values: []string{"route,source"}}))

				Expect(fakeCloudControllerClient.GetDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainArgsForCall(0)).To(Equal("domain-guid"))
			})
		})

		When("source names are resolved from CAPI included resources", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "p-app", Source: "cf:app:app-guid", RouteGUID: "r-guid"},
						{GUID: "p-space", Source: "cf:space:space-guid", RouteGUID: "r-guid"},
						{GUID: "p-org", Source: "cf:org:org-guid", RouteGUID: "r-guid"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "r-guid", Host: "backend", DomainGUID: "domain-guid"},
						},
						Apps:          []resources.Application{{GUID: "app-guid", Name: "my-app"}},
						Spaces:        []resources.Space{{GUID: "space-guid", Name: "my-space"}},
						Organizations: []resources.Organization{{GUID: "org-guid", Name: "my-org"}},
					},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{Name: "apps.example.com", GUID: "domain-guid"},
					ccv3.Warnings{},
					nil,
				)
			})

			It("resolves source names from included resources without additional API calls", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(policiesWithRoutes).To(HaveLen(3))

				Expect(policiesWithRoutes[0].ScopeType).To(Equal("app"))
				Expect(policiesWithRoutes[0].SourceName).To(Equal("my-app"))

				Expect(policiesWithRoutes[1].ScopeType).To(Equal("space"))
				Expect(policiesWithRoutes[1].SourceName).To(Equal("my-space"))

				Expect(policiesWithRoutes[2].ScopeType).To(Equal("org"))
				Expect(policiesWithRoutes[2].SourceName).To(Equal("my-org"))

				// No additional API calls for name resolution
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(0))
			})
		})

		When("there are no route policies", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns an empty slice and no error", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(policiesWithRoutes).To(BeEmpty())
			})
		})

		When("getting route policies fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					nil,
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-policies-warning"},
					errors.New("get-policies-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("get-policies-error"))
				Expect(warnings).To(ConsistOf("get-policies-warning"))
			})
		})

		When("filtering by domain name", func() {
			BeforeEach(func() {
				domainFilter = "apps.example.com"
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "p-match", Source: "cf:any", RouteGUID: "r-match"},
						{GUID: "p-other", Source: "cf:any", RouteGUID: "r-other"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "r-match", Host: "backend", DomainGUID: "target-domain-guid"},
							{GUID: "r-other", Host: "other", DomainGUID: "other-domain-guid"},
						},
					},
					ccv3.Warnings{"get-policies-warning"},
					nil,
				)
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "apps.example.com", GUID: "target-domain-guid"}},
					ccv3.Warnings{"domain-filter-warning"},
					nil,
				)
			})

			It("returns only matching policies with domain name from cache, no GetDomain call", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(policiesWithRoutes).To(HaveLen(1))
				Expect(policiesWithRoutes[0].GUID).To(Equal("p-match"))
				Expect(policiesWithRoutes[0].DomainName).To(Equal("apps.example.com"))
				// Cache pre-populated from GetDomainByName — GetDomain must not be called
				Expect(fakeCloudControllerClient.GetDomainCallCount()).To(Equal(0))
				Expect(warnings).To(ConsistOf("get-policies-warning", "domain-filter-warning"))
			})

			When("GetDomainByName fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						nil,
						ccv3.Warnings{"domain-filter-warning"},
						errors.New("domain-lookup-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("domain-lookup-error"))
					Expect(warnings).To(ContainElement("domain-filter-warning"))
				})
			})
		})

		When("filtering by hostname", func() {
			BeforeEach(func() {
				hostnameFilter = "backend"
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "p-match", Source: "cf:any", RouteGUID: "r-match"},
						{GUID: "p-other", Source: "cf:any", RouteGUID: "r-other"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "r-match", Host: "backend", DomainGUID: "domain-guid"},
							{GUID: "r-other", Host: "other", DomainGUID: "domain-guid"},
						},
					},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{Name: "apps.example.com", GUID: "domain-guid"},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns only policies for routes matching the hostname", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(policiesWithRoutes).To(HaveLen(1))
				Expect(policiesWithRoutes[0].GUID).To(Equal("p-match"))
			})
		})

		When("filtering by path", func() {
			BeforeEach(func() {
				pathFilter = "/api"
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "p-match", Source: "cf:any", RouteGUID: "r-match"},
						{GUID: "p-other", Source: "cf:any", RouteGUID: "r-other"},
					},
					ccv3.IncludedResources{
						Routes: []resources.Route{
							{GUID: "r-match", Host: "backend", Path: "/api", DomainGUID: "domain-guid"},
							{GUID: "r-other", Host: "backend", Path: "/other", DomainGUID: "domain-guid"},
						},
					},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{Name: "apps.example.com", GUID: "domain-guid"},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns only policies for routes matching the path", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(policiesWithRoutes).To(HaveLen(1))
				Expect(policiesWithRoutes[0].GUID).To(Equal("p-match"))
			})
		})
	})
})
