package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("RunTask", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationTaskReturns(
					resources.Task{
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
				expectedTask := resources.Task{
					Command:    "some command",
					Name:       "some-task-name",
					MemoryInMB: 123,
					DiskInMB:   321,
				}
				task, warnings, err := actor.RunTask("some-app-guid", expectedTask)
				Expect(err).ToNot(HaveOccurred())

				Expect(task).To(Equal(resources.Task{
					SequenceID: 3,
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.CreateApplicationTaskCallCount()).To(Equal(1))
				appGUIDArg, taskArg := fakeCloudControllerClient.CreateApplicationTaskArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(taskArg).To(Equal(resources.Task(expectedTask)))
			})
		})

		When("the cloud controller client returns an error", func() {
			var warnings Warnings
			var err error
			var expectedErr error

			JustBeforeEach(func() {
				_, warnings, err = actor.RunTask("some-app-guid", resources.Task{Command: "some command"})
			})

			When("the cloud controller error is generic", func() {
				BeforeEach(func() {
					expectedErr = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.CreateApplicationTaskReturns(
						resources.Task{},
						ccv3.Warnings{"warning-1", "warning-2"},
						expectedErr,
					)
				})

				It("returns the same error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("the error is a TaskWorkersUnavailableError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateApplicationTaskReturns(
						resources.Task{},
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
		When("the application exists", func() {
			When("there are associated tasks", func() {
				var (
					task1 resources.Task
					task2 resources.Task
					task3 resources.Task
				)

				BeforeEach(func() {
					task1 = resources.Task{
						GUID:       "task-1-guid",
						SequenceID: 1,
						Name:       "task-1",
						State:      constant.TaskSucceeded,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task2 = resources.Task{
						GUID:       "task-2-guid",
						SequenceID: 2,
						Name:       "task-2",
						State:      constant.TaskFailed,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					task3 = resources.Task{
						GUID:       "task-3-guid",
						SequenceID: 3,
						Name:       "task-3",
						State:      constant.TaskRunning,
						CreatedAt:  "some-time",
						Command:    "some-command",
					}
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]resources.Task{task3, task1, task2},
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns all tasks associated with the application and all warnings", func() {
					tasks, warnings, err := actor.GetApplicationTasks("some-app-guid", Descending)
					Expect(err).ToNot(HaveOccurred())

					Expect(tasks).To(Equal([]resources.Task{resources.Task(task3), resources.Task(task2), resources.Task(task1)}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					tasks, warnings, err = actor.GetApplicationTasks("some-app-guid", Ascending)
					Expect(err).ToNot(HaveOccurred())

					Expect(tasks).To(Equal([]resources.Task{resources.Task(task1), resources.Task(task2), resources.Task(task3)}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.GetApplicationTasksCallCount()).To(Equal(2))
					appGUID, query := fakeCloudControllerClient.GetApplicationTasksArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(query).To(BeNil())
				})
			})

			When("there are no associated tasks", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]resources.Task{},
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

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationTasksReturns(
					[]resources.Task{},
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
		When("the cloud controller client does not return an error", func() {
			When("the task is found", func() {
				var task1 resources.Task

				BeforeEach(func() {
					task1 = resources.Task{
						GUID:       "task-1-guid",
						SequenceID: 1,
					}
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]resources.Task{task1},
						ccv3.Warnings{"get-task-warning-1"},
						nil,
					)
				})

				It("returns the task and warnings", func() {
					task, warnings, err := actor.GetTaskBySequenceIDAndApplication(1, "some-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(task).To(Equal(resources.Task(task1)))
					Expect(warnings).To(ConsistOf("get-task-warning-1"))
				})
			})

			When("the task is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationTasksReturns(
						[]resources.Task{},
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

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("generic-error")
				fakeCloudControllerClient.GetApplicationTasksReturns(
					[]resources.Task{},
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
		When("the task exists", func() {
			var returnedTask resources.Task

			BeforeEach(func() {
				returnedTask = resources.Task{
					GUID:       "some-task-guid",
					SequenceID: 1,
				}
				fakeCloudControllerClient.UpdateTaskCancelReturns(
					returnedTask,
					ccv3.Warnings{"update-task-warning"},
					nil)
			})

			It("returns the task and warnings", func() {
				task, warnings, err := actor.TerminateTask("some-task-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("update-task-warning"))
				Expect(task).To(Equal(resources.Task(returnedTask)))
			})
		})

		When("the cloud controller returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("cc-error")
				fakeCloudControllerClient.UpdateTaskCancelReturns(
					resources.Task{},
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
