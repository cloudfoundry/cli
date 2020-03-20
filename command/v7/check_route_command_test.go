package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("check-route Command", func() {
	var (
		cmd             v7.CheckRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCheckRouteActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCheckRouteActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = v7.CheckRouteCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.Domain{Domain: "some-domain.com"},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("no current user"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("no current user"))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("targeting an org and logged in", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		})

		When("checking for existing route returns an error", func() {
			BeforeEach(func() {
				fakeActor.CheckRouteReturns(
					false,
					v7action.Warnings{"check-route-warning"},
					errors.New("failed to check route"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("failed to check route"))

				Expect(testUI.Out).To(Say("Checking for route..."))

				Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
				givenDomain, givenHostname, givenPath := fakeActor.CheckRouteArgsForCall(0)
				Expect(givenDomain).To(Equal("some-domain.com"))
				Expect(givenHostname).To(Equal(""))
				Expect(givenPath).To(Equal(""))
			})
		})

		When("checking for route returns true", func() {
			BeforeEach(func() {
				fakeActor.CheckRouteReturns(
					true,
					v7action.Warnings{"check-route-warning"},
					nil,
				)
			})

			It("checks the route and displays the result", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Checking for route..."))
				Expect(testUI.Out).To(Say(`Route 'some-domain.com' does exist\.`))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
				givenDomain, givenHostname, givenPath := fakeActor.CheckRouteArgsForCall(0)
				Expect(givenDomain).To(Equal("some-domain.com"))
				Expect(givenHostname).To(Equal(""))
				Expect(givenPath).To(Equal(""))
			})
		})

		When("checking for route returns false", func() {
			BeforeEach(func() {
				fakeActor.CheckRouteReturns(
					false,
					v7action.Warnings{"check-route-warning"},
					nil,
				)
			})

			It("checks the route and displays the result", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Checking for route..."))
				Expect(testUI.Out).To(Say(`Route 'some-domain\.com' does not exist\.`))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
				givenDomain, givenHostname, givenPath := fakeActor.CheckRouteArgsForCall(0)
				Expect(givenDomain).To(Equal("some-domain.com"))
				Expect(givenHostname).To(Equal(""))
				Expect(givenPath).To(Equal(""))
			})
		})

		When("passing hostname and path flags", func() {
			BeforeEach(func() {
				cmd.Path.Path = "/some-path"
				cmd.Hostname = "some-host"

				fakeActor.CheckRouteReturns(
					true,
					v7action.Warnings{"check-route-warning"},
					nil,
				)
			})

			It("checks the route with correct arguments and displays the result", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Checking for route..."))
				Expect(testUI.Out).To(Say(`Route 'some-host\.some-domain\.com/some-path' does exist\.`))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
				givenDomain, givenHostname, givenPath := fakeActor.CheckRouteArgsForCall(0)
				Expect(givenDomain).To(Equal("some-domain.com"))
				Expect(givenHostname).To(Equal("some-host"))
				Expect(givenPath).To(Equal("/some-path"))
			})
		})
	})
})
