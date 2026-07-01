package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	. "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("route-policies Command", func() {
	var (
		cmd             RoutePoliciesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = RoutePoliciesCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.APIVersionReturns(ccversion.MinVersionRoutePolicies)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org", GUID: "org-guid"})
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "space-guid"})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version check fails", func() {
		BeforeEach(func() {
			fakeConfig.APIVersionReturns("0.0.0")
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionRoutePolicies,
			}))
		})
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: "faceman"})
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("getting the current user fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("user-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("user-error"))
		})
	})

	When("GetRoutePoliciesForSpace returns an error", func() {
		BeforeEach(func() {
			fakeActor.GetRoutePoliciesForSpaceReturns(nil, v7action.Warnings{"some-warning"}, errors.New("list-error"))
		})

		It("returns the error and displays warnings", func() {
			Expect(executeErr).To(MatchError("list-error"))
			Expect(testUI.Err).To(Say("some-warning"))
		})
	})

	When("there are no route policies", func() {
		BeforeEach(func() {
			fakeActor.GetRoutePoliciesForSpaceReturns([]v7action.RoutePolicyWithRoute{}, v7action.Warnings{}, nil)
		})

		It("displays 'No route policies found.'", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("No route policies found\\."))
			Expect(fakeActor.GetRoutePoliciesForSpaceCallCount()).To(Equal(1))
		})
	})

	When("there are route policies", func() {
		BeforeEach(func() {
			fakeActor.GetRoutePoliciesForSpaceReturns(
				[]v7action.RoutePolicyWithRoute{
					{
						RoutePolicy: resources.RoutePolicy{GUID: "p-guid", Source: "cf:any"},
						Route:       resources.Route{Host: "backend", Path: "/api"},
						DomainName:  "apps.example.com",
						ScopeType:   "any",
						SourceName:  "",
					},
				},
				v7action.Warnings{"some-warning"},
				nil,
			)
		})

		It("displays the table and passes the right args to the actor", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say("host"))
			Expect(testUI.Out).To(Say("backend"))
			Expect(testUI.Out).To(Say("apps.example.com"))
			Expect(testUI.Out).To(Say("cf:any"))

			Expect(fakeActor.GetRoutePoliciesForSpaceCallCount()).To(Equal(1))
			spaceGUIDArg, domainArg, hostnameArg, pathArg, labelsArg := fakeActor.GetRoutePoliciesForSpaceArgsForCall(0)
			Expect(spaceGUIDArg).To(Equal("space-guid"))
			Expect(domainArg).To(Equal(""))
			Expect(hostnameArg).To(Equal(""))
			Expect(pathArg).To(Equal(""))
			Expect(labelsArg).To(Equal(""))
		})

		When("the --domain flag is set", func() {
			BeforeEach(func() {
				cmd.Domain = "apps.example.com"
			})

			It("passes the domain value to the actor", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				_, domainArg, _, _, _ := fakeActor.GetRoutePoliciesForSpaceArgsForCall(0)
				Expect(domainArg).To(Equal("apps.example.com"))
			})
		})
	})
})
