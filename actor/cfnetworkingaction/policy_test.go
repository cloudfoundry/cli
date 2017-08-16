package cfnetworkingaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction/cfnetworkingactionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cfnetworking/cfnetv1"
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

	JustBeforeEach(func() {
		spaceGuid := "space"
		srcApp := "appA"
		destApp := "appB"
		protocol := "tcp"
		startPort := 8080
		endPort := 8090
		warnings, executeErr = actor.AllowNetworkAccess(spaceGuid, srcApp, destApp, protocol, startPort, endPort)
	})

	Describe("AllowNetworkAccess", func() {
		It("creates policies", func() {
			Expect(warnings).To(Equal(Warnings([]string{"v3ActorWarningA", "v3ActorWarningB"})))
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeV3Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
			sourceAppName, spaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(sourceAppName).To(Equal("appA"))
			Expect(spaceGUID).To(Equal("space"))

			destAppName, spaceGUID := fakeV3Actor.GetApplicationByNameAndSpaceArgsForCall(1)
			Expect(destAppName).To(Equal("appB"))
			Expect(spaceGUID).To(Equal("space"))

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

		Context("when getting the source app fails ", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, []string{"v3ActorWarningA"}, errors.New("banana"))
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(Equal(Warnings([]string{"v3ActorWarningA"})))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		Context("when getting the destination app fails ", func() {
			BeforeEach(func() {
				fakeV3Actor.GetApplicationByNameAndSpaceStub = func(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error) {
					if appName == "appB" {
						return v3action.Application{}, []string{"v3ActorWarningB"}, errors.New("banana")
					}
					return v3action.Application{}, []string{"v3ActorWarningA"}, nil
				}
			})
			It("returns a sensible error", func() {
				Expect(warnings).To(Equal(Warnings([]string{"v3ActorWarningA", "v3ActorWarningB"})))
				Expect(executeErr).To(MatchError("banana"))
			})
		})

		Context("when creating the policy fails", func() {
			BeforeEach(func() {
				fakeNetworkingClient.CreatePoliciesReturns(errors.New("apple"))
			})
			It("returns a sensible error", func() {
				Expect(executeErr).To(MatchError("apple"))
			})
		})
	})
})
