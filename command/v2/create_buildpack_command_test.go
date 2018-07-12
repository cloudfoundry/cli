package v2_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
)

var _ = Describe("CreateBuildpackCommand", func() {
	var (
		cmd             CreateBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeCreateBuildpackActor
		input           *Buffer
		binaryName      string

		executeErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeCreateBuildpackActor)

		cmd = CreateBuildpackCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeFalse())
			Expect(checkTargetedSpaceArg).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		Context("when getting the current user fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error that happened")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})
		})

		Context("when getting the current user succeeds", func() {
			var (
				fakeUser configv3.User
			)

			BeforeEach(func() {
				fakeUser = configv3.User{Name: "some-user"}
				fakeConfig.CurrentUserReturns(fakeUser, nil)
			})

			Context("when creating the buildpack fails", func() {
				BeforeEach(func() {
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"some-create-bp-warning"}, errors.New("some-create-bp-error"))
				})

				It("returns an error and warnings", func() {
					Expect(testUI.Err).To(Say("some-create-bp-warning"))
					Expect(executeErr).To(MatchError("some-create-bp-error"))
				})
			})

			Context("when creating the buildpack succeeds", func() {
				BeforeEach(func() {
					cmd.RequiredArgs.Buildpack = "bp-name"
					cmd.RequiredArgs.Position.Value = 3
					cmd.RequiredArgs.Path = "some/path"
					cmd.Disable = true
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{GUID: "some-guid"}, v2action.Warnings{"some-create-bp-warning"}, nil)
				})

				It("displays that the buildpack was created successfully", func() {
					Expect(testUI.Out).To(Say("OK"))
					Expect(executeErr).ToNot(HaveOccurred())
				})

				Context("when uploading the buildpack fails", func() {
					BeforeEach(func() {
						fakeActor.UploadBuildpackReturns(v2action.Warnings{"some-upload-bp-warning"}, errors.New("some-upload-bp-error"))
					})

					It("returns an error and warnings", func() {
						Expect(testUI.Err).To(Say("some-create-bp-warning"))
						Expect(testUI.Err).To(Say("some-upload-bp-warning"))
						Expect(executeErr).To(MatchError("some-upload-bp-error"))
					})

				})

				Context("when uploading the buildpack succeeds", func() {
					It("displays that the buildpack was uploaded successfully", func() {
						Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))

						name, position, enabled := fakeActor.CreateBuildpackArgsForCall(0)
						Expect(name).To(Equal("bp-name"))
						Expect(position).To(Equal(3))
						Expect(enabled).To(BeFalse())

						guid, path, _ := fakeActor.UploadBuildpackArgsForCall(0)
						Expect(guid).To(Equal("some-guid"))
						Expect(path).To(Equal("some/path"))

						Expect(testUI.Out).To(Say("Done uploading"))
						Expect(testUI.Out).To(Say("OK"))
						Expect(executeErr).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
})
