package cfnetworkingaction_test

import (
	"errors"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction/cfnetworkingactionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy", func() {
	var (
		actor                *Actor
		fakeV3Actor          *cfnetworkingactionfakes.FakeV3Actor
		fakeNetworkingClient *cfnetworkingactionfakes.FakeNetworkingClient

		warnings   Warnings
		executeErr error
	)

	BeforeEach(func() {
		fakeV3Actor = new(cfnetworkingactionfakes.FakeV3Actor)
		fakeNetworkingClient = new(cfnetworkingactionfakes.FakeNetworkingClient)

		fakeV3Actor.GetApplicationByNameAndSpaceStub = func(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error) {
			if appName == "appA" {
				return v3action.Application{GUID: "appAGUID"}, []string{"v3ActorWarningA"}, nil
			} else if appName == "appB" {
				return v3action.Application{GUID: "appBGUID"}, []string{"v3ActorWarningB"}, nil
			}
			return v3action.Application{}, nil, nil
		}

		actor = NewActor(fakeNetworkingClient, fakeV3Actor)
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

			Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
			sourceAppName, srcSpaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(sourceAppName).To(Equal("appA"))
			Expect(srcSpaceGUID).To(Equal("src-space"))

			destAppName, destSpaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(1)
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
				fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, []string{"v3ActorWarningA"}, errors.New("banana"))
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the destination app fails ", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationByNameAndSpaceStub = func(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error) {
					if appName == "appB" {
						return v3action.Application{}, []string{"v3ActorWarningB"}, errors.New("banana")
					}
					return v3action.Application{}, []string{"v3ActorWarningA"}, nil
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

			fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{
				Name:      "appA",
				GUID:      "appAGUID",
				SpaceGUID: "spaceAGUID",
			}, []string{"GetApplicationByNameAndSpaceWarning"}, nil)

			fakeV3Actor.GetApplicationsByGUIDsReturns([]v3action.Application{
				{
					Name:      "appB",
					GUID:      "appBGUID",
					SpaceGUID: "spaceAGUID",
				},
				{
					Name:      "appC",
					GUID:      "appCGUID",
					SpaceGUID: "spaceCGUID",
				},
			}, []string{"GetApplicationsByGUIDsWarning"}, nil)

			fakeV3Actor.GetSpacesByGUIDsReturns([]v3action.Space{
				{
					Name: "spaceA",
					GUID: "spaceAGUID",
				},
				{
					Name: "spaceC",
					GUID: "spaceCGUID",
				},
			}, []string{"GetSpacesByGUIDsWarning"}, nil)
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
					},
					{
						SourceName:           "appA",
						DestinationName:      "appC",
						Protocol:             "tcp",
						StartPort:            8080,
						EndPort:              8080,
						DestinationSpaceName: "spaceC",
					},
				},
				))
			})

			It("passes through the source app argument", func() {
				Expect(warnings).To(ConsistOf("GetApplicationByNameAndSpaceWarning", "GetApplicationsByGUIDsWarning", "GetSpacesByGUIDsWarning"))
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				sourceAppName, spaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(sourceAppName).To(Equal("appA"))
				Expect(spaceGUID).To(Equal("space"))

				Expect(fakeNetworkingClient.ListPoliciesCallCount()).To(Equal(1))
				Expect(fakeNetworkingClient.ListPoliciesArgsForCall(0)).To(Equal([]string{"appAGUID"}))

				Expect(fakeV3Actor.GetApplicationsByGUIDsCallCount()).To(Equal(1))
				Expect(fakeV3Actor.GetApplicationsByGUIDsArgsForCall(0)).To(ConsistOf("appBGUID", "appCGUID"))

				Expect(fakeV3Actor.GetSpacesByGUIDsCallCount()).To(Equal(1))
				Expect(fakeV3Actor.GetSpacesByGUIDsArgsForCall(0)).To(ConsistOf("spaceAGUID", "spaceCGUID"))
			})
		})

		When("getting the applications by name and space fails", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, []string{"GetApplicationsBySpaceWarning"}, errors.New("banana"))
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
				fakeV3Actor.GetApplicationsByGUIDsReturns([]v3action.Application{}, []string{"GetApplicationsByGUIDsWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetApplicationsByGUIDsWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the spaces by guids fails", func() {
			BeforeEach(func() {
				fakeV3Actor.GetSpacesByGUIDsReturns([]v3action.Space{}, []string{"GetSpacesByGUIDsWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetSpacesByGUIDsWarning"))
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

			fakeV3Actor.GetApplicationsBySpaceReturns([]v3action.Application{
				{
					Name:      "appA",
					GUID:      "appAGUID",
					SpaceGUID: "spaceAGUID",
				},
				{
					Name:      "appB",
					GUID:      "appBGUID",
					SpaceGUID: "spaceAGUID",
				},
				{
					Name:      "appC",
					GUID:      "appCGUID",
					SpaceGUID: "spaceCGUID",
				},
			}, []string{"GetApplicationsBySpaceWarning"}, nil)

			fakeV3Actor.GetApplicationsByGUIDsReturns([]v3action.Application{
				{
					GUID:      "appBGUID",
					Name:      "appB",
					SpaceGUID: "spaceAGUID",
				},
				{
					GUID:      "appCGUID",
					Name:      "appC",
					SpaceGUID: "spaceCGUID",
				},
			}, []string{"GetApplicationsByGUIDsWarning"}, nil)

			fakeV3Actor.GetSpacesByGUIDsReturns([]v3action.Space{
				{
					GUID: "spaceAGUID",
					Name: "spaceA",
				},
				{
					GUID: "spaceCGUID",
					Name: "spaceC",
				},
			}, []string{"GetSpaceByGUIDsWarning"}, nil)
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
				}, {
					SourceName:           "appB",
					DestinationName:      "appB",
					Protocol:             "tcp",
					StartPort:            8080,
					EndPort:              8080,
					DestinationSpaceName: "spaceA",
				}, {
					SourceName:           "appA",
					DestinationName:      "appC",
					Protocol:             "tcp",
					StartPort:            8080,
					EndPort:              8080,
					DestinationSpaceName: "spaceC",
				}},
			))
			Expect(warnings).To(ConsistOf("GetApplicationsBySpaceWarning", "GetApplicationsByGUIDsWarning", "GetSpaceByGUIDsWarning"))
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeV3Actor.GetApplicationsBySpaceCallCount()).To(Equal(1))
			Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))

			Expect(fakeNetworkingClient.ListPoliciesCallCount()).To(Equal(1))
			Expect(fakeNetworkingClient.ListPoliciesArgsForCall(0)).To(ConsistOf("appAGUID", "appBGUID", "appCGUID"))

			Expect(fakeV3Actor.GetApplicationsByGUIDsCallCount()).To(Equal(1))
			Expect(fakeV3Actor.GetApplicationsByGUIDsArgsForCall(0)).To(ConsistOf("appBGUID", "appCGUID"))

			Expect(fakeV3Actor.GetSpacesByGUIDsCallCount()).To(Equal(1))
			Expect(fakeV3Actor.GetSpacesByGUIDsArgsForCall(0)).To(ConsistOf("spaceAGUID", "spaceCGUID"))
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

		When("getting the applications fails", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationsBySpaceReturns([]v3action.Application{}, []string{"GetApplicationsBySpaceWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ConsistOf("GetApplicationsBySpaceWarning"))
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
				fakeV3Actor.GetApplicationsByGUIDsReturns([]v3action.Application{}, []string{"GetApplicationsByGUIDsWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetApplicationsByGUIDsWarning"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the spaces by guids fails", func() {
			BeforeEach(func() {
				fakeV3Actor.GetSpacesByGUIDsReturns([]v3action.Space{}, []string{"GetSpacesByGUIDsWarning"}, errors.New("banana"))
			})

			It("returns a sensible error", func() {
				Expect(policies).To(Equal([]Policy{}))
				Expect(warnings).To(ContainElement("GetSpacesByGUIDsWarning"))
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

			Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
			sourceAppName, spaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(sourceAppName).To(Equal("appA"))
			Expect(spaceGUID).To(Equal("spaceA"))

			destAppName, spaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(1)
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
				fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, []string{"v3ActorWarningA"}, errors.New("banana"))
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(ConsistOf("v3ActorWarningA"))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		When("getting the destination app fails ", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationByNameAndSpaceReturnsOnCall(0, v3action.Application{}, []string{"v3ActorWarningA"}, nil)
				fakeV3Actor.GetApplicationByNameAndSpaceReturnsOnCall(1, v3action.Application{}, []string{"v3ActorWarningB"}, errors.New("banana"))
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
