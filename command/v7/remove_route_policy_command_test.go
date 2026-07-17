package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
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

var _ = Describe("remove-route-policy Command", func() {
	var (
		cmd             RemoveRoutePolicyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = RemoveRoutePolicyCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.RoutePolicyArgs{Domain: "apps.example.com"},
			Hostname:     "myapp",
			RoutePolicySourceFlags: RoutePolicySourceFlags{
				SourceAny: true,
			},
		}

		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.APIVersionReturns("3.999.0")
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

	When("source flag validation fails", func() {
		BeforeEach(func() {
			cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{}
		})

		It("returns an error", func() {
			Expect(executeErr).To(HaveOccurred())
		})
	})

	When("--source-org is used with --source-app but without --source-space", func() {
		BeforeEach(func() {
			cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{
				SourceApp: "my-app",
				SourceOrg: "my-org",
			}
		})

		It("returns a RequiredFlagsError for --source-space", func() {
			Expect(executeErr).To(MatchError(translatableerror.RequiredFlagsError{
				Arg1: "--source-org",
				Arg2: "--source-space",
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

	When("the -f flag is set (force)", func() {
		BeforeEach(func() {
			cmd.Force = true
			fakeActor.DeleteRoutePolicyBySourceReturns(v7action.Warnings{"some-warning"}, nil)
		})

		It("removes the policy without prompting and displays OK", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Removing route policy"))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.DeleteRoutePolicyBySourceCallCount()).To(Equal(1))
			domainArg, sourceArg, hostnameArg, pathArg := fakeActor.DeleteRoutePolicyBySourceArgsForCall(0)
			Expect(domainArg).To(Equal("apps.example.com"))
			Expect(sourceArg).To(Equal("cf:any"))
			Expect(hostnameArg).To(Equal("myapp"))
			Expect(pathArg).To(Equal(""))
		})
	})

	When("the -f flag is NOT set", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		When("the user answers no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not remove the policy", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("Really remove route policy"))
				Expect(testUI.Out).To(Say("Route policy has not been removed\\."))
				Expect(fakeActor.DeleteRoutePolicyBySourceCallCount()).To(Equal(0))
			})
		})

		When("the user answers yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).NotTo(HaveOccurred())
				fakeActor.DeleteRoutePolicyBySourceReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("removes the policy and displays OK", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("Really remove route policy"))
				Expect(testUI.Out).To(Say("Removing route policy"))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeActor.DeleteRoutePolicyBySourceCallCount()).To(Equal(1))
				domainArg, sourceArg, hostnameArg, pathArg := fakeActor.DeleteRoutePolicyBySourceArgsForCall(0)
				Expect(domainArg).To(Equal("apps.example.com"))
				Expect(sourceArg).To(Equal("cf:any"))
				Expect(hostnameArg).To(Equal("myapp"))
				Expect(pathArg).To(Equal(""))
			})
		})
	})

	When("deleting the route policy fails", func() {
		BeforeEach(func() {
			cmd.Force = true
			fakeActor.DeleteRoutePolicyBySourceReturns(v7action.Warnings{"some-warning"}, errors.New("delete-error"))
		})

		It("returns the error and displays warnings", func() {
			Expect(executeErr).To(MatchError("delete-error"))
			Expect(testUI.Err).To(Say("some-warning"))
		})
	})
})
