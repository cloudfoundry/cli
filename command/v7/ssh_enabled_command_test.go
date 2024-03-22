package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("ssh-enabled Command", func() {
	var (
		cmd                 SSHEnabledCommand
		testUI              *ui.UI
		fakeConfig          *commandfakes.FakeConfig
		fakeSharedActor     *commandfakes.FakeSharedActor
		fakeSSHEnabledActor *v7fakes.FakeActor

		appName         string
		binaryName      string
		currentUserName string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeSSHEnabledActor = new(v7fakes.FakeActor)

		cmd = SSHEnabledCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeSSHEnabledActor,
			},
		}

		appName = "some-app"
		cmd.RequiredArgs.AppName = appName

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		currentUserName = "some-user"
		fakeSSHEnabledActor.GetCurrentUserReturns(configv3.User{Name: currentUserName}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking if SSH is enabled succeeds", func() {
		When("SSH is enabled", func() {
			BeforeEach(func() {
				fakeSSHEnabledActor.GetSSHEnabledByAppNameReturns(
					ccv3.SSHEnabled{
						Enabled: true,
						Reason:  "Enabled globally",
					},
					v7action.Warnings{"warning-1"},
					nil,
				)
			})

			It("prints appropriate output", func() {
				Expect(testUI.Out).To(Say(`ssh support is enabled for app '%s'\.`, appName))
				Expect(testUI.Err).To(Say("warning-1"))
			})
		})

		When("SSH is disabled", func() {
			BeforeEach(func() {
				fakeSSHEnabledActor.GetSSHEnabledByAppNameReturns(
					ccv3.SSHEnabled{
						Enabled: false,
						Reason:  "Disabled globally",
					},
					v7action.Warnings{"warning-1"},
					nil,
				)
			})

			It("prints appropriate output", func() {
				Expect(testUI.Out).To(Say(`ssh support is disabled for app '%s'\.`, appName))
				Expect(testUI.Out).To(Say("Disabled globally"))
				Expect(testUI.Err).To(Say("warning-1"))
			})
		})
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("check-target-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("check-target-error"))

			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("checking if SSH is enabled returns an error", func() {
		BeforeEach(func() {
			fakeSSHEnabledActor.GetSSHEnabledByAppNameReturns(
				ccv3.SSHEnabled{},
				v7action.Warnings{"warning-1"},
				errors.New("get-ssh-enabled-error"),
			)
		})

		It("displays warnings and returns the error", func() {
			Expect(testUI.Err).To(Say("warning-1"))
			Expect(executeErr).To(MatchError("get-ssh-enabled-error"))
		})
	})
})
