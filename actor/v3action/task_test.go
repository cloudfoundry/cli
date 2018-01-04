package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("RunTask", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationTaskReturns(
					ccv3.Task{
						SequenceID: 3,
					},
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("creates and returns the task and all warnings", func() {
				expectedTask := Task{
					Command:    "some command",
					Name:       "some-task-name",
					MemoryInMB: 123,
					DiskInMB:   321,
				}
				task, warnings, err := actor.RunTask("some-app-guid", expectedTask)
				Expect(err).ToNot(HaveOccurred())

				Expect(task).To(Equal(Task{
					SequenceID: 3,
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.CreateApplicationTaskCallCount()).To(Equal(1))
				appGUIDArg, taskArg := fakeCloudControllerClient.CreateApplicationTaskArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(taskArg).To(Equal(ccv3.Task(expectedTask)))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var warnings Warnings
			var err error
			var expectedErr error

			JustBeforeEach(func() {
				_, warnings, err = actor.RunTask("some-app-guid", Task{Command: "some command"})
			})

			Context("when the cloud controller error is generic", func() {
				BeforeEach(func() {
					expectedErr = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.CreateApplicationTaskReturns(
						ccv3.Task{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("returns the same error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when the error is a TaskWorkersUnavailableError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateApplicationTaskReturns(
						ccv3.Task{},
						ccv3.Warnings{"warning-1", "warning-2"},
						ccerror.TaskWorkersUnavailableError{Message: "banana babans"},
					)
				})

				It("returns a TaskWorkersUnavailableError and all warnings", func() {
					Expect(err).To(MatchError(actionerror.TaskWorkersUnavailableError{Message: "banana babans"}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})
	})

	Describe("GetApplicationTasks", func() {
		Context("when the application exists", func() {
			Context("when there are associated tasks", func() {
				var (
					task1 ccv3.Task
					task2 ccv3.Task
					task3 ccv3.Task
				)

				BeforeEach(func() {
					task1 = ccv3.Task{
						GUID:       "task-1-guid",
						SequenceID: 1,
						Name:       "task-1",
						State:      constant.TaskSucceeded,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task2 = ccv3.Task{
						GUID:       "task-2-guid",
						SequenceID: 2,
						Name:       "task-2",
						State:      constant.TaskFailed,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task3 = ccv3.Task{
						GUID:       "task-3-guid",
						SequenceID: 3,
						Name:       "task-3",
						State:      constant.TaskRunning,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]ccv3.Task{task3, task1, task2},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns all tasks associated with the application and all warnings", func() {
					tasks, warnings, err := actor.GetApplicationTasks("some-app-guid", Descending)
					Expect(err).ToNot(HaveOccurred())

					Expect(tasks).To(Equal([]Task{Task(task3), Task(task2), Task(task1)}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					tasks, warnings, err = actor.GetApplicationTasks("some-app-guid", Ascending)
					Expect(err).ToNot(HaveOccurred())

					Expect(tasks).To(Equal([]Task{Task(task1), Task(task2), Task(task3)}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.GetApplicationTasksCallCount()).To(Equal(2))
					appGUID, query := fakeCloudControllerClient.GetApplicationTasksArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(query).To(BeNil())
				})
			})

			Context("when there are no associated tasks", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]ccv3.Task{},
						nil,
						nil,
					)
				})

				It("returns an empty list of tasks", func() {
					tasks, _, err := actor.GetApplicationTasks("some-app-guid", Descending)
					Expect(err).ToNot(HaveOccurred())
					Expect(tasks).To(BeEmpty())
				})
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationTasksReturns(
					[]ccv3.Task{},
					ccv3.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				_, warnings, err := actor.GetApplicationTasks("some-app-guid", Descending)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetTaskBySequenceIDAndApplication", func() {
		Context("when the cloud controller client does not return an error", func() {
			Context("when the task is found", func() {
				var task1 ccv3.Task

				BeforeEach(func() {
					task1 = ccv3.Task{
						GUID:       "task-1-guid",
						SequenceID: 1,
					}
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]ccv3.Task{task1},
						ccv3.Warnings{"get-task-warning-1"},
						nil,
					)
				})

				It("returns the task and warnings", func() {
					task, warnings, err := actor.GetTaskBySequenceIDAndApplication(1, "some-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(task).To(Equal(Task(task1)))
					Expect(warnings).To(ConsistOf("get-task-warning-1"))
				})
			})

			Context("when the task is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]ccv3.Task{},
						ccv3.Warnings{"get-task-warning-1"},
						nil,
					)
				})

				It("returns a TaskNotFoundError and warnings", func() {
					_, warnings, err := actor.GetTaskBySequenceIDAndApplication(1, "some-app-guid")
					Expect(err).To(MatchError(actionerror.TaskNotFoundError{SequenceID: 1}))
					Expect(warnings).To(ConsistOf("get-task-warning-1"))
				})
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("generic-error")
				fakeCloudControllerClient.GetApplicationTasksReturns(
					[]ccv3.Task{},
					ccv3.Warnings{"get-task-warning-1"},
					expectedErr,
				)
			})

			It("returns the same error and warnings", func() {
				_, warnings, err := actor.GetTaskBySequenceIDAndApplication(1, "some-app-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-task-warning-1"))
			})
		})
	})

	Describe("TerminateTask", func() {
		Context("when the task exists", func() {
			var returnedTask ccv3.Task

			BeforeEach(func() {
				returnedTask = ccv3.Task{
					GUID:       "some-task-guid",
					SequenceID: 1,
				}
				fakeCloudControllerClient.UpdateTaskReturns(
					returnedTask,
					ccv3.Warnings{"update-task-warning"},
					nil)
			})

			It("returns the task and warnings", func() {
				task, warnings, err := actor.TerminateTask("some-task-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("update-task-warning"))
				Expect(task).To(Equal(Task(returnedTask)))
			})
		})

		Context("when the cloud controller returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("cc-error")
				fakeCloudControllerClient.UpdateTaskReturns(
					ccv3.Task{},
					ccv3.Warnings{"update-task-warning"},
					expectedErr)
			})

			It("returns the same error and warnings", func() {
				_, warnings, err := actor.TerminateTask("some-task-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("update-task-warning"))
			})
		})
	})
})
