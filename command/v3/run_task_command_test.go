package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("run-task Command", func() {
	var (
		cmd             v3.RunTaskCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeRunTaskActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeRunTaskActor)

		cmd = v3.RunTaskCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app-name"
		cmd.RequiredArgs.Command = "some command"

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
				expectedErr = errors.New("got bananapants??")
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
				})

				Context("when the task name is not provided", func() {
					BeforeEach(func() {
						fakeActor.RunTaskReturns(
							v3action.Task{
								Name:       "31337ddd",
								SequenceID: 3,
							},
							v3action.Warnings{"get-application-warning-3"},
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
						Expect(task).To(Equal(v3action.Task{Command: "some command"}))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task has been submitted successfully for execution.
task name:   31337ddd
task id:     3
`))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
					})
				})

				Context("when the task name is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						fakeActor.RunTaskReturns(
							v3action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v3action.Warnings{"get-application-warning-3"},
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
						Expect(task).To(Equal(v3action.Task{Command: "some command", Name: "some-task-name"}))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task has been submitted successfully for execution.
task name:   some-task-name
task id:     3`,
						))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
					})
				})

				Context("when task disk space is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						cmd.Disk = flag.Megabytes{NullUint64: types.NullUint64{Value: 321, IsSet: true}}
						fakeActor.RunTaskReturns(
							v3action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v3action.Warnings{"get-application-warning-3"},
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
						Expect(task).To(Equal(v3action.Task{
							Command:  "some command",
							Name:     "some-task-name",
							DiskInMB: 321,
						}))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task has been submitted successfully for execution.
task name:   some-task-name
task id:     3`,
						))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
					})
				})

				Context("when task memory is provided", func() {
					BeforeEach(func() {
						cmd.Name = "some-task-name"
						cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 123, IsSet: true}}
						fakeActor.RunTaskReturns(
							v3action.Task{
								Name:       "some-task-name",
								SequenceID: 3,
							},
							v3action.Warnings{"get-application-warning-3"},
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
						Expect(task).To(Equal(v3action.Task{
							Command:    "some command",
							Name:       "some-task-name",
							MemoryInMB: 123,
						}))

						Expect(testUI.Out).To(Say(`Creating task for app some-app-name in org some-org / space some-space as some-user...
OK

Task has been submitted successfully for execution.
task name:   some-task-name
task id:     3`,
						))
						Expect(testUI.Err).To(Say(`get-application-warning-1
get-application-warning-2
get-application-warning-3`))
					})
				})
			})

			Context("when there are errors", func() {
				Context("when the error is translatable", func() {
					Context("when getting the app returns the error", func() {
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

					Context("when running the task returns the error", func() {
						var returnedErr error

						BeforeEach(func() {
							returnedErr = ccerror.UnverifiedServerError{URL: "some-url"}
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								nil,
								nil)
							fakeActor.RunTaskReturns(
								v3action.Task{},
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
							expectedErr = errors.New("got bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								expectedErr)
						})

						It("return the error and all warnings", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Err).To(Say("get-application-warning-1"))
							Expect(testUI.Err).To(Say("get-application-warning-2"))
						})
					})

					Context("when running the task returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("got bananapants??")
							fakeActor.GetApplicationByNameAndSpaceReturns(
								v3action.Application{GUID: "some-app-guid"},
								v3action.Warnings{"get-application-warning-1", "get-application-warning-2"},
								nil)
							fakeActor.RunTaskReturns(
								v3action.Task{},
								v3action.Warnings{"run-task-warning-1", "run-task-warning-2"},
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
