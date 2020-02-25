package v6_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/ui"
)

var _ = Describe("DeleteBuildpackCommand", func() {
	var (
		cmd             DeleteBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeDeleteBuildpackActor

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeDeleteBuildpackActor)

		cmd = DeleteBuildpackCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Buildpack = "bp-name"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("a stack is specified", func() {
		BeforeEach(func() {
			cmd.Stack = "some-stack"
		})

		When("the api version is below minimum for stack association", func() {
			BeforeEach(func() {
				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinSupportedV2ClientVersion)
			})

			It("returns a version error", func() {
				Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
					Command:        "Option '-s'",
					CurrentVersion: ccversion.MinSupportedV2ClientVersion,
					MinimumVersion: ccversion.MinVersionBuildpackStackAssociationV2,
				}))
			})
		})

		When("the api version is at or above minimum for stack association", func() {
			BeforeEach(func() {
				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionBuildpackStackAssociationV2)
			})

			It("returns the unrefactored command error", func() {
				Expect(executeErr).To(MatchError(translatableerror.UnrefactoredCommandError{}))
			})
		})
	})
})
