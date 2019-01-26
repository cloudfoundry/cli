package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"errors"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UpdateBuildpackCommand", func() {
	var (
		cmd             UpdateBuildpackCommand
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		input           *Buffer
		fakeActor       *v7fakes.FakeUpdateBuildpackActor
		fakeConfig      *commandfakes.FakeConfig
		buildpackGUID   = "buildpack-guid"
		buildpackName   = "some-bp"
		binaryName      = "faceman"

		executeErr  error
		expectedErr error
	)

	BeforeEach(func() {
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeUpdateBuildpackActor)
		fakeConfig = new(commandfakes.FakeConfig)
		buildpackGUID = "some guid"

		cmd = UpdateBuildpackCommand{
			RequiredArgs: flag.BuildpackName{Buildpack: buildpackName},
			UI:           testUI,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			Config:       fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("invalid flag combinations", func() {
		When("the --lock and --unlock flags are provided", func() {
			BeforeEach(func() {
				cmd.Lock = true
				cmd.Unlock = true
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--lock", "--unlock"},
				}))
			})
		})

		When("the --enable and --disable flags are provided", func() {
			BeforeEach(func() {
				cmd.Enable = true
				cmd.Disable = true
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--enable", "--disable"},
				}))
			})
		})

		When("the --path and --lock flags are provided", func() {
			BeforeEach(func() {
				cmd.Lock = true
				cmd.Path = "asdf"
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--path", "--lock"},
				}))
			})
		})

		When("the --path and --assign-stack flags are provided", func() {
			BeforeEach(func() {
				cmd.Path = "asdf"
				cmd.NewStack = "some-new-stack"
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--path", "--assign-stack"},
				}))
			})
		})

		When("the --stack and --assign-stack flags are provided", func() {
			BeforeEach(func() {
				cmd.CurrentStack = "current-stack"
				cmd.NewStack = "some-new-stack"
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--stack", "--assign-stack"},
				}))
			})
		})
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
		When("getting the current user fails", func() {
			BeforeEach(func() {
				expectedErr = errors.New("some-error that happened")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("getting the current user succeeds", func() {
			var userName string

			BeforeEach(func() {
				userName = "some-user"
				fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)
			})

			When("preparing buildpack bits causes an error", func() {
				var emptyDirectoryError error
				BeforeEach(func() {
					emptyDirectoryError = actionerror.EmptyBuildpackDirectoryError{Path: "some-directory"}
					fakeActor.PrepareBuildpackBitsReturns("", emptyDirectoryError)
					cmd.Path = "some empty directory"
				})

				It("exits without updating if the path points to an empty directory", func() {
					Expect(executeErr).To(MatchError(emptyDirectoryError))
					Expect(fakeActor.UpdateBuildpackByNameAndStackCallCount()).To(Equal(0))
				})
			})

			When("updating the buildpack fails", func() {
				BeforeEach(func() {
					cmd.Path = "path/to/buildpack"
					fakeActor.PrepareBuildpackBitsReturns("path/to/prepared/bits", nil)
					expectedErr = errors.New("update-error")
					fakeActor.UpdateBuildpackByNameAndStackReturns(
						v7action.Buildpack{},
						v7action.Warnings{"update-bp-warning1", "update-bp-warning2"},
						expectedErr,
					)
				})

				It("returns the error and prints any warnings", func() {
					Expect(testUI.Err).To(Say("update-bp-warning1"))
					Expect(testUI.Err).To(Say("update-bp-warning2"))
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(0))
				})
			})

			When("the --lock flag is provided", func() {
				BeforeEach(func() {
					cmd.Lock = true
				})

				It("sets the locked value to true when updating the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Locked.IsSet).To(Equal(true))
					Expect(buildpack.Locked.Value).To(Equal(true))
				})
			})

			When("the --unlock flag is provided", func() {
				BeforeEach(func() {
					cmd.Unlock = true
				})

				It("sets the locked value to false when updating the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Locked.IsSet).To(Equal(true))
					Expect(buildpack.Locked.Value).To(Equal(false))
				})
			})

			When("the --enable flag is provided", func() {
				BeforeEach(func() {
					cmd.Enable = true
				})

				It("sets the enabled value to true when updating the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Enabled.IsSet).To(Equal(true))
					Expect(buildpack.Enabled.Value).To(Equal(true))
				})
			})

			When("the --disable flag is provided", func() {
				BeforeEach(func() {
					cmd.Disable = true
				})

				It("sets the enabled value to false when updating the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Enabled.IsSet).To(Equal(true))
					Expect(buildpack.Enabled.Value).To(Equal(false))
				})
			})

			When("the --index flag is provided", func() {
				BeforeEach(func() {
					cmd.Position = types.NullInt{IsSet: true, Value: 99}
				})

				It("sets the new buildpack order when updating the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Position.IsSet).To(Equal(true))
					Expect(buildpack.Position.Value).To(Equal(99))
				})
			})

			When("the --assign-stack flag is provided", func() {
				BeforeEach(func() {
					cmd.NewStack = "some-new-stack"
				})

				It("sets the new stack on the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(testUI.Out).ToNot(Say("Updating buildpack %s", buildpackName))
					Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
					Expect(buildpack.Stack).To(Equal("some-new-stack"))
				})

				Context("and the --index flag is provided", func() {
					BeforeEach(func() {
						cmd.Position = types.NullInt{IsSet: true, Value: 3}
					})

					It("sets the new stack and updates the priority of the buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
						Expect(testUI.Out).To(Say("Updating buildpack %s with stack %s...", buildpackName, cmd.NewStack))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("and the --lock flag is provided", func() {
					BeforeEach(func() {
						cmd.Lock = true
					})

					It("sets the new stack and locks the buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
						Expect(testUI.Out).To(Say("Updating buildpack %s with stack %s...", buildpackName, cmd.NewStack))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("and the --unlock flag is provided", func() {
					BeforeEach(func() {
						cmd.Unlock = true
					})

					It("sets the new stack and unlocks the buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
						Expect(testUI.Out).To(Say("Updating buildpack %s with stack %s...", buildpackName, cmd.NewStack))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("and the --enable flag is provided", func() {
					BeforeEach(func() {
						cmd.Enable = true
					})

					It("sets the new stack and enables the buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
						Expect(testUI.Out).To(Say("Updating buildpack %s with stack %s...", buildpackName, cmd.NewStack))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				Context("and the --disable flag is provided", func() {
					BeforeEach(func() {
						cmd.Disable = true
					})

					It("sets the new stack and disables the buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Assigning stack %s to %s as %s...", cmd.NewStack, buildpackName, userName))
						Expect(testUI.Out).To(Say("Updating buildpack %s with stack %s...", buildpackName, cmd.NewStack))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

			})

			When("the --rename flag is provided", func() {
				BeforeEach(func() {
					cmd.NewName = "new-buildpack-name"
				})

				It("sets the new name on the buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
					Expect(buildpack.Name).To(Equal("new-buildpack-name"))

					Expect(testUI.Out).ToNot(Say("Updating buildpack %s", buildpackName))
					Expect(testUI.Out).To(Say(
						"Renaming buildpack %s to %s as %s...", buildpackName, cmd.NewName, userName))
					Expect(testUI.Out).To(Say("OK"))
				})

				Context("and the --assign-stack flag is provided", func() {
					BeforeEach(func() {
						cmd.NewStack = "new-stack"
					})

					It("sets the new name/stack on the buildpack and refers to the new name going forward", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						_, _, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
						Expect(buildpack.Name).To(Equal("new-buildpack-name"))
						Expect(buildpack.Stack).To(Equal("new-stack"))

						Expect(testUI.Out).To(Say(
							"Renaming buildpack %s to %s as %s...", buildpackName, cmd.NewName, userName))

						Expect(testUI.Out).To(Say(
							"Assigning stack %s to %s as %s", cmd.NewStack, cmd.NewName, userName))

						Expect(testUI.Out).ToNot(Say("Updating buildpack %s", buildpackName))

						Expect(testUI.Out).To(Say("OK"))
					})
				})
			})

			When("updating the buildpack succeeds", func() {
				BeforeEach(func() {
					fakeActor.UpdateBuildpackByNameAndStackReturns(
						v7action.Buildpack{GUID: buildpackGUID},
						v7action.Warnings{"update-bp-warning1", "update-bp-warning2"},
						nil,
					)
				})

				When("no arguments are specified", func() {
					It("makes the actor call to update the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackByNameAndStackCallCount()).To(Equal(1))
						_, newStack, buildpack := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
						Expect(buildpack.Name).To(Equal(""))
						Expect(buildpack.Stack).To(Equal(""))
						Expect(buildpack.Position.IsSet).To(BeFalse())
						Expect(buildpack.Locked.IsSet).To(BeFalse())
						Expect(buildpack.Enabled.IsSet).To(BeFalse())
						Expect(newStack).To(Equal(""))

						Expect(testUI.Err).To(Say("update-bp-warning1"))
						Expect(testUI.Err).To(Say("update-bp-warning2"))
						Expect(testUI.Out).To(Say("Updating buildpack %s as %s...", buildpackName, userName))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("a path is specified", func() {
					BeforeEach(func() {
						cmd.Path = "some path"
					})

					It("makes the call to update the buildpack", func() {
						Expect(fakeActor.UpdateBuildpackByNameAndStackCallCount()).To(Equal(1))
						buildpackNameArg, _, _ := fakeActor.UpdateBuildpackByNameAndStackArgsForCall(0)
						Expect(buildpackNameArg).To(Equal(buildpackName))

						Expect(testUI.Err).To(Say("update-bp-warning1"))
						Expect(testUI.Err).To(Say("update-bp-warning2"))
						Expect(testUI.Out).To(Say("Updating buildpack %s as %s...", buildpackName, userName))
						Expect(testUI.Out).To(Say("OK"))
					})

					When("preparing the bits succeeds", func() {
						buildpackBitsPath := "some path on the file system"
						BeforeEach(func() {
							fakeActor.PrepareBuildpackBitsReturns(buildpackBitsPath, nil)
						})

						It("uploads the new buildpack bits", func() {
							Expect(testUI.Out).To(Say("Uploading buildpack some-bp as some-user..."))
							Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(1))
							buildpackGUIDUsed, pathUsed, _ := fakeActor.UploadBuildpackArgsForCall(0)
							Expect(buildpackGUIDUsed).To(Equal(buildpackGUID))
							Expect(pathUsed).To(Equal(buildpackBitsPath))
						})

						When("uploading the buildpack fails", func() {
							BeforeEach(func() {
								expectedErr = errors.New("upload error")
								fakeActor.UploadBuildpackReturns("", v7action.Warnings{"upload-warning1", "upload-warning2"}, expectedErr)
							})

							It("returns all warnings and an error", func() {
								Expect(testUI.Err).To(Say("update-bp-warning1"))
								Expect(testUI.Err).To(Say("update-bp-warning2"))
								Expect(testUI.Err).To(Say("upload-warning1"))
								Expect(testUI.Err).To(Say("upload-warning2"))
								Expect(executeErr).To(MatchError(expectedErr))
							})
						})

						When("uploading the buildpack succeeds", func() {
							BeforeEach(func() {
								fakeActor.UploadBuildpackReturns(
									"example.com/job/url/",
									v7action.Warnings{"upload-warning1", "upload-warning2"},
									nil,
								)
							})

							When("polling the buildpack job fails", func() {
								BeforeEach(func() {
									expectedErr = ccerror.JobTimeoutError{JobGUID: "job-guid"}
									fakeActor.PollUploadBuildpackJobReturns(
										v7action.Warnings{"poll-warning1", "poll-warning2"},
										expectedErr,
									)
								})

								It("returns all warnings and an error", func() {
									Expect(testUI.Err).To(Say("update-bp-warning1"))
									Expect(testUI.Err).To(Say("update-bp-warning2"))
									Expect(testUI.Err).To(Say("poll-warning1"))
									Expect(testUI.Err).To(Say("poll-warning2"))
									Expect(executeErr).To(MatchError(expectedErr))
								})
							})

							When("polling the buildpack job succeeds", func() {
								BeforeEach(func() {
									fakeActor.PollUploadBuildpackJobReturns(
										v7action.Warnings{"poll-warning1", "poll-warning2"},
										nil,
									)
								})

								It("displays success test and any warnings", func() {
									Expect(testUI.Out).To(Say(`Uploading buildpack %s`, buildpackName))
									Expect(testUI.Err).To(Say("upload-warning1"))
									Expect(testUI.Err).To(Say("upload-warning2"))
									Expect(testUI.Out).To(Say("OK"))

									Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, buildpackName))
									Expect(testUI.Err).To(Say("poll-warning1"))
									Expect(testUI.Err).To(Say("poll-warning2"))
									Expect(testUI.Out).To(Say("OK"))
								})
							})
						})
					})
				})
			})
		})
	})
})
