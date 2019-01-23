package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"

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

		When("preparing the buildpack bits fails", func() {
			BeforeEach(func() {
				fakeActor.PrepareBuildpackBitsReturns("some/invalid/path", errors.New("some-prepare-bp-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("some-prepare-bp-error"))
				Expect(fakeActor.PrepareBuildpackBitsCallCount()).To(Equal(1))
			})
		})

		When("Preparing the buildpack bits succeeds", func() {

			BeforeEach(func() {
				fakeActor.PrepareBuildpackBitsReturns("buildpack.zip", nil)
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

			When("The disabled flag is set", func() {
				BeforeEach(func() {
					cmd.Disable = true
					buildpack := v7action.Buildpack{
						Name:    "buildpack-1",
						Enabled: types.NullBool{Value: false, IsSet: true},
					}
					fakeActor.CreateBuildpackReturns(buildpack, v7action.Warnings{"some-create-warning-1"}, nil)
				})

				It("correctly creates a disabled buildpack", func() {
					buildpack := fakeActor.CreateBuildpackArgsForCall(0)
					Expect(buildpack.Name).To(Equal("some-buildpack"))
					Expect(buildpack.Enabled.Value).To(BeFalse())
				})
			})

			When("creating buildpack succeeds", func() {
				BeforeEach(func() {
					buildpack := v7action.Buildpack{
						Name:     "buildpack-1",
						Position: types.NullInt{Value: 1, IsSet: true},
						Enabled:  types.NullBool{Value: true, IsSet: true},
						Locked:   types.NullBool{Value: false, IsSet: true},
						Filename: "buildpack-1.file",
						Stack:    "buildpack-1-stack",
						GUID:     "some-guid",
					}
					fakeActor.CreateBuildpackReturns(buildpack, v7action.Warnings{"some-create-warning-1"}, nil)
				})

				It("correctly created the buildpack", func() {
					buildpack := fakeActor.CreateBuildpackArgsForCall(0)
					Expect(buildpack.Name).To(Equal("some-buildpack"))
					Expect(buildpack.Position.Value).To(Equal(7))
				})

				It("prints any warnings and uploads the bits", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Err).To(Say("some-create-warning-1"))
					Expect(testUI.Out).To(Say(`Uploading buildpack some-buildpack as the-user\.\.\.`))
				})

				It("Displays it is starting the upload", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Uploading buildpack some-buildpack as the-user"))

					Expect(fakeActor.PrepareBuildpackBitsCallCount()).To(Equal(1))
					path, _, _ := fakeActor.PrepareBuildpackBitsArgsForCall(0)
					Expect(path).To(Equal("/path/to/buildpack.zip"))
				})

				// TODO: ENUMERATE ERRORS HERE
				When("Uploading the buildpack fails due to a name+stack collision", func() {
					BeforeEach(func() {
						fakeActor.UploadBuildpackReturns(
							ccv3.JobURL(""),
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
							ccv3.JobURL("http://example.com/some-job-url"),
							v7action.Warnings{"some-upload-warning-1"},
							nil,
						)
					})

					It("prints all warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(testUI.Out).To(Say("Done uploading"))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("some-upload-warning-1"))

						Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(1))
						guid, path, _ := fakeActor.UploadBuildpackArgsForCall(0)
						Expect(guid).To(Equal("some-guid"))
						Expect(path).To(Equal("buildpack.zip"))
					})

					Describe("polling the upload job", func() {
						It("polls for job completion/failure", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(testUI.Out).To(Say("Done uploading"))
							Expect(testUI.Out).To(Say("OK"))

							Expect(fakeActor.PollUploadBuildpackJobCallCount()).To(Equal(1))
							url := fakeActor.PollUploadBuildpackJobArgsForCall(0)

							Expect(url).To(Equal(ccv3.JobURL("http://example.com/some-job-url")))
						})

						When("polling the upload job returns completed", func() {
							BeforeEach(func() {
								fakeActor.PollUploadBuildpackJobReturns(v7action.Warnings{"poll-warning"}, nil)
							})

							It("prints all warnings and exits successfully", func() {
								Expect(executeErr).NotTo(HaveOccurred())
								Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, "some-buildpack"))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Err).To(Say("poll-warning"))
							})
						})

						When("the polling job returns a ccerror.BuildpackAlreadyExistsForStackError", func() {
							BeforeEach(func() {
								fakeActor.PollUploadBuildpackJobReturns(
									v7action.Warnings{"poll-warning"},
									ccerror.BuildpackAlreadyExistsForStackError{},
								)
							})

							It("prints all warnings, and returns the error", func() {
								Expect(executeErr).To(Equal(ccerror.BuildpackAlreadyExistsForStackError{}))
								Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, "some-buildpack"))
								Expect(testUI.Err).To(Say("poll-warning"))
								Consistently(testUI.Out).ShouldNot(Say("OK"))
							})
						})

						When("the polling job returns a ccerror.BuildpackAlreadyExistsWithoutStackError", func() {
							BeforeEach(func() {
								fakeActor.PollUploadBuildpackJobReturns(
									v7action.Warnings{"poll-warning"},
									ccerror.BuildpackAlreadyExistsWithoutStackError{},
								)
							})

							It("prints all warnings, and returns the error", func() {
								Expect(executeErr).To(Equal(ccerror.BuildpackAlreadyExistsWithoutStackError{}))
								Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, "some-buildpack"))
								Expect(testUI.Err).To(Say("poll-warning"))
								Consistently(testUI.Out).ShouldNot(Say("OK"))
							})
						})

					})
				})
			})
		})
	})
})
