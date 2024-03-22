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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("check-route Command", func() {
	var (
		cmd             v7.CheckRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = v7.CheckRouteCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Domain{Domain: "some-domain.com"},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		fakeActor.CheckRouteReturns(
			true,
			v7action.Warnings{"check-route-warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the target", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(checkTargetedOrg).To(BeTrue())
		Expect(checkTargetedSpace).To(BeFalse())
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
		})
	})

	It("checks if the user is logged in", func() {
		Expect(fakeActor.GetCurrentUserCallCount()).To(Equal(1))
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("no current user"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("no current user"))
		})
	})

	It("displays it's checking for the route", func() {
		Expect(testUI.Out).To(Say("Checking for route..."))
	})

	It("checks if the route exists, displaying warnings", func() {
		Expect(testUI.Err).To(Say("check-route-warning"))

		Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
		domain, host, path, port := fakeActor.CheckRouteArgsForCall(0)
		Expect(domain).To(Equal(cmd.RequiredArgs.Domain))
		Expect(host).To(Equal(cmd.Hostname))
		Expect(path).To(Equal(cmd.Path.Path))
		Expect(port).To(Equal(0))
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
		})
	})

	It("confirms that the route exists", func() {
		Expect(testUI.Out).To(Say(`Route 'some-domain.com' does exist\.`))
		Expect(testUI.Out).To(Say("OK"))
	})

	When("there's no matching route", func() {
		BeforeEach(func() {
			fakeActor.CheckRouteReturns(
				false,
				v7action.Warnings{"check-route-warning"},
				nil,
			)
		})

		It("checks the route and displays the result", func() {
			Expect(testUI.Out).To(Say(`Route 'some-domain\.com' does not exist\.`))
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
			givenDomain, givenHostname, givenPath, givenPort := fakeActor.CheckRouteArgsForCall(0)
			Expect(givenDomain).To(Equal("some-domain.com"))
			Expect(givenHostname).To(Equal("some-host"))
			Expect(givenPath).To(Equal("/some-path"))
			Expect(givenPort).To(Equal(0))
		})
	})

	When("passing in a port flag (for TCP routes)", func() {
		BeforeEach(func() {
			cmd.Port = 1024

			fakeActor.CheckRouteReturns(
				true,
				v7action.Warnings{"check-route-warning"},
				nil,
			)
		})

		It("checks the route with correct arguments and displays the result", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Checking for route..."))
			Expect(testUI.Out).To(Say(`Route 'some-domain\.com:1024' does exist\.`))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.CheckRouteCallCount()).To(Equal(1))
			_, _, _, givenPort := fakeActor.CheckRouteArgsForCall(0)
			Expect(givenPort).To(Equal(1024))
		})
	})

	It("displays 'OK' and returns nil (no error)", func() {
		Expect(testUI.Out).To(Say("OK"))

		Expect(executeErr).NotTo(HaveOccurred())
	})
})
