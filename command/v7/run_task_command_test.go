package v7_test

import (
	"errors"

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("run-task Command", func() {
	var (
		cmd             RunTaskCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeRunTaskActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeRunTaskActor)

		cmd = RunTaskCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app-name"
		cmd.Command = "some command"

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
				expectedErr = errors.New("got bananapants??")
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
				})

				When("the task name is not provided", func() {
					BeforeEach(func() {
						fakeActor.RunTaskReturns(
							v7action.Task{
								Name:       "31337ddd",
								SequenceID: 3,
							},
							v7action.Warnings{"get-application-warning-3"},
							nil)
					})

					It("creates a new task and displays all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app-name"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, task := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(task).To(Equal(v7action.Task{Command: "some command"}))

						Expect(testUI.Out).To(Say("Creating task for app some-app-name in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Out).To(Say("Task has been submitted successfully for execution."))
						Expect(testUI.Out).To(Say(`task name:\s+31337ddd`))
						Expect(testUI.Out).To(Say(`task id:\s+3`))

						Expect(testUI.Err).To(Say("get-application-warning-1"))
						Expect(testUI.Err).To(Say("get-application-warning-2"))
						Expect(testUI.Err).To(Say("get-application-warning-3"))
					})
				})

				When("task disk space is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						cmd.Disk = flag.Megabytes{NullUint64: types.NullUint64{Value: 321, IsSet: true}}
						cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 123, IsSet: true}}
						fakeActor.RunTaskReturns(
							v7action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v7action.Warnings{"get-application-warning-3"},
							nil)
					})

					It("creates a new task and outputs all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app-name"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, task := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(task).To(Equal(v7action.Task{
							Command:    "some command",
							Name:       "some-task-name",
							DiskInMB:   321,
							MemoryInMB: 123,
						}))

						Expect(testUI.Out).To(Say("Creating task for app some-app-name in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Out).To(Say("Task has been submitted successfully for execution."))
						Expect(testUI.Out).To(Say(`task name:\s+some-task-name`))
						Expect(testUI.Out).To(Say(`task id:\s+3`))

						Expect(testUI.Err).To(Say("get-application-warning-1"))
						Expect(testUI.Err).To(Say("get-application-warning-2"))
						Expect(testUI.Err).To(Say("get-application-warning-3"))
					})
				})

				When("process is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						cmd.Process = "process-template-name"
						cmd.Command = ""
						fakeActor.RunTaskReturns(
							v7action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v7action.Warnings{"get-application-warning-3"},
							nil)
						fakeActor.GetProcessByTypeAndApplicationReturns(
							v7action.Process{
								GUID: "some-process-guid",
							},
							v7action.Warnings{"get-process-warning"},
							nil)
					})

					It("creates a new task and outputs all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, task := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(task).To(Equal(v7action.Task{
							Command: "",
							Name:    "some-task-name",
							Template: &ccv3.TaskTemplate{
								Process: ccv3.TaskProcessTemplate{Guid: "some-process-guid"},
							},
						}))

						Expect(testUI.Out).To(Say("Creating task for app some-app-name in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Out).To(Say("Task has been submitted successfully for execution."))
						Expect(testUI.Out).To(Say(`task name:\s+some-task-name`))
						Expect(testUI.Out).To(Say(`task id:\s+3`))

						Expect(testUI.Err).To(Say("get-application-warning-1"))
						Expect(testUI.Err).To(Say("get-application-warning-2"))
						Expect(testUI.Err).To(Say("get-process-warning"))
						Expect(testUI.Err).To(Say("get-application-warning-3"))
					})
				})

				When("neither command nor process template are provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						cmd.Command = ""
						fakeActor.RunTaskReturns(
							v7action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v7action.Warnings{"get-application-warning-3"},
							nil)
						fakeActor.GetProcessByTypeAndApplicationReturns(
							v7action.Process{
								GUID: "some-process-guid",
							},
							v7action.Warnings{"get-process-warning"},
							nil)
					})

					It("creates a new task using 'task' as the template process type and outputs all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetProcessByTypeAndApplicationCallCount()).To(Equal(1))
						processType, _ := fakeActor.GetProcessByTypeAndApplicationArgsForCall(0)
						Expect(processType).To(Equal("task"))

						Expect(fakeActor.RunTaskCallCount()).To(Equal(1))
						appGUID, task := fakeActor.RunTaskArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(task).To(Equal(v7action.Task{
							Command: "",
							Name:    "some-task-name",
							Template: &ccv3.TaskTemplate{
								Process: ccv3.TaskProcessTemplate{Guid: "some-process-guid"},
							},
						}))

						Expect(testUI.Out).To(Say("Creating task for app some-app-name in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Out).To(Say("Task has been submitted successfully for execution."))
						Expect(testUI.Out).To(Say(`task name:\s+some-task-name`))
						Expect(testUI.Out).To(Say(`task id:\s+3`))

						Expect(testUI.Err).To(Say("get-application-warning-1"))
						Expect(testUI.Err).To(Say("get-application-warning-2"))
						Expect(testUI.Err).To(Say("get-process-warning"))
						Expect(testUI.Err).To(Say("get-application-warning-3"))
					})
				})
			})

			When("there are errors", func() {
				When("the error is translatable", func() {
					When("getting the app returns the error", func() {
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

					When("running the task returns the error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = ccerror.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.RunTaskReturns(
								v7action.Task{},
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
							expectedErr = errors.New("got bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								expectedErr)
						})

						It("return the error and all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					When("running the task returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("got bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v7action.Application{GUID: "some-app-guid"},
								v7action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								nil)
							fakeActor.RunTaskReturns(
								v7action.Task{},
								v7action.Warnings{"run-task-warning-1", "run-task-warning-2"},
								expectedErr)
						})

						It("returns the error and all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
							Expect(testUI.Err).To(Say("run-task-warning-1"))
							Expect(testUI.Err).To(Say("run-task-warning-2"))
						})
					})
				})
			})
		})
	})
})
