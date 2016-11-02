package v3actions_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v3actions/v3actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "code.cloudfoundry.org/cli/actors/v3actions"

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
		Context("when the app exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.RunTaskReturns(
					ccv3.Task{
						SequenceID: 3,
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("runs the task and returns the task and warnings", func() {
				task, warnings, err := actor.RunTask("some-app-guid", "some command")
				Expect(err).ToNot(HaveOccurred())
				Expect(task).To(Equal(Task{
					SequenceID: 3,
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.RunTaskCallCount()).To(Equal(1))
				appGUIDArg, commandArg := fakeCloudControllerClient.RunTaskArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(commandArg).To(Equal("some command"))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.RunTaskReturns(
					ccv3.Task{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.RunTask("some-app-guid", "some command")
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(fakeCloudControllerClient.RunTaskCallCount()).To(Equal(1))
				appGUIDArg, commandArg := fakeCloudControllerClient.RunTaskArgsForCall(0)
				Expect(appGUIDArg).To(Equal("some-app-guid"))
				Expect(commandArg).To(Equal("some command"))
			})
		})
	})
})
