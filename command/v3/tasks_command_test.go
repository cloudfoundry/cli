package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("tasks Command", func() {
	var (
		cmd             v3.TasksCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeTasksActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeTasksActor)

		cmd = v3.TasksCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app-name"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionRunTaskV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionRunTaskV3,
			}))
		})
	})

	Context("when checking target fails", func() {
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

	Context("when the user is logged in, and a space and org are targeted", func() {
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

		Context("when getting the current user returns an error", func() {
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

		Context("when getting the current user does not return an error", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			Context("when provided a valid application name", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{GUID: "some-app-guid"},
						v3action.Warnings{"get-application-warning-1", "get-application-warning-2"},
						nil)
					fakeActor.GetApplicationTasksReturns(
						[]v3action.Task{
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
						v3action.Warnings{"get-tasks-warning-1"},
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
					Expect(order).To(Equal(v3action.Descending))

					Expect(testUI.Out).To(Say(`Getting tasks for app some-app-name in org some-org / space some-space as some-user...
OK

id   name     state       start time                      command
3    task-3   RUNNING     Tue, 08 Nov 2016 22:26:02 UTC   some-command
2    task-2   FAILED      Tue, 08 Nov 2016 22:26:02 UTC   some-command
1    task-1   SUCCEEDED   Tue, 08 Nov 2016 22:26:02 UTC   some-command`,
					))
					Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-tasks-warning-1`))
				})

				Context("when the tasks' command fields are returned as empty strings", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns(
							[]v3action.Task{
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
							v3action.Warnings{"get-tasks-warning-1"},
							nil)
					})

					It("outputs [hidden] for the tasks' commands", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`
2    task-2   FAILED      Tue, 08 Nov 2016 22:26:02 UTC   \[hidden\]
1    task-1   SUCCEEDED   Tue, 08 Nov 2016 22:26:02 UTC   \[hidden\]`,
						))
					})
				})

				Context("when there are no tasks associated with the application", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationTasksReturns([]v3action.Task{}, nil, nil)
					})

					It("outputs an empty table", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`
id   name   state   start time   command
`,
						))
						Expect(testUI.Out).NotTo(Say("1"))
					})
				})
			})

			Context("when there are errors", func() {
				Context("when the error is translatable", func() {
					Context("when getting the application returns the error", func() {
						var (
							returnedErr error
							expectedErr error
						)

						BeforeEach(func() {
							expectedErr = errors.New("request-error")
							returnedErr = ccerror.RequestError{Err: expectedErr}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(ccerror.RequestError{Err: expectedErr}))
						})
					})

					Context("when getting the app's tasks returns the error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = ccerror.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.GetApplicationTasksReturns(
								[]v3action.Task{},
								nil,
								returnedErr)
						})

						It("returns a translatable error", func() {
							Expect(executeErr).To(MatchError(returnedErr))
						})
					})
				})

				Context("when the error is not translatable", func() {
					Context("when getting the app returns the error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								expectedErr)
						})

						It("return the error and outputs all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					Context("when getting the app's tasks returns the error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								nil)
							fakeActor.GetApplicationTasksReturns(
								nil,
								v3action.Warnings{"get-tasks-warning-1", "get-tasks-warning-2"},
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
