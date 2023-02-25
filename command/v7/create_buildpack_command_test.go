package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create buildpack Command", func() {
	var (
		cmd             CreateBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
		binaryName      string
		buildpackName   string
		buildpackPath   string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		args = nil
		buildpackName = "some-buildpack"
		buildpackPath = "/path/to/buildpack.zip"

		cmd = CreateBuildpackCommand{
			RequiredArgs: flag.CreateBuildpackArgs{
				Buildpack: buildpackName,
				Path:      flag.PathWithExistenceCheckOrURL(buildpackPath),
				Position:  7,
			},
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("should print text indicating it is creating a buildpack", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating buildpack %s as the-user\.\.\.`, buildpackName))
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
						resources.Buildpack{},
						v7action.Warnings{"warning-1"},
						actionerror.BuildpackNameTakenError{Name: "this-error-occurred"},
					)
				})
				It("errors and prints all warnings", func() {
					Expect(executeErr).To(Equal(actionerror.BuildpackNameTakenError{Name: "this-error-occurred"}))
					Expect(testUI.Err).To(Say("warning-1"))
				})
			})

			When("The disabled flag is set", func() {
				BeforeEach(func() {
					cmd.Disable = true
					buildpack := resources.Buildpack{
						Name:    buildpackName,
						Enabled: types.NullBool{Value: false, IsSet: true},
					}
					fakeActor.CreateBuildpackReturns(buildpack, v7action.Warnings{"some-create-warning-1"}, nil)
				})

				It("correctly creates a disabled buildpack", func() {
					buildpack := fakeActor.CreateBuildpackArgsForCall(0)
					Expect(buildpack.Name).To(Equal(buildpackName))
					Expect(buildpack.Enabled.Value).To(BeFalse())
				})
			})

			When("creating buildpack succeeds", func() {
				BeforeEach(func() {
					buildpack := resources.Buildpack{
						Name:     buildpackName,
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
					Expect(buildpack.Name).To(Equal(buildpackName))
					Expect(buildpack.Position.Value).To(Equal(7))
				})

				It("prints any warnings and uploads the bits", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Err).To(Say("some-create-warning-1"))
					Expect(testUI.Out).To(Say(`Uploading buildpack %s as the-user\.\.\.`, buildpackName))
				})

				It("Displays it is starting the upload", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Uploading buildpack %s as the-user", buildpackName))

					Expect(fakeActor.PrepareBuildpackBitsCallCount()).To(Equal(1))
					path, _, _ := fakeActor.PrepareBuildpackBitsArgsForCall(0)
					Expect(path).To(Equal(buildpackPath))
				})

				When("uploading the buildpack fails due to an auth token expired error", func() {
					BeforeEach(func() {
						fakeActor.UploadBuildpackReturns(
							ccv3.JobURL(""),
							v7action.Warnings{"some-create-bp-with-auth-warning"},
							ccerror.InvalidAuthTokenError{Message: "token expired"},
						)
					})

					It("alerts the user and retries the upload", func() {
						Expect(testUI.Err).To(Say("Failed to upload buildpack due to auth token expiration, retrying..."))
						Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(2))
					})
				})

				When("Uploading the buildpack fails due to a generic error", func() {
					BeforeEach(func() {
						fakeActor.UploadBuildpackReturns(
							ccv3.JobURL(""),
							v7action.Warnings{"warning-2"},
							errors.New("some-error"),
						)
					})

					It("errors, prints a tip and all warnings", func() {
						Expect(executeErr).To(MatchError(translatableerror.TipDecoratorError{
							BaseError: errors.New("some-error"),
							Tip:       "A buildpack with name '{{.BuildpackName}}' and nil stack has been created. Use '{{.CfDeleteBuildpackCommand}}' to delete it or '{{.CfUpdateBuildpackCommand}}' to try again.",
							TipKeys: map[string]interface{}{
								"BuildpackName":            cmd.RequiredArgs.Buildpack,
								"CfDeleteBuildpackCommand": cmd.Config.BinaryName() + " delete-buildpack",
								"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack",
							},
						}))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Out).To(Say("Uploading buildpack %s", buildpackName))
						Consistently(testUI.Out).ShouldNot(Say("OK"))
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
						Expect(testUI.Out).To(Say("Uploading buildpack %s", buildpackName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("some-upload-warning-1"))

						Expect(fakeActor.UploadBuildpackCallCount()).To(Equal(1))
						guid, path, _ := fakeActor.UploadBuildpackArgsForCall(0)
						Expect(guid).To(Equal("some-guid"))
						Expect(path).To(Equal("buildpack.zip"))
					})

					Describe("polling the upload-to-blobstore job", func() {
						It("polls for job completion/failure", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(testUI.Out).To(Say("Uploading buildpack %s", buildpackName))
							Expect(testUI.Out).To(Say("OK"))

							Expect(fakeActor.PollUploadBuildpackJobCallCount()).To(Equal(1))
							url := fakeActor.PollUploadBuildpackJobArgsForCall(0)

							Expect(url).To(Equal(ccv3.JobURL("http://example.com/some-job-url")))
						})

						When("the job completes successfully", func() {
							BeforeEach(func() {
								fakeActor.PollUploadBuildpackJobReturns(v7action.Warnings{"poll-warning"}, nil)
							})

							It("prints all warnings and exits successfully", func() {
								Expect(executeErr).NotTo(HaveOccurred())
								Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, buildpackName))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Err).To(Say("poll-warning"))
							})
						})

						When("the job fails with an error", func() {
							BeforeEach(func() {
								fakeActor.PollUploadBuildpackJobReturns(
									v7action.Warnings{"poll-warning"},
									errors.New("some-error"),
								)
							})

							It("prints all warnings and a tip, then returns the error", func() {
								Expect(executeErr).To(MatchError(translatableerror.TipDecoratorError{
									BaseError: errors.New("some-error"),
									Tip:       "A buildpack with name '{{.BuildpackName}}' and nil stack has been created. Use '{{.CfDeleteBuildpackCommand}}' to delete it or '{{.CfUpdateBuildpackCommand}}' to try again.",
									TipKeys: map[string]interface{}{
										"BuildpackName":            cmd.RequiredArgs.Buildpack,
										"CfDeleteBuildpackCommand": cmd.Config.BinaryName() + " delete-buildpack",
										"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack",
									},
								}))
								Expect(testUI.Err).To(Say("poll-warning"))
								Expect(testUI.Out).To(Say(`Processing uploaded buildpack %s\.\.\.`, buildpackName))
								Consistently(testUI.Out).ShouldNot(Say("OK"))
							})
						})
					})
				})
			})
		})
	})
})
