package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Terminate Task Command", func() {
	var (
		cmd        v3.TerminateTaskCommand
		testUI     *ui.UI
		fakeActor  *v3fakes.FakeTerminateTaskActor
		fakeConfig *commandfakes.FakeConfig
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v3fakes.FakeTerminateTaskActor)
		fakeActor.CloudControllerAPIVersionReturns("3.0.0")
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v3.TerminateTaskCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}

		cmd.RequiredArgs.AppName = "some-app-name"
		cmd.RequiredArgs.SequenceID = "1"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is too small", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionError", func() {
			Expect(executeErr).To(MatchError(command.MinimumAPIVersionError{
				CurrentVersion: "0.0.0",
				MinimumVersion: "3.0.0",
			}))
		})
	})

	Context("when the task id argument is not an integer", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SequenceID = "not-an-integer"
		})

		It("returns an ParseArgumentError", func() {
			Expect(executeErr).To(MatchError(command.ParseArgumentError{
				ArgumentName: "TASK_ID",
				ExpectedType: "integer",
			}))
		})
	})

	Context("when the user is not logged in", func() {
		It("returns a NotLoggedInError", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{}))
		})
	})

	Context("when an organization is not targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
		})

		It("returns a NoTargetedOrgError", func() {
			Expect(executeErr).To(MatchError(command.NoTargetedOrgError{}))
		})
	})

	Context("when a space is not targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
		})

		It("returns a NoTargetedSpaceError", func() {
			Expect(executeErr).To(MatchError(command.NoTargetedSpaceError{}))
		})
	})

	Context("when the user is logged in, and a space and an org is targetted", func() {
		BeforeEach(func() {
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		Context("when getting the logged in user results in an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get current user error")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the same error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when getting the logged in user does not result in an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "some-user",
				}, nil)
			})

			Context("when provided a valid application name and task sequence ID", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{GUID: "some-app-guid"},
						v3action.Warnings{"get-application-warning"},
						nil,
					)
					fakeActor.GetTaskBySequenceIDAndApplicationReturns(
						v3action.Task{GUID: "some-task-guid"},
						v3action.Warnings{"get-task-warning"},
						nil,
					)
					fakeActor.TerminateTaskReturns(
						v3action.Task{},
						v3action.Warnings{"terminate-task-warning"},
						nil,
					)
				})

				It("cancels the task", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app-name"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(fakeActor.GetTaskBySequenceIDAndApplicationCallCount()).To(Equal(1))
					sequenceID, applicationGUID := fakeActor.GetTaskBySequenceIDAndApplicationArgsForCall(0)
					Expect(sequenceID).To(Equal(1))
					Expect(applicationGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.TerminateTaskCallCount()).To(Equal(1))
					taskGUID := fakeActor.TerminateTaskArgsForCall(0)
					Expect(taskGUID).To(Equal("some-task-guid"))

					Expect(testUI.Err).To(Say("get-application-warning"))
					Expect(testUI.Err).To(Say("get-task-warning"))
					Expect(testUI.Out).To(Say("Terminating task 1 of app some-app-name in org some-org / space some-space as some-user..."))
					Expect(testUI.Err).To(Say("terminate-task-warning"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			Context("when there are errors", func() {
				Context("when a translated error is returned", func() {
					var (
						returnedErr error
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = errors.New("request-error")
						returnedErr = cloudcontroller.RequestError{Err: expectedErr}
					})

					Context("when GetApplicationByNameAndSpace returns a translatable error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(command.APIRequestError{Err: expectedErr}))
						})
					})

					Context("when GetTaskBySequenceIDAndApplication returns a translatable error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v3action.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(command.APIRequestError{Err: expectedErr}))
						})
					})

					Context("when TerminateTask returns a translatable error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v3action.Task{GUID: "some-task-guid"},
								nil,
								nil)
							fakeActor.TerminateTaskReturns(
								v3action.Task{GUID: "some-task-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(command.APIRequestError{Err: expectedErr}))
						})
					})
				})

				Context("when an untranslatable error is returned", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("bananapants")
					})

					Context("when GetApplicationByNameAndSpace returns an error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{
									"get-application-warning-1",
									"get-application-warning-2",
								}, expectedErr)
						})

						It("return the same error and outputs the warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					Context("when GetTaskBySequenceIDAndApplication returns an error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(v3action.Task{},
								v3action.Warnings{
									"get-task-warning-1",
									"get-task-warning-2",
								}, expectedErr)
						})

						It("return the same error and outputs the warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-task-warning-1"))
							Expect(testUI.Err).To(Say("get-task-warning-2"))
						})
					})

					Context("when TerminateTask returns an error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(v3action.Task{GUID: "some-task-guid"},
								nil,
								nil)
							fakeActor.TerminateTaskReturns(
								v3action.Task{},
								v3action.Warnings{
									"terminate-task-warning-1",
									"terminate-task-warning-2",
								}, expectedErr)
						})

						It("returns the same error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("terminate-task-warning-1"))
							Expect(testUI.Err).To(Say("terminate-task-warning-2"))
						})
					})
				})
			})
		})
	})
})
