package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("terminate-task Command", func() {
	var (
		cmd             v7.TerminateTaskCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeTerminateTaskActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeTerminateTaskActor)

		cmd = v7.TerminateTaskCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app-name"
		cmd.RequiredArgs.SequenceID = "1"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the task id argument is not an integer", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.SequenceID = "not-an-integer"
		})

		It("returns an ParseArgumentError", func() {
			Expect(executeErr).To(MatchError(translatableerror.ParseArgumentError{
				ArgumentName: "TASK_ID",
				ExpectedType: "integer",
			}))
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in, and a space and org are targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		When("getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("getting the current user does not return an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("provided a valid application name and task sequence ID", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v7action.Application{GUID: "some-app-guid"},
						v7action.Warnings{"get-application-warning"},
						nil)
					fakeActor.GetTaskBySequenceIDAndApplicationReturns(
						v7action.Task{GUID: "some-task-guid"},
						v7action.Warnings{"get-task-warning"},
						nil)
					fakeActor.TerminateTaskReturns(
						v7action.Task{},
						v7action.Warnings{"terminate-task-warning"},
						nil)
				})

				It("cancels the task and displays all warnings", func() {
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

			When("there are errors", func() {
				When("the error is translatable", func() {
					var (
						returnedErr error
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = errors.New("request-error")
						returnedErr = ccerror.RequestError{Err: expectedErr}
					})

					When("getting the app returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})

					When("getting the task returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v7action.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})

					When("terminating the task returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v7action.Task{GUID: "some-task-guid"},
								nil,
								nil)
							fakeActor.TerminateTaskReturns(
								v7action.Task{GUID: "some-task-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})
				})

				When("the error is not translatable", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("bananapants")
					})

					When("getting the app returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								expectedErr)
						})

						It("return the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					When("getting the task returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v7action.Task{},
								v7action.Warnings{"get-task-warning-1", "get-task-warning-2"},
								expectedErr)
						})

						It("return the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-task-warning-1"))
							Expect(testUI.Err).To(Say("get-task-warning-2"))
						})
					})

					When("terminating the task returns the error", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								v7action.Task{GUID: "some-task-guid"},
								nil,
								nil)
							fakeActor.TerminateTaskReturns(
								v7action.Task{},
								v7action.Warnings{"terminate-task-warning-1", "terminate-task-warning-2"},
								expectedErr)
						})

						It("returns the error and outputs all warnings", func() {
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
