package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("buildpacks Command", func() {
	var (
		cmd             BuildpacksCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeBuildpacksActor
		executeErr      error
		args            []string
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeBuildpacksActor)
		args = nil

		cmd = BuildpacksCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
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
			Expect(testUI.Out).To(Say(`Getting buildpacks as apple\.\.\.`))
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {
				fakeActor.GetBuildpacksReturns(nil, v7action.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError("some-error"))

				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		When("getting buildpacks succeeds", func() {
			When("buildpacks exist", func() {
				BeforeEach(func() {
					buildpacks := []v7action.Buildpack{
						{Name: "buildpack-1", Position: 1, Enabled: true, Locked: false, Filename: "buildpack-1.file", Stack: "buildpack-1-stack"},
						{Name: "buildpack-2", Position: 2, Enabled: false, Locked: true, Filename: "buildpack-2.file", Stack: ""},
					}
					fakeActor.GetBuildpacksReturns(buildpacks, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
				})
				It("prints a table of buildpacks", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).To(Say(`position\s+name\s+stack\s+enabled\s+locked\s+filename`))
					Expect(testUI.Out).To(Say(`1\s+buildpack-1\s+buildpack-1-stack\s+true\s+false\s+buildpack-1.file`))
					Expect(testUI.Out).To(Say(`2\s+buildpack-2\s+false\s+true\s+buildpack-2.file`))
				})
			})
			When("there are no buildpacks", func() {
				BeforeEach(func() {
					buildpacks := []v7action.Buildpack{}
					fakeActor.GetBuildpacksReturns(buildpacks, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
				})
				It("prints a table of buildpacks", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).To(Say("No buildpacks found"))
				})
			})
		})
	})
})
