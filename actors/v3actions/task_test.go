package v3actions_test

import (
	"errors"
	"net/url"

	. "code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/actors/v3actions/v3actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionsfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionsfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("RunTask", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.RunTaskReturns(
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
				task, warnings, err := actor.RunTask("some-app-guid", "some command")
				Expect(err).ToNot(HaveOccurred())

				Expect(task).To(Equal(Task{
					SequenceID: 3,
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.RunTaskCallCount()).To(Equal(1))
				appGUIDArg, commandArg := fakeCloudControllerClient.RunTaskArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(commandArg).To(Equal("some command"))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.RunTaskReturns(
					ccv3.Task{},
					ccv3.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				_, warnings, err := actor.RunTask("some-app-guid", "some command")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
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

				It("returns a TasksNotFoundError", func() {
					_, _, err := actor.GetApplicationTasks("some-app-guid", Descending)
					Expect(err).To(MatchError(TasksNotFoundError{}))
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
})
