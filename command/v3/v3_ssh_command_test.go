package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-ssh Command", func() {
	var (
		cmd             v3.V3SSHCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3SSHActor
		executeErr      error
		appName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3SSHActor)

		appName = "some-app"
		cmd = v3.V3SSHCommand{
			RequiredArgs: flag.AppName{AppName: appName},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionV3,
			}))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: "steve"})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: "steve"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is targeted to an organization and space", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "some-space-guid"})
		})

		Context("when executing the secure shell fails", func() {
			BeforeEach(func() {
				fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexReturns(v3action.Warnings{"some-warnings"}, errors.New("some-error"))
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(testUI.Err).To(Say("some-warnings"))
			})
		})

		Context("when executing the secure shell succeeds", func() {
			BeforeEach(func() {
				fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexReturns(v3action.Warnings{"some-warnings"}, nil)
			})

			It("returns nil and displays all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warnings"))

				Expect(fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg, processTypeArg, processIndexArg, sshOptionsArg := fakeActor.ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndexArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
				Expect(processTypeArg).To(Equal("web"))
				Expect(processIndexArg).To(Equal(uint(0)))
				Expect(sshOptionsArg).To(Equal(v3action.SSHOptions{}))
			})
		})
	})
})
