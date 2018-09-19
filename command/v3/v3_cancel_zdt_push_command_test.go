package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("v3-cancel-deployment Command", func() {
	var (
		cmd                      v3.V3CancelZdtPushCommand
		testUI                   *ui.UI
		fakeConfig               *commandfakes.FakeConfig
		fakeSharedActor          *commandfakes.FakeSharedActor
		fakeV3CancelZdtPushActor *v3fakes.FakeV3CancelZdtPushActor
		executeErr               error
		app                      string
		userName                 string
		spaceName                string
		orgName                  string
		binaryName               string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeV3CancelZdtPushActor = new(v3fakes.FakeV3CancelZdtPushActor)

		app = "some-app"
		userName = "banana"
		spaceName = "some-space"
		orgName = "some-org"

		cmd = v3.V3CancelZdtPushCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:                 testUI,
			Config:             fakeConfig,
			CancelZdtPushActor: fakeV3CancelZdtPushActor,
			SharedActor:        fakeSharedActor,
		}
		fakeV3CancelZdtPushActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeV3CancelZdtPushActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionApplicationFlowV3,
			}))
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeV3CancelZdtPushActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: spaceName, GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName, GUID: "some-org-guid"})
			fakeV3CancelZdtPushActor.CancelDeploymentByAppNameAndSpaceReturns(v3action.Warnings{"get-warning"}, errors.New("some-error"))
		})

		It("cancels the deployment", func() {
			Expect(fakeV3CancelZdtPushActor.CancelDeploymentByAppNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGuid := fakeV3CancelZdtPushActor.CancelDeploymentByAppNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal(app))
			Expect(spaceGuid).To(Equal("some-space-guid"))

			Expect(executeErr).To(MatchError("some-error"))
			Expect(testUI.Err).To(Say("get-warning"))
		})

		Context("when the application doesn't exist", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("dropped iphone error")
				fakeV3CancelZdtPushActor.CancelDeploymentByAppNameAndSpaceReturns(v3action.Warnings{"get-warning"}, expectedErr)
			})
			It("displays the warnings and error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Err).To(Say("get-warning"))
				Expect(testUI.Out).ToNot(Say("app some-app in org some-org / space some-space as banana..."))
			})
		})
	})
})
