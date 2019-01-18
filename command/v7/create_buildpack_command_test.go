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

var _ = Describe("create buildpack Command", func() {
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
			RequiredArgs: flag.CreateBuildpackArgs{Buildpack: "some-buildpack", Path: "/path/to/buildpack.zip", Position: 7},
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
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("should print text indicating it is creating a buildpack", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating buildpack some-buildpack as the-user\.\.\.`))
		})

		When("creating the buildpack fails", func() {
			BeforeEach(func() {
				fakeActor.CreateBuildpackReturns(
					v7action.Buildpack{},
					v7action.Warnings{"warning-1"},
					actionerror.BuildpackNameTakenError{Name: "this-error-occurred"},
				)
			})
			It("it errors and prints all warnings", func() {
				Expect(executeErr).To(Equal(actionerror.BuildpackNameTakenError{Name: "this-error-occurred"}))
				Expect(testUI.Err).To(Say("warning-1"))
			})
		})

		When("creating buildpack succeeds", func() {
			BeforeEach(func() {
				buildpack := v7action.Buildpack{Name: "buildpack-1", Position: 1, Enabled: true, Locked: false, Filename: "buildpack-1.file", Stack: "buildpack-1-stack"}
				fakeActor.CreateBuildpackReturns(buildpack, v7action.Warnings{"some-create-warning-1"}, nil)
			})

			It("correctly created the buildpack", func() {
				buildpack := fakeActor.CreateBuildpackArgsForCall(0)
				Expect(buildpack.Name).To(Equal("some-buildpack"))
				Expect(buildpack.Position).To(Equal(7))
			})

			It("prints any warnings and uploads the bits", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("some-create-warning-1"))
				Expect(testUI.Out).To(Say(`Uploading buildpack some-buildpack as the-user\.\.\.`))
			})

			When("Uploading the buildpack fails", func() {
				BeforeEach(func() {
					fakeActor.UploadBuildpackReturns(
						v7action.Warnings{"warning-2"},
						actionerror.BuildpackStackChangeError{
							BuildpackName: "buildpack-name",
							BinaryName:    "faceman"},
					)
				})

				It("it errors and prints all warnings", func() {
					Expect(executeErr).To(Equal(actionerror.BuildpackStackChangeError{
						BuildpackName: "buildpack-name",
						BinaryName:    "faceman",
					}))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("Uploading the buildpack succeeds", func() {
				BeforeEach(func() {
					fakeActor.UploadBuildpackReturns(
						v7action.Warnings{"some-upload-warning-1"},
						nil,
					)
				})

				It("prints all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-upload-warning-1"))
				})
			})
		})
	})
})
