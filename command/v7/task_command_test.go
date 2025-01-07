package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	. "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("task Command", func() {
	var (
		cmd             TaskCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = TaskCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.AppName = "some-app-name"
		cmd.RequiredArgs.TaskID = 3

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
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
				fakeActor.GetCurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("getting the current user does not return an error", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			When("provided a valid application name", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{GUID: "some-app-guid"},
						v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
						nil)
					fakeActor.GetTaskBySequenceIDAndApplicationReturns(
						resources.Task{
							GUID:              "task-3-guid",
							SequenceID:        3,
							Name:              "task-3",
							State:             constant.TaskRunning,
							CreatedAt:         "2016-11-08T22:26:02Z",
							Command:           "some-command",
							MemoryInMB:        100,
							DiskInMB:          200,
							LogRateLimitInBPS: 300,
							Result: &resources.TaskResult{
								FailureReason: "some failure message",
							},
						},
						v7action.Warnings{"get-task-warning-1"},
						nil)
				})

				It("outputs the task and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app-name"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(fakeActor.GetTaskBySequenceIDAndApplicationCallCount()).To(Equal(1))
					taskId, appGuid := fakeActor.GetTaskBySequenceIDAndApplicationArgsForCall(0)
					Expect(taskId).To(Equal(3))
					Expect(appGuid).To(Equal("some-app-guid"))

					Expect(testUI.Out).To(Say("Getting task 3 for app some-app-name in org some-org / space some-space as some-user..."))

					Expect(testUI.Out).To(Say(`id:\s+3`))
					Expect(testUI.Out).To(Say(`name:\s+task-3`))
					Expect(testUI.Out).To(Say(`state:\s+RUNNING`))
					Expect(testUI.Out).To(Say(`start time:\s+2016-11-08T22:26:02Z`))
					Expect(testUI.Out).To(Say(`command:\s+some-command`))
					Expect(testUI.Out).To(Say(`memory in mb:\s+100`))
					Expect(testUI.Out).To(Say(`disk in mb:\s+200`))
					Expect(testUI.Out).To(Say(`log rate limit:\s+300`))
					Expect(testUI.Out).To(Say(`failure reason:\s+some failure message`))

					Expect(testUI.Err).To(Say("get-application-warning-1"))
					Expect(testUI.Err).To(Say("get-application-warning-2"))
					Expect(testUI.Err).To(Say("get-task-warning-1"))
				})

				When("the API does not return a command", func() {
					BeforeEach(func() {
						fakeActor.GetTaskBySequenceIDAndApplicationReturns(
							resources.Task{
								GUID:              "task-3-guid",
								SequenceID:        3,
								Name:              "task-3",
								State:             constant.TaskRunning,
								CreatedAt:         "2016-11-08T22:26:02Z",
								Command:           "",
								MemoryInMB:        100,
								DiskInMB:          200,
								LogRateLimitInBPS: 300,
								Result: &resources.TaskResult{
									FailureReason: "some failure message",
								},
							},
							v7action.Warnings{"get-task-warning-1"},
							nil)
					})
					It("displays [hidden] for the command", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say(`.*command:\s+\[hidden\]`))
					})
				})
			})

			When("there are errors", func() {
				When("the error is translatable", func() {
					When("getting the application returns the error", func() {
						var (
							returnedErr error
							expectedErr error
						)

						BeforeEach(func() {
							expectedErr = errors.New("request-error")
							returnedErr = ccerror.RequestError{Err: expectedErr}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								resources.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})

					When("getting the app task returns the error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = ccerror.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								resources.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								resources.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(returnedErr))
						})
					})
				})

				When("the error is not translatable", func() {
					When("getting the app returns the error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								resources.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								expectedErr)
						})

						It("return the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					When("getting the app task returns the error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								resources.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								nil)
							fakeActor.GetTaskBySequenceIDAndApplicationReturns(
								resources.Task{},
								v7action.Warnings{"get-task-warning-1", "get-task-warning-2"},
								expectedErr)
						})

						It("returns the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
							Expect(testUI.Err).To(Say("get-task-warning-1"))
							Expect(testUI.Err).To(Say("get-task-warning-2"))
						})
					})
				})
			})
		})
	})
})
