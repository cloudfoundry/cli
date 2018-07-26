package v2_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
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

		cmd.RequiredArgs.Buildpack = "bp-name"
		cmd.RequiredArgs.Position = 3

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
			var fakeUser configv3.User

			BeforeEach(func() {
				fakeUser = configv3.User{Name: "some-user"}
				fakeConfig.CurrentUserReturns(fakeUser, nil)
			})

			Context("when creating the buildpack fails because a buildpack already exists", func() {
				BeforeEach(func() {
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"some-create-bp-warning"}, actionerror.BuildpackNameTakenError("bp-name"))
				})

				It("prints the error message as a warning but does not return it", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("some-create-bp-warning"))
					Expect(testUI.Err).To(Say("Buildpack bp-name already exists"))
					Expect(testUI.Out).To(Say("TIP: use 'faceman update-buildpack' to update this buildpack"))
				})
			})

			Context("when creating the buildpack fails because a buildpack with the nil stack already exists", func() {
				BeforeEach(func() {
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"some-create-bp-warning"}, actionerror.BuildpackAlreadyExistsWithoutStackError("bp-name"))
					cmd.RequiredArgs.Buildpack = "bp-name"
				})

				It("prints the error message as a warning but does not return it", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("some-create-bp-warning"))
					Expect(testUI.Err).To(Say("Buildpack bp-name already exists without a stack"))
					Expect(testUI.Out).To(Say("TIP: use 'faceman buildpacks' and 'faceman delete-buildpack' to delete buildpack bp-name without a stack"))
				})
			})

			Context("when creating the buildpack fails with a generic error", func() {
				BeforeEach(func() {
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"some-create-bp-warning"}, errors.New("some-create-bp-error"))
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError("some-create-bp-error"))
					Expect(testUI.Err).To(Say("some-create-bp-warning"))
				})
			})

			Context("when creating the buildpack succeeds", func() {
				BeforeEach(func() {
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{GUID: "some-guid"}, v2action.Warnings{"some-create-bp-warning"}, nil)
				})

				Context("when a zip is provided as the path", func() {
					BeforeEach(func() {
						cmd.RequiredArgs.Path = "some.zip"
					})

					It("displays that the buildpack was created successfully", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("OK"))

						Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))
						bpName, bpPosition, enabled := fakeActor.CreateBuildpackArgsForCall(0)
						Expect(bpName).To(Equal("bp-name"))
						Expect(bpPosition).To(Equal(3))
						Expect(enabled).To(Equal(true))
					})
				})

				Context("when a directory is provided as the path", func() {
					BeforeEach(func() {
						cmd.RequiredArgs.Path = "some/path"
					})

					It("displays that the buildpack was created successfully", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("OK"))

						Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))
						bpName, bpPosition, enabled := fakeActor.CreateBuildpackArgsForCall(0)
						Expect(bpName).To(Equal("bp-name"))
						Expect(bpPosition).To(Equal(3))
						Expect(enabled).To(Equal(true))
					})
				})

				Context("when a URL is provided as the path", func() {
					BeforeEach(func() {
						cmd.RequiredArgs.Path = "https://foo.buildpack.zip"
					})

					It("displays that the buildpack was created successfully", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("OK"))

						Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))
						bpName, bpPosition, enabled := fakeActor.CreateBuildpackArgsForCall(0)
						Expect(bpName).To(Equal("bp-name"))
						Expect(bpPosition).To(Equal(3))
						Expect(enabled).To(Equal(true))
					})

					Context("when preparing the buildpack bits fails", func() {
						BeforeEach(func() {
							fakeActor.PrepareBuildpackBitsReturns("some/invalid/path", errors.New("some-prepare-bp-error"))
						})

						It("returns an error", func() {
							Expect(executeErr).To(MatchError("some-prepare-bp-error"))
							Expect(fakeActor.PrepareBuildpackBitsCallCount()).To(Equal(1))
							path, _, _ := fakeActor.PrepareBuildpackBitsArgsForCall(0)
							Expect(path).To(Equal("https://foo.buildpack.zip"))
						})
					})

					Context("when preparing the buildpack bits succeeds", func() {
						BeforeEach(func() {
							fakeActor.PrepareBuildpackBitsReturns("foo.buildpack.zip", nil)
						})

						It("displays that upload is starting", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say("Uploading buildpack bp-name as some-user"))

							Expect(fakeActor.PrepareBuildpackBitsCallCount()).To(Equal(1))
							path, _, _ := fakeActor.PrepareBuildpackBitsArgsForCall(0)
							Expect(path).To(Equal("https://foo.buildpack.zip"))
						})

						Context("when uploading the buildpack fails because a buildpack with that stack already exists", func() {
							BeforeEach(func() {
								fakeActor.UploadBuildpackReturns(v2action.Warnings{"some-upload-bp-warning"}, actionerror.BuildpackAlreadyExistsForStackError{Message: "bp error"})
							})

							It("prints the error message as a warning but does not return it", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Err).To(Say("some-upload-bp-warning"))
								Expect(testUI.Err).To(Say("bp error"))
								Expect(testUI.Out).To(Say("TIP: use 'faceman update-buildpack' to update this buildpack"))
							})
						})

						Context("when uploading the buildpack fails with a generic error", func() {
							BeforeEach(func() {
								fakeActor.UploadBuildpackReturns(v2action.Warnings{"some-upload-bp-warning"}, errors.New("some-upload-bp-error"))
							})

							It("returns an error and warnings", func() {
								Expect(executeErr).To(MatchError("some-upload-bp-error"))
								Expect(testUI.Err).To(Say("some-create-bp-warning"))
								Expect(testUI.Err).To(Say("some-upload-bp-warning"))
							})

						})

						Context("when uploading the buildpack succeeds", func() {
							BeforeEach(func() {
								fakeActor.UploadBuildpackReturns(v2action.Warnings{"some-upload-bp-warning"}, nil)
							})

							It("displays that the buildpack was uploaded successfully", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).To(Say("Done uploading"))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Err).To(Say("some-upload-bp-warning"))

								Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(1))
								guid, path, _ := fakeActor.UploadBuildpackArgsForCall(0)
								Expect(guid).To(Equal("some-guid"))
								Expect(path).To(Equal("foo.buildpack.zip"))
							})
						})
					})

				})

			})

			Context("when both --enable and --disable are provided", func() {
				BeforeEach(func() {
					cmd.Enable = true
					cmd.Disable = true
				})

				It("returns an argument combination error", func() {
					argumentCombinationError := translatableerror.ArgumentCombinationError{
						Args: []string{"--enable", "--disable"},
					}
					Expect(executeErr).To(MatchError(argumentCombinationError))
				})
			})

			Context("when --enable is provided", func() {
				BeforeEach(func() {
					cmd.Enable = true
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{GUID: "some-guid"}, v2action.Warnings{"some-create-bp-warning"}, nil)
				})

				It("successfully creates a buildpack with enabled set to true", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Uploading buildpack bp-name as some-user"))
					Expect(testUI.Out).To(Say("Done uploading"))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))
					_, _, enabled := fakeActor.CreateBuildpackArgsForCall(0)
					Expect(enabled).To(BeTrue())
				})
			})

			Context("when --disable is provided", func() {
				BeforeEach(func() {
					cmd.Disable = true
					fakeActor.CreateBuildpackReturns(v2action.Buildpack{GUID: "some-guid"}, v2action.Warnings{"some-create-bp-warning"}, nil)
				})

				It("successfully creates a buildpack with enabled set to false", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Uploading buildpack bp-name as some-user"))
					Expect(testUI.Out).To(Say("Done uploading"))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateBuildpackCallCount()).To(Equal(1))
					_, _, enabled := fakeActor.CreateBuildpackArgsForCall(0)
					Expect(enabled).To(BeFalse())
				})
			})

		})
	})

})
