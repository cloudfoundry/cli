package cfnetworkingaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction/cfnetworkingactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *cfnetworkingactionfakes.FakeCloudControllerClient
		fakeNetworkingClient      *cfnetworkingactionfakes.FakeNetworkingClient

		warnings   Warnings
		executeErr error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(cfnetworkingactionfakes.FakeCloudControllerClient)
		fakeNetworkingClient = new(cfnetworkingactionfakes.FakeNetworkingClient)

		fakeCloudControllerClient.GetApplicationByNameAndSpaceStub = func(appName string, spaceGUID string) (resources.Application, ccv3.Warnings, error) {
			if appName == "appA" {
				return resources.Application{GUID: "appAGUID"}, []string{"v3ActorWarningA"}, nil
			} else if appName == "appB" {
				return resources.Application{GUID: "appBGUID"}, []string{"v3ActorWarningB"}, nil
			}
			return resources.Application{}, nil, nil
		}

		actor = NewActor(fakeNetworkingClient, fakeCloudControllerClient)
	})

	Describe("AddNetworkPolicy", func() {
		JustBeforeEach(func() {
			srcSpaceGuid := "src-space"
			srcApp := "appA"
			destSpaceGuid := "dst-space"
			destApp := "appB"
			protocol := "tcp"
			startPort := 8080
			endPort := 8090
			warnings, executeErr = actor.AddNetworkPolicy(srcSpaceGuid, srcApp, destSpaceGuid, destApp, protocol, startPort, endPort)
		})

		It("creates policies", func() {
			Expect(warnings).To(ConsistOf("v3ActorWarningA", "v3ActorWarningB"))
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeCloudControllerClient.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
			sourceAppName, srcSpaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(sourceAppName).To(Equal("appA"))
			Expect(srcSpaceGUID).To(Equal("src-space"))

			destAppName, destSpaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(1)
			Expect(destAppName).To(Equal("appB"))
			Expect(destSpaceGUID).To(Equal("dst-space"))

			Expect(fakeNetworkingClient.CreatePoliciesCallCount()).To(Equal(1))
			Expect(fakeNetworkingClient.CreatePoliciesArgsForCall(0)).To(Equal([]cfnetv1.Policy{
				{
					Source: cfnetv1.PolicySource{
						ID: "appAGUID",
					},
					Destination: cfnetv1.PolicyDestination{
						ID:       "appBGUID",
						Protocol: "tcp",
						Ports: cfnetv1.Ports{
							Start: 8080,
							End:   8090,
						},
					},
				},
			}))
		})

		When("getting the source app fails ", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(resources.Application{}, []string{"v3ActorWarningA"}, errors.New("banana"))
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the destination app fails ", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationByNameAndSpaceStub = func(appName string, spaceGUID string) (resources.Application, ccv3.Warnings, error) {
					if appName == "appB" {
						return resources.Application{}, []string{"v3ActorWarningB"}, errors.New("banana")
					}
					return resources.Application{}, []string{"v3ActorWarningA"}, nil
				}
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA", "v3ActorWarningB"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("creating the policy fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.CreatePoliciesReturns(errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})
	})

	Describe("NetworkPoliciesBySpaceAndAppName", func() {
		var (
			policies []Policy
			srcApp   string
		)

		BeforeEach(func() {
			fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{{
				Source: cfnetv1.PolicySource{
					ID: "appAGUID",
				},
				Destination: cfnetv1.PolicyDestination{
					ID:       "appBGUID",
					Protocol: "tcp",
					Ports: cfnetv1.Ports{
						Start: 8080,
						End:   8080,
					},
				},
			}, {
				Source: cfnetv1.PolicySource{
					ID: "appAGUID",
				},
				Destination: cfnetv1.PolicyDestination{
					ID:       "appCGUID",
					Protocol: "tcp",
					Ports: cfnetv1.Ports{
						Start: 8080,
						End:   8080,
					},
				},
			}}, nil)

			fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(resources.Application{
				Name: "appA",
				GUID: "appAGUID",
				Relationships: map[constant.RelationshipType]resources.Relationship{
					constant.RelationshipTypeSpace: {GUID: "spaceAGUID"},
				},
			}, []string{"GetApplicationByNameAndSpaceWarning"}, nil)

			fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{
				{
					Name: "appB",
					GUID: "appBGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceAGUID"},
					},
				},
				{
					Name: "appC",
					GUID: "appCGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceCGUID"},
					},
				},
			}, []string{"GetApplicationsWarning"}, nil)

			fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{
				{
					Name: "spaceA",
					GUID: "spaceAGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeOrganization: {GUID: "orgAGUID"},
					},
				},
				{
					Name: "spaceC",
					GUID: "spaceCGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeOrganization: {GUID: "orgCGUID"},
					},
				},
			}, ccv3.IncludedResources{}, []string{"GetSpacesWarning"}, nil)

			fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{
				{
					Name: "orgA",
					GUID: "orgAGUID",
				},
				{
					Name: "orgC",
					GUID: "orgCGUID",
				},
			}, []string{"GetOrganizationsWarning"}, nil)
		})

		JustBeforeEach(func() {
			srcSpaceGuid := "space"
			policies, warnings, executeErr = actor.NetworkPoliciesBySpaceAndAppName(srcSpaceGuid, srcApp)
		})

		When("listing policies based on a source app", func() {
			BeforeEach(func() {
				srcApp = "appA"
			})

			It("lists only policies for which the app is a source", func() {
				Expect(policies).To(Equal([]Policy{
					{
						SourceName:           "appA",
						DestinationName:      "appB",
						Protocol:             "tcp",
						StartPort:            8080,
						EndPort:              8080,
						DestinationSpaceName: "spaceA",
						DestinationOrgName:   "orgA",
					},
					{
						SourceName:           "appA",
						DestinationName:      "appC",
						Protocol:             "tcp",
						StartPort:            8080,
						EndPort:              8080,
						DestinationSpaceName: "spaceC",
						DestinationOrgName:   "orgC",
					},
				},
				))
			})

			It("passes through the source app argument", func() {
				Expect(warnings).To(ConsistOf(
					"GetApplicationByNameAndSpaceWarning",
					"GetApplicationsWarning",
					"GetSpacesWarning",
					"GetOrganizationsWarning",
				))
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				sourceAppName, spaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(sourceAppName).To(Equal("appA"))
				Expect(spaceGUID).To(Equal("space"))

				Expect(fakeNetworkingClient.ListPoliciesCallCount()).To(Equal(1))
				Expect(fakeNetworkingClient.ListPoliciesArgsForCall(0)).To(Equal([]string{"appAGUID"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(Equal([]ccv3.Query{
					{Key: ccv3.GUIDFilter, Values: []string{"appBGUID", "appCGUID"}},
				}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv3.Query{
					{Key: ccv3.GUIDFilter, Values: []string{"spaceAGUID", "spaceCGUID"}},
				}))
			})
		})

		When("getting the applications by name and space fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(resources.Application{}, []string{"GetApplicationsBySpaceWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetApplicationsBySpaceWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("listing the policy fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{}, errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})

		When("getting the applications by guids fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{}, []string{"GetApplicationsWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetApplicationsWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the spaces by guids fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{}, ccv3.IncludedResources{}, []string{"GetSpacesWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetSpacesWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})
	})

	Describe("NetworkPoliciesBySpace", func() {
		var (
			policies []Policy
		)

		BeforeEach(func() {
			fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{{
				Source: cfnetv1.PolicySource{
					ID: "appAGUID",
				},
				Destination: cfnetv1.PolicyDestination{
					ID:       "appBGUID",
					Protocol: "tcp",
					Ports: cfnetv1.Ports{
						Start: 8080,
						End:   8080,
					},
				},
			}, {
				Source: cfnetv1.PolicySource{
					ID: "appBGUID",
				},
				Destination: cfnetv1.PolicyDestination{
					ID:       "appBGUID",
					Protocol: "tcp",
					Ports: cfnetv1.Ports{
						Start: 8080,
						End:   8080,
					},
				},
			}, {
				Source: cfnetv1.PolicySource{
					ID: "appAGUID",
				},
				Destination: cfnetv1.PolicyDestination{
					ID:       "appCGUID",
					Protocol: "tcp",
					Ports: cfnetv1.Ports{
						Start: 8080,
						End:   8080,
					},
				},
			}}, nil)

			fakeCloudControllerClient.GetApplicationsReturnsOnCall(0, []resources.Application{
				{
					Name: "appA",
					GUID: "appAGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceAGUID"},
					},
				},
				{
					Name: "appB",
					GUID: "appBGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceAGUID"},
					},
				},
				{
					Name: "appC",
					GUID: "appCGUID",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceCGUID"},
					},
				},
			}, []string{"filter-apps-by-space-warning"}, nil)

			fakeCloudControllerClient.GetApplicationsReturnsOnCall(1, []resources.Application{
				{
					GUID: "appBGUID",
					Name: "appB",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceAGUID"},
					},
				},
				{
					GUID: "appCGUID",
					Name: "appC",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeSpace: {GUID: "spaceCGUID"},
					},
				},
			}, []string{"filter-apps-by-guid-warning"}, nil)

			fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{
				{
					GUID: "spaceAGUID",
					Name: "spaceA",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeOrganization: {GUID: "orgAGUID"},
					},
				},
				{
					GUID: "spaceCGUID",
					Name: "spaceC",
					Relationships: map[constant.RelationshipType]resources.Relationship{
						constant.RelationshipTypeOrganization: {GUID: "orgCGUID"},
					},
				},
			}, ccv3.IncludedResources{}, []string{"GetSpaceWarning"}, nil)

			fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{
				{
					GUID: "orgAGUID",
					Name: "orgA",
				},
				{
					GUID: "orgCGUID",
					Name: "orgC",
				},
			}, []string{"GetOrganizationsWarning"}, nil)
		})

		JustBeforeEach(func() {
			spaceGuid := "space"
			policies, warnings, executeErr = actor.NetworkPoliciesBySpace(spaceGuid)
		})

		It("lists policies", func() {
			Expect(policies).To(Equal(
				[]Policy{{
					SourceName:           "appA",
					DestinationName:      "appB",
					Protocol:             "tcp",
					StartPort:            8080,
					EndPort:              8080,
					DestinationSpaceName: "spaceA",
					DestinationOrgName:   "orgA",
				}, {
					SourceName:           "appB",
					DestinationName:      "appB",
					Protocol:             "tcp",
					StartPort:            8080,
					EndPort:              8080,
					DestinationSpaceName: "spaceA",
					DestinationOrgName:   "orgA",
				}, {
					SourceName:           "appA",
					DestinationName:      "appC",
					Protocol:             "tcp",
					StartPort:            8080,
					EndPort:              8080,
					DestinationSpaceName: "spaceC",
					DestinationOrgName:   "orgC",
				}},
			))
			Expect(warnings).To(ConsistOf(
				"filter-apps-by-space-warning",
				"filter-apps-by-guid-warning",
				"GetSpaceWarning",
				"GetOrganizationsWarning",
			))
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeNetworkingClient.ListPoliciesCallCount()).To(Equal(1))
			Expect(fakeNetworkingClient.ListPoliciesArgsForCall(0)).To(ConsistOf("appAGUID", "appBGUID", "appCGUID"))

			Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(2))
			Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(Equal([]ccv3.Query{
				{Key: ccv3.SpaceGUIDFilter, Values: []string{"space"}},
			}))
			Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(1)).To(Equal([]ccv3.Query{
				{Key: ccv3.GUIDFilter, Values: []string{"appBGUID", "appCGUID"}},
			}))

			Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv3.Query{
				{Key: ccv3.GUIDFilter, Values: []string{"spaceAGUID", "spaceCGUID"}},
			}))

			Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv3.Query{
				{Key: ccv3.GUIDFilter, Values: []string{"orgAGUID", "orgCGUID"}},
			}))
		})

		// policy server returns policies that match the give app guid in the source or destination
		// we only care about the policies that match the source guid.
		When("the policy server returns policies that have matching destination app guids", func() {
			BeforeEach(func() {
				fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{{
					Source: cfnetv1.PolicySource{
						ID: "appDGUID",
					},
					Destination: cfnetv1.PolicyDestination{
						ID:       "appAGUID",
						Protocol: "tcp",
						Ports: cfnetv1.Ports{
							Start: 8080,
							End:   8080,
						},
					},
				}}, nil)
			})

			It("filters them out ", func() {
				Expect(policies).To(BeEmpty())
			})
		})

		When("getting the applications with a space guids filter fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturnsOnCall(0, []resources.Application{}, []string{"filter-apps-by-space-warning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ConsistOf("filter-apps-by-space-warning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("listing the policy fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{}, errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})

		When("getting the applications by guids fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturnsOnCall(1, []resources.Application{}, []string{"filter-apps-by-guid-warning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("filter-apps-by-guid-warning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the spaces by guids fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{}, ccv3.IncludedResources{}, []string{"GetSpacesWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetSpacesWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})
	})

	Describe("RemoveNetworkPolicy", func() {
		BeforeEach(func() {
			fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{
				{
					Source: cfnetv1.PolicySource{
						ID: "appAGUID",
					},
					Destination: cfnetv1.PolicyDestination{
						ID:       "appBGUID",
						Protocol: "udp",
						Ports: cfnetv1.Ports{
							Start: 123,
							End:   345,
						},
					},
				},
			}, nil)
		})

		JustBeforeEach(func() {
			srcSpaceGuid := "spaceA"
			srcApp := "appA"
			destSpaceGuid := "spaceB"
			destApp := "appB"
			protocol := "udp"
			startPort := 123
			endPort := 345
			warnings, executeErr = actor.RemoveNetworkPolicy(srcSpaceGuid, srcApp, destSpaceGuid, destApp, protocol, startPort, endPort)
		})
		It("removes policies", func() {
			Expect(warnings).To(ConsistOf("v3ActorWarningA", "v3ActorWarningB"))
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeCloudControllerClient.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
			sourceAppName, spaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(sourceAppName).To(Equal("appA"))
			Expect(spaceGUID).To(Equal("spaceA"))

			destAppName, spaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(1)
			Expect(destAppName).To(Equal("appB"))
			Expect(spaceGUID).To(Equal("spaceB"))

			Expect(fakeNetworkingClient.ListPoliciesCallCount()).To(Equal(1))

			Expect(fakeNetworkingClient.RemovePoliciesCallCount()).To(Equal(1))
			Expect(fakeNetworkingClient.RemovePoliciesArgsForCall(0)).To(Equal([]cfnetv1.Policy{
				{
					Source: cfnetv1.PolicySource{
						ID: "appAGUID",
					},
					Destination: cfnetv1.PolicyDestination{
						ID:       "appBGUID",
						Protocol: "udp",
						Ports: cfnetv1.Ports{
							Start: 123,
							End:   345,
						},
					},
				},
			}))
		})

		When("the policy does not exist", func() {
			BeforeEach(func() {
				fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{}, nil)
			})

			It("returns an error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA", "v3ActorWarningB"))
				Expect(executeErr).To(MatchError(actionerror.PolicyDoesNotExistError{}))
			})
		})

		When("getting the source app fails ", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(resources.Application{}, []string{"v3ActorWarningA"}, errors.New("banana"))
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the destination app fails ", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationByNameAndSpaceReturnsOnCall(0, resources.Application{}, []string{"v3ActorWarningA"}, nil)
				fakeCloudControllerClient.GetApplicationByNameAndSpaceReturnsOnCall(1, resources.Application{}, []string{"v3ActorWarningB"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA", "v3ActorWarningB"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("listing policies fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.ListPoliciesReturns([]cfnetv1.Policy{}, errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})

		When("removing the policy fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.RemovePoliciesReturns(errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})
	})
})
