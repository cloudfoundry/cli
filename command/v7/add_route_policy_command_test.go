package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("add-route-policy Command", func() {
	var (
		cmd             AddRoutePolicyCommand
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

		cmd = AddRoutePolicyCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.AddRoutePolicyArgs{Domain: "apps.example.com"},
			Hostname:     "myapp",
			RoutePolicySourceFlags: RoutePolicySourceFlags{
				SourceAny: true,
			},
		}

		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.APIVersionReturns("3.999.0")
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
				MinimumVersion: "3.999.0",
			}))
		})
	})

	When("source flag validation fails", func() {
		BeforeEach(func() {
			cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.RequiredArgumentError{
				ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --source",
			}))
		})
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: "faceman"})
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
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

	When("the happy path", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeActor.AddRoutePolicyReturns(v7action.Warnings{"some-warning"}, nil)
		})

		It("displays the progress message and calls actor with right args", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Adding route policy"))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.AddRoutePolicyCallCount()).To(Equal(1))
			domainArg, sourceArg, hostnameArg, pathArg := fakeActor.AddRoutePolicyArgsForCall(0)
			Expect(domainArg).To(Equal("apps.example.com"))
			Expect(sourceArg).To(Equal("cf:any"))
			Expect(hostnameArg).To(Equal("myapp"))
			Expect(pathArg).To(Equal(""))
		})
	})

	When("adding the route policy fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeActor.AddRoutePolicyReturns(v7action.Warnings{"some-warning"}, errors.New("add-error"))
		})

		It("returns the error and displays warnings", func() {
			Expect(executeErr).To(MatchError("add-error"))
			Expect(testUI.Err).To(Say("some-warning"))
		})
	})
})
