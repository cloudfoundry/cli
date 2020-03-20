package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("tasks Command", func() {
	var (
		cmd             TasksCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeTasksActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeTasksActor)

		cmd = TasksCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app-name"

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

			When("provided a valid application name", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v7action.Application{GUID: "some-app-guid"},
						v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
						nil)
					fakeActor.GetApplicationTasksReturns(
						[]v7action.Task{
							{
								GUID:       "task-3-guid",
								SequenceID: 3,
								Name:       "task-3",
								State:      constant.TaskRunning,
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
							{
								GUID:       "task-2-guid",
								SequenceID: 2,
								Name:       "task-2",
								State:      constant.TaskFailed,
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
							{
								GUID:       "task-1-guid",
								SequenceID: 1,
								Name:       "task-1",
								State:      constant.TaskSucceeded,
								CreatedAt:  "2016-11-08T22:26:02Z",
								Command:    "some-command",
							},
						},
						v7action.Warnings{"get-tasks-warning-1"},
						nil)
				})

				It("outputs all tasks associated with the application and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app-name"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(fakeActor.GetApplicationTasksCallCount()).To(Equal(1))
					guid, order := fakeActor.GetApplicationTasksArgsForCall(0)
					Expect(guid).To(Equal("some-app-guid"))
					Expect(order).To(Equal(v7action.Descending))

					Expect(testUI.Out).To(Say("Getting tasks for app some-app-name in org some-org / space some-space as some-user..."))

					Expect(testUI.Out).To(Say(`id\s+name\s+state\s+start time\s+command`))
					Expect(testUI.Out).To(Say(`3\s+task-3\s+RUNNING\s+Tue, 08 Nov 2016 22:26:02 UTC\s+some-command`))
					Expect(testUI.Out).To(Say(`2\s+task-2\s+FAILED\s+Tue, 08 Nov 2016 22:26:02 UTC\s+some-command`))
					Expect(testUI.Out).To(Say(`1\s+task-1\s+SUCCEEDED\s+Tue, 08 Nov 2016 22:26:02 UTC\s+some-command`))
					Expect(testUI.Err).To(Say("get-application-warning-1"))
					Expect(testUI.Err).To(Say("get-application-warning-2"))
					Expect(testUI.Err).To(Say("get-tasks-warning-1"))
				})

				When("the tasks' command fields are returned as empty strings", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns(
							[]v7action.Task{
								{
									GUID:       "task-2-guid",
									SequenceID: 2,
									Name:       "task-2",
									State:      constant.TaskFailed,
									CreatedAt:  "2016-11-08T22:26:02Z",
									Command:    "",
								},
								{
									GUID:       "task-1-guid",
									SequenceID: 1,
									Name:       "task-1",
									State:      constant.TaskSucceeded,
									CreatedAt:  "2016-11-08T22:26:02Z",
									Command:    "",
								},
							},
							v7action.Warnings{"get-tasks-warning-1"},
							nil)
					})

					It("outputs [hidden] for the tasks' commands", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`2\s+task-2\s+FAILED\s+Tue, 08 Nov 2016 22:26:02 UTC\s+\[hidden\]`))
						Expect(testUI.Out).To(Say(`1\s+task-1\s+SUCCEEDED\s+Tue, 08 Nov 2016 22:26:02 UTC\s+\[hidden\]`))
					})
				})

				When("there are no tasks associated with the application", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns([]v7action.Task{}, nil, nil)
					})

					It("outputs an empty table", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`No tasks found for application.`))
						Expect(testUI.Out).NotTo(Say("1"))
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
								v7action.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})

					When("getting the app's tasks returns the error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = ccerror.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetApplicationTasksReturns(
								[]v7action.Task{},
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

					When("getting the app's tasks returns the error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								nil)
							fakeActor.GetApplicationTasksReturns(
								nil,
								v7action.Warnings{"get-tasks-warning-1", "get-tasks-warning-2"},
								expectedErr)
						})

						It("returns the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
							Expect(testUI.Err).To(Say("get-tasks-warning-1"))
							Expect(testUI.Err).To(Say("get-tasks-warning-2"))
						})
					})
				})
			})
		})
	})
})
