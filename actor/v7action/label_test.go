package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	. "code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("labels", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig
		warnings                  Warnings
		executeErr                error
		resourceName              string
		spaceGUID                 string
		orgGUID                   string
		labels                    map[string]types.NullString
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil, nil, nil)
		resourceName = "some-resource"
		orgGUID = "some-org-guid"
		spaceGUID = "some-space-guid"
	})

	Describe("UpdateApplicationLabelsByApplicationName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateApplicationLabelsByApplicationName(resourceName, spaceGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-app-labels-warnings"},
					nil,
				)
			})

			It("gets the application", func() {
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
				))
			})

			It("sets the app labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, appGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("app"))
				Expect(appGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
			})
		})

		When("there are client errors", func() {
			When("GetApplications fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]resources.Application{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-apps-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-apps-error"))
				})
			})

			When("UpdateApplication fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]resources.Application{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-app-labels-warnings"},
						errors.New("update-application-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
					Expect(executeErr).To(MatchError("update-application-error"))
				})
			})
		})
	})

	Describe("UpdateDomainLabelsByDomainName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateDomainLabelsByDomainName(resourceName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"warning-updating-metadata"},
					nil,
				)
			})

			It("gets the domain", func() {
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("sets the domain labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, domainGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("domain"))
				Expect(domainGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-updating-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the domain fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-domains-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-domains-error"))
				})
			})

			When("updating the domain fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"warning-updating-metadata"},
						errors.New("update-domain-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-updating-metadata"))
					Expect(executeErr).To(MatchError("update-domain-error"))
				})
			})
		})
	})

	Describe("UpdateOrganizationLabelsByOrganizationName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateOrganizationLabelsByOrganizationName(resourceName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-org"},
					nil,
				)
			})

			It("gets the organization", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("sets the org labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, orgGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("org"))
				Expect(orgGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-org"))
			})
		})

		When("there are client errors", func() {
			When("fetching the organization fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]resources.Organization{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-orgs-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-orgs-error"))
				})
			})

			When("updating the organization fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]resources.Organization{{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-org"},
						errors.New("update-orgs-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-org"))
					Expect(executeErr).To(MatchError("update-orgs-error"))
				})
			})
		})
	})

	Describe("UpdateRouteLabels", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateRouteLabels("sub.example.com/my-route/path", "space-guid", labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", URL: "hostname.domain-name", Path: "/the-path"},
					},
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					nil,
				)

				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-route-warning"},
					nil,
				)
			})

			It("gets the domain", func() {
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"sub.example.com"}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("gets the route", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}},
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{"domain-guid"}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{""}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{"/my-route/path"}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("sets the route labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, routeGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("route"))
				Expect(routeGUID).To(BeEquivalentTo("route-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2", "set-route-warning"))
			})
		})

		When("there are client errors", func() {
			When("fetching the route fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						nil,
						ccv3.Warnings{"get-domains-warning"},
						errors.New("get-domain-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-domains-warning"))
					Expect(executeErr).To(MatchError("get-domain-error"))
				})
			})

			When("updating the route fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{
							{Name: "domain-name", GUID: "domain-guid"},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)

					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{
							{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", URL: "hostname.domain-name", Path: "/the-path"},
						},
						ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
						nil,
					)

					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-route-warning"},
						errors.New("update-route-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2", "set-route-warning"))
				})
			})
		})
	})

	Describe("UpdateRoutePolicyLabels", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateRoutePolicyLabels("sub.example.com/my-path", "space-guid", "", labels)
		})

		When("there are no client errors and exactly one policy", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{{GUID: "policy-guid", Source: "cf:app:app1-guid"}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-policies-warning"},
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"", ccv3.Warnings{"set-label-warning"}, nil,
				)
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("queries GetRoutePolicies with route GUID filter", func() {
				Expect(fakeCloudControllerClient.GetRoutePoliciesCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetRoutePoliciesArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{"route-guid"}},
				))
			})

			It("calls UpdateResourceMetadata with the policy GUID", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, policyGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(resourceType).To(Equal("route-policy"))
				Expect(policyGUID).To(Equal("policy-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning", "get-policies-warning", "set-label-warning"))
			})
		})

		When("a source is given and it matches a policy", func() {
			JustBeforeEach(func() {
				warnings, executeErr = actor.UpdateRoutePolicyLabels("sub.example.com/my-path", "space-guid", "cf:app:app2-guid", labels)
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-1", Source: "cf:app:app1-guid"},
						{GUID: "policy-2", Source: "cf:app:app2-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns("", ccv3.Warnings{}, nil)
			})

			It("selects the policy matching the given source", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				_, policyGUID, _ := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(policyGUID).To(Equal("policy-2"))
			})
		})

		When("no source is given and there are multiple policies", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-1", Source: "cf:app:app1-guid"},
						{GUID: "policy-2", Source: "cf:app:app2-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns a RoutePolicyAmbiguityError", func() {
				Expect(executeErr).To(MatchError(actionerror.RoutePolicyAmbiguityError{
					RouteURL: "sub.example.com/my-path",
					Count:    2,
				}))
			})
		})

		When("a source is given but no policy matches", func() {
			JustBeforeEach(func() {
				warnings, executeErr = actor.UpdateRoutePolicyLabels("sub.example.com/my-path", "space-guid", "cf:app:unknown-guid", labels)
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-1", Source: "cf:app:app1-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns a RoutePolicyNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.RoutePolicyNotFoundError{Source: "cf:app:unknown-guid"}))
			})
		})

		When("fetching the route fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"get-domains-warning"},
					errors.New("get-route-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("get-route-error"))
				Expect(warnings).To(ConsistOf("get-domains-warning"))
			})
		})

		When("fetching route policies fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
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

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("get-policies-error"))
				Expect(warnings).To(ContainElement("get-policies-warning"))
			})
		})

		When("updating the metadata fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{{GUID: "policy-guid", Source: "cf:app:app1-guid"}},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"", ccv3.Warnings{"metadata-warning"}, errors.New("update-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("update-error"))
				Expect(warnings).To(ContainElement("metadata-warning"))
			})
		})
	})

	Describe("GetRoutePolicyLabels", func() {
		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetRoutePolicyLabels("sub.example.com/my-path", spaceGUID, "")
		})

		When("there are no client errors and exactly one policy", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{{
						GUID:   "policy-guid",
						Source: "cf:app:app1-guid",
						Metadata: &resources.Metadata{
							Labels: map[string]types.NullString{
								"key1": types.NewNullString("value1"),
							},
						},
					}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-policies-warning"},
					nil,
				)
			})

			It("returns the labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(labels).To(Equal(map[string]types.NullString{"key1": types.NewNullString("value1")}))
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning", "get-policies-warning"))
			})
		})

		When("the policy has no labels", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{{GUID: "policy-guid", Source: "cf:app:app1-guid"}},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns an empty map", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(labels).To(BeEmpty())
			})
		})

		When("there are multiple policies and no source given", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-1", Source: "cf:app:app1-guid"},
						{GUID: "policy-2", Source: "cf:app:app2-guid"},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns a RoutePolicyAmbiguityError", func() {
				Expect(executeErr).To(MatchError(actionerror.RoutePolicyAmbiguityError{
					RouteURL: "sub.example.com/my-path",
					Count:    2,
				}))
			})
		})

		When("a source is given and it matches", func() {
			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetRoutePolicyLabels("sub.example.com/my-path", spaceGUID, "cf:app:app2-guid")
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{Name: "sub.example.com", GUID: "domain-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "route-guid"}},
					ccv3.Warnings{},
					nil,
				)
				fakeCloudControllerClient.GetRoutePoliciesReturns(
					[]resources.RoutePolicy{
						{GUID: "policy-1", Source: "cf:app:app1-guid"},
						{GUID: "policy-2", Source: "cf:app:app2-guid", Metadata: &resources.Metadata{
							Labels: map[string]types.NullString{"env": types.NewNullString("prod")},
						}},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns the labels for the matching policy", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(labels).To(Equal(map[string]types.NullString{"env": types.NewNullString("prod")}))
			})
		})

		When("there is a client error fetching the route", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"get-domains-warning"},
					errors.New("get-route-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("get-route-error"))
				Expect(warnings).To(ConsistOf("get-domains-warning"))
			})
		})
	})

	Describe("UpdateSpaceLabelsBySpaceName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateSpaceLabelsBySpaceName(resourceName, orgGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{resources.Space{GUID: "some-guid"}},
					ccv3.IncludedResources{},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-space-metadata"},
					nil,
				)
			})

			It("gets the space", func() {
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("sets the space labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, spaceGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("space"))
				Expect(spaceGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-space-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the space fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{resources.Space{GUID: "some-guid"}},
						ccv3.IncludedResources{},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-spaces-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-spaces-error"))
				})
			})

			When("updating the space fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{resources.Space{GUID: "some-guid"}},
						ccv3.IncludedResources{},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-space"},
						errors.New("update-space-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-space"))
					Expect(executeErr).To(MatchError("update-space-error"))
				})
			})
		})
	})

	Describe("UpdateStackLabelsByStackName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateStackLabelsByStackName(resourceName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStacksReturns(
					[]resources.Stack{resources.Stack{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-stack-metadata"},
					nil,
				)
			})

			It("gets the stack", func() {
				Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetStacksArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			It("sets the stack labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, stackGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("stack"))
				Expect(stackGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-stack-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the stack fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{resources.Stack{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-stacks-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-stacks-error"))
				})
			})

			When("updating the stack fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{resources.Stack{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-stack"},
						errors.New("update-stack-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-stack"))
					Expect(executeErr).To(MatchError("update-stack-error"))
				})
			})
		})
	})

	Describe("UpdateServiceBrokerLabelsByServiceBrokerName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateServiceBrokerLabelsByServiceBrokerName(resourceName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{{GUID: "some-broker-guid", Name: resourceName}},
					[]string{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					ccv3.JobURL("fake-job-url"),
					ccv3.Warnings{"set-service-broker-metadata"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"poll-job-warning"}, nil)
			})

			It("gets the service broker", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
			})

			It("sets the service-broker labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, serviceBrokerGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(resourceType).To(BeEquivalentTo("service-broker"))
				Expect(serviceBrokerGUID).To(BeEquivalentTo("some-broker-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("polls the job", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(BeEquivalentTo("fake-job-url"))
			})

			It("aggregates warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-broker-metadata", "poll-job-warning"))
			})
		})

		When("there are client errors", func() {
			When("fetching the service-broker fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns(
						[]resources.ServiceBroker{},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-service-broker-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-service-broker-error"))
				})
			})

			When("updating the service-broker fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns(
						[]resources.ServiceBroker{{GUID: "some-guid", Name: resourceName}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						ccv3.JobURL(""),
						ccv3.Warnings{"set-service-broker"},
						errors.New("update-service-broker-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-broker"))
					Expect(executeErr).To(MatchError("update-service-broker-error"))
				})
			})

			When("polling the job fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns(
						[]resources.ServiceBroker{{GUID: "some-guid", Name: resourceName}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						ccv3.JobURL("fake-job-url"),
						ccv3.Warnings{"set-service-broker-metadata"},
						nil,
					)

					fakeCloudControllerClient.PollJobReturns(
						ccv3.Warnings{"another-poll-job-warning"},
						errors.New("polling-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-broker-metadata", "another-poll-job-warning"))
					Expect(executeErr).To(MatchError("polling-error"))
				})
			})
		})
	})

	Describe("UpdateServiceInstanceLabels", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateServiceInstanceLabels(resourceName, spaceGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						GUID: "fake-si-guid",
						Name: resourceName,
					},
					ccv3.IncludedResources{},
					[]string{"warning-1", "warning-2"},
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-si-labels-warnings"},
					nil,
				)
			})

			It("gets the service instance", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(resourceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})

			It("sets the service instance labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, siGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("service-instance"))
				Expect(siGUID).To(BeEquivalentTo("fake-si-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-si-labels-warnings"))
			})
		})

		When("there are client errors", func() {
			When("GetServiceInstanceByNameAndSpace fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						[]string{"warning-failure-1", "warning-failure-2"},
						errors.New("get-si-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-si-error"))
				})
			})

			When("UpdateResourceMetadata fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{
							GUID: "fake-si-guid",
							Name: resourceName,
						},
						ccv3.IncludedResources{},
						[]string{"warning-1", "warning-2"},
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-app-labels-warnings"},
						errors.New("update-si-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
					Expect(executeErr).To(MatchError("update-si-error"))
				})
			})
		})
	})

	Describe("UpdateServiceOfferingLabels", func() {
		const serviceBrokerName = "fake-service-broker"

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateServiceOfferingLabels(resourceName, serviceBrokerName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					resources.ServiceOffering{GUID: "some-service-offering-guid", Name: resourceName},
					[]string{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-service-offering-metadata"},
					nil,
				)
			})

			It("gets the service offering", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerCallCount()).To(Equal(1))
				requestedServiceName, requestedBrokerName := fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerArgsForCall(0)
				Expect(requestedServiceName).To(Equal(resourceName))
				Expect(requestedBrokerName).To(Equal(serviceBrokerName))
			})

			It("sets the service offering labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, serviceBrokerGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(resourceType).To(BeEquivalentTo("service-offering"))
				Expect(serviceBrokerGUID).To(BeEquivalentTo("some-service-offering-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-offering-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the service offering fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
						resources.ServiceOffering{},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-service-offerings-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-service-offerings-error"))
				})
			})

			When("updating the service offering fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
						resources.ServiceOffering{GUID: "some-guid"},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-service-offering"},
						errors.New("update-service-offering-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-offering"))
					Expect(executeErr).To(MatchError("update-service-offering-error"))
				})
			})
		})
	})

	Describe("UpdateServicePlanLabels", func() {
		const serviceBrokerName = "fake-service-broker"
		const serviceOfferingName = "fake-service-offering"

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateServicePlanLabels(resourceName, serviceOfferingName, serviceBrokerName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]resources.ServicePlan{{GUID: "some-service-plan-guid", Name: resourceName}},
					[]string{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					"",
					ccv3.Warnings{"set-service-plan-metadata"},
					nil,
				)
			})

			It("gets the service plan", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}},
					ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{serviceOfferingName}},
				))
			})

			It("sets the service plan labels", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, servicePlanGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(resourceType).To(BeEquivalentTo("service-plan"))
				Expect(servicePlanGUID).To(BeEquivalentTo("some-service-plan-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-plan-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the service plan fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-service-plan-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-service-plan-error"))
				})
			})

			When("updating the service plan fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{{GUID: "some-guid"}},
						[]string{"warning-1", "warning-2"},
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						"",
						ccv3.Warnings{"set-service-plan"},
						errors.New("update-service-plan-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-service-plan"))
					Expect(executeErr).To(MatchError("update-service-plan-error"))
				})
			})
		})
	})

	Describe("GetDomainLabels", func() {
		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetDomainLabels(resourceName)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
			})

			It("gets the domain", func() {
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			When("there are no labels on a domain", func() {
				It("returns an empty map", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(labels).To(BeEmpty())
				})
			})

			When("there are labels", func() {
				var expectedLabels map[string]types.NullString

				BeforeEach(func() {
					expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{{
							GUID: "some-guid",
							Metadata: &resources.Metadata{
								Labels: expectedLabels,
							},
						}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
				})
				It("returns the labels", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(labels).To(Equal(expectedLabels))
				})
			})
		})

		When("there is a client error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					errors.New("get-domains-error"),
				)
			})
			When("GetDomainByName fails", func() {
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).To(MatchError("get-domains-error"))
				})
			})
		})
	})

	Describe("GetRouteLabels", func() {
		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetRouteLabels("sub.example.com/my-route/path", spaceGUID)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{GUID: "route-guid", SpaceGUID: "some-space-guid", DomainGUID: "domain-guid", Host: "hostname", URL: "hostname.domain-name", Path: "/the-path"},
					},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
			})

			It("gets the route", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{"domain-guid"}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{""}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{"/my-route/path"}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			When("there are no labels on a route", func() {
				It("returns an empty map", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-domains-warning", "warning-1", "warning-2"))
					Expect(labels).To(BeEmpty())
				})
			})

			When("there are labels", func() {
				var expectedLabels map[string]types.NullString

				BeforeEach(func() {
					expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{
							{
								GUID: "some-guid",
								Metadata: &resources.Metadata{
									Labels: expectedLabels,
								},
							},
						},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
				})
				It("returns the labels", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-domains-warning", "warning-1", "warning-2"))
					Expect(labels).To(Equal(expectedLabels))
				})
			})
		})

		When("there is a client error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					errors.New("get-routes-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-domains-warning", "warning-1", "warning-2"))
				Expect(executeErr).To(MatchError("get-routes-error"))
			})
		})
	})

	Describe("GetStackLabels", func() {
		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetStackLabels(resourceName)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStacksReturns(
					[]resources.Stack{resources.Stack{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
			})

			It("gets the stack", func() {
				Expect(fakeCloudControllerClient.GetStacksCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetStacksArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
					ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
				))
			})

			When("there are no labels on a stack", func() {
				It("returns an empty map", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(labels).To(BeEmpty())
				})
			})

			When("there are labels", func() {
				var expectedLabels map[string]types.NullString

				BeforeEach(func() {
					expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
					fakeCloudControllerClient.GetStacksReturns(
						[]resources.Stack{resources.Stack{
							GUID: "some-guid",
							Metadata: &resources.Metadata{
								Labels: expectedLabels,
							},
						}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
				})
				It("returns the labels", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(labels).To(Equal(expectedLabels))
				})
			})
		})

		When("there is a client error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetStacksReturns(
					[]resources.Stack{resources.Stack{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					errors.New("get-stacks-error"),
				)
			})
			When("GetStackByName fails", func() {
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).To(MatchError("get-stacks-error"))
				})
			})
		})
	})

	Describe("GetServiceBrokerLabels", func() {
		When("service broker does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceBrokerLabels(resourceName)
			})

			It("returns a service broker not found error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr.Error()).To(ContainSubstring("Service broker 'some-resource' not found"))
			})
		})

		When("client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{},
					[]string{"warning"},
					errors.New("some random error"),
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceBrokerLabels(resourceName)
			})

			It("returns error and prints warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(executeErr).To(MatchError("some random error"))
			})
		})
		When("service broker has labels", func() {
			var expectedLabels map[string]types.NullString

			BeforeEach(func() {
				expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{resources.ServiceBroker{
						GUID: "some-guid",
						Name: resourceName,
						Metadata: &resources.Metadata{
							Labels: expectedLabels,
						},
					}},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceBrokerLabels(resourceName)
			})

			It("returns labels associated with the service broker as well as warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(labels).To(Equal(expectedLabels))
			})
		})

	})

	Describe("GetServiceInstanceLabels", func() {
		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetServiceInstanceLabels(resourceName, spaceGUID)
		})

		When("does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					[]string{"warning-1", "warning-2"},
					ccerror.ServiceInstanceNotFoundError{},
				)
			})

			It("returns a service instance not found error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					[]string{"warning"},
					errors.New("some random error"),
				)
			})

			It("returns error and warnings", func() {
				Expect(executeErr).To(MatchError("some random error"))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})

		When("service instance has labels", func() {
			var expectedLabels map[string]types.NullString

			BeforeEach(func() {
				expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						GUID: "some-guid",
						Name: resourceName,
						Metadata: &resources.Metadata{
							Labels: expectedLabels,
						},
					},
					ccv3.IncludedResources{},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns labels associated with the service broker as well as warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(labels).To(Equal(expectedLabels))
			})
		})
	})

	Describe("GetServiceOfferingLabels", func() {
		const serviceBrokerName = "my-service-broker"

		When("service offering does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					resources.ServiceOffering{},
					[]string{"warning-1", "warning-2"},
					ccerror.ServiceOfferingNotFoundError{
						ServiceOfferingName: resourceName,
						ServiceBrokerName:   serviceBrokerName,
					},
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceOfferingLabels(resourceName, serviceBrokerName)
			})

			It("returns a service offering not found error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr.Error()).To(ContainSubstring("Service offering '%s' for service broker '%s' not found", resourceName, serviceBrokerName))
			})
		})

		When("client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					resources.ServiceOffering{},
					[]string{"warning"},
					errors.New("some random error"),
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceOfferingLabels(resourceName, serviceBrokerName)
			})

			It("returns error and prints warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(executeErr).To(MatchError("some random error"))
			})
		})

		When("service offering has labels", func() {
			var expectedLabels map[string]types.NullString

			BeforeEach(func() {
				expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					resources.ServiceOffering{
						GUID: "some-guid",
						Name: resourceName,
						Metadata: &resources.Metadata{
							Labels: expectedLabels,
						},
					},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServiceOfferingLabels(resourceName, serviceBrokerName)
			})

			It("queries the right names", func() {
				Expect(fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerCallCount()).To(Equal(1))
				requestedServiceOffering, requestedServiceBroker := fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerArgsForCall(0)
				Expect(requestedServiceOffering).To(Equal(resourceName))
				Expect(requestedServiceBroker).To(Equal(serviceBrokerName))
			})

			It("returns labels associated with the service offering as well as warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(labels).To(Equal(expectedLabels))
			})
		})
	})

	Describe("GetServicePlanLabels", func() {
		When("service plan does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]resources.ServicePlan{},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServicePlanLabels(resourceName, "", "")
			})

			It("returns a service plan not found error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr.Error()).To(ContainSubstring("Service plan '%s' not found", resourceName))
			})
		})

		When("client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]resources.ServicePlan{},
					[]string{"warning"},
					errors.New("some random error"),
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServicePlanLabels(resourceName, "", "")
			})

			It("returns error and prints warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(executeErr).To(MatchError("some random error"))
			})
		})

		When("service plan has labels", func() {
			var expectedLabels map[string]types.NullString

			BeforeEach(func() {
				expectedLabels = map[string]types.NullString{"key1": types.NewNullString("value1"), "key2": types.NewNullString("value2")}
				fakeCloudControllerClient.GetServicePlansReturns(
					[]resources.ServicePlan{{
						GUID: "some-guid",
						Name: resourceName,
						Metadata: &resources.Metadata{
							Labels: expectedLabels,
						},
					}},
					[]string{"warning-1", "warning-2"},
					nil,
				)
			})

			JustBeforeEach(func() {
				labels, warnings, executeErr = actor.GetServicePlanLabels(resourceName, "serviceOfferingName", "serviceBrokerName")
			})

			It("queries the right names", func() {
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{resourceName}},
					ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{"serviceOfferingName"}},
					ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{"serviceBrokerName"}},
				))
			})

			It("returns labels associated with the service plan as well as warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(labels).To(Equal(expectedLabels))
			})
		})
	})
})
