package v3action_test

import (
	"errors"
	"net/url"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("RunTask", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.NewTaskReturns(
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

			Context("when the task name is empty", func() {
				It("creates and returns the task and all warnings", func() {
					task, warnings, err := actor.RunTask("some-app-guid", "some command", "")
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{
						SequenceID: 3,
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.NewTaskCallCount()).To(Equal(1))
					appGUIDArg, commandArg, name := fakeCloudControllerClient.NewTaskArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(commandArg).To(Equal("some command"))
					Expect(name).To(Equal(""))
				})
			})

			Context("when the task name is not empty", func() {
				It("creates and returns the task and all warnings", func() {
					task, warnings, err := actor.RunTask("some-app-guid", "some command", "some-task-name")
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{
						SequenceID: 3,
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.NewTaskCallCount()).To(Equal(1))
					appGUIDArg, commandArg, name := fakeCloudControllerClient.NewTaskArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(commandArg).To(Equal("some command"))
					Expect(name).To(Equal("some-task-name"))
				})
			})
		})

		Context("when the cloud controller client returns an error", func() {
			Context("when the error is a TaskWorkersUnavailableError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.NewTaskReturns(
						ccv3.Task{},
						nil,
						cloudcontroller.TaskWorkersUnavailableError{Message: "banana babans"},
					)
				})

				It("returns a TaskWorkersUnavailableError", func() {
					_, _, err := actor.RunTask("some-app-guid", "some command", "")
					Expect(err).To(MatchError(TaskWorkersUnavailableError{Message: "banana babans"}))
				})
			})

			Context("when the cloud controller error is generic", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.NewTaskReturns(
						ccv3.Task{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("returns the same error and all warnings", func() {
					_, warnings, err := actor.RunTask("some-app-guid", "some command", "")
					Expect(err).To(MatchError(expectedErr))
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
						State:      "SUCCEEDED",
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task2 = ccv3.Task{
						GUID:       "task-2-guid",
						SequenceID: 2,
						Name:       "task-2",
						State:      "FAILED",
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task3 = ccv3.Task{
						GUID:       "task-3-guid",
						SequenceID: 3,
						Name:       "task-3",
						State:      "RUNNING",
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]ccv3.Task{task1, task2, task3},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns all tasks associated with the application and all warnings", func() {
					tasks, warnings, err := actor.GetApplicationTasks("some-app-guid", Descending)
					Expect(err).ToNot(HaveOccurred())

					Expect(tasks).To(ConsistOf(Task(task1), Task(task2), Task(task3)))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.GetApplicationTasksCallCount()).To(Equal(1))
					appGUID, query := fakeCloudControllerClient.GetApplicationTasksArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(query).To(Equal(
						url.Values{
							"order_by": []string{"-created_at"},
						},
					))
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
					Expect(err).To(MatchError(TaskNotFoundError{SequenceID: 1}))
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
