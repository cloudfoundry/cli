package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = PDescribe("create buildpack Command", func() {
	var (
		cmd             CreateBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateBuildpackActor
		executeErr      error
		args            []string
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateBuildpackActor)
		args = nil

		cmd = CreateBuildpackCommand{
			RequiredArgs: flag.BuildpackName{Buildpack: "some-app"},
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "apple"}, nil)
		})

		It("should print text indicating its runnning", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating buildpack NAME as USER\.\.\.`))
		})

		When("creating the buildpack fails", func() {
			It("it errors and prints all warnings", func() {
				Expect(executeErr).To(HaveOccurred())
			})

		})

		When("creating buildpack succeeds", func() {
			BeforeEach(func() {
				buildpack := v7action.Buildpack{Name: "buildpack-1", Position: 1, Enabled: true, Locked: false, Filename: "buildpack-1.file", Stack: "buildpack-1-stack"}
				fakeActor.CreateBuildpackReturns(buildpack, v7action.Warnings{"some-create-warning-1"}, nil)
			})

			It("prints any warnings and uploads the bits", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("some-create-warning-1"))
				Expect(testUI.Out).To(Say(`Uploading buildpack NAME as USER\.\.\.`))
			})

			When("Uploading the buildpack fails", func() {

			})

			When("Uploading the buildpack succeeds", func() {
				BeforeEach(func() {
					// set upload returns
				})

				It("prints all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-upload-warning-1"))
				})

				When("Polling the job fails", func() {

				})

				When("Polling the job times out", func() {

				})

				When("Polling the job succeeds", func() {
					It("does not error and prints warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(testUI.Err).To(Say("some-poll-warning-1"))
						Expect(testUI.Out).To(Say(`Done uploading`))
					})
				})
			})
		})
	})
})
