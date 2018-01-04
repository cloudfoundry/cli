package v3action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("instance actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("Instance", func() {
		Describe("StartTime", func() {
			It("returns the time that the instance started", func() {
				instance := ProcessInstance{Uptime: 86400}
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now().Add(-24*time.Hour), 10*time.Second))
			})
		})
	})

	Describe("DeleteInstanceByApplicationNameSpaceProcessTypeAndIndex", func() {
		var (
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteInstanceByApplicationNameSpaceProcessTypeAndIndex("some-app-name", "some-space-guid", "some-process-type", 666)
		})

		Context("when getting the application returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{}, ccv3.Warnings{"some-get-app-warning"}, errors.New("some-get-app-error"))
			})

			It("returns all warnings and the error", func() {
				Expect(executeErr).To(MatchError("some-get-app-error"))
				Expect(warnings).To(ConsistOf("some-get-app-warning"))
			})
		})

		Context("when getting the application succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{GUID: "some-app-guid"}}, ccv3.Warnings{"some-get-app-warning"}, nil)
			})

			Context("when deleting the instance returns ProcessNotFoundError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationProcessInstanceReturns(ccv3.Warnings{"some-delete-warning"}, ccerror.ProcessNotFoundError{})
				})

				It("returns all warnings and the ProcessNotFoundError error", func() {
					Expect(executeErr).To(Equal(actionerror.ProcessNotFoundError{ProcessType: "some-process-type"}))
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-warning"))
				})
			})

			Context("when deleting the instance returns InstanceNotFoundError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationProcessInstanceReturns(ccv3.Warnings{"some-delete-warning"}, ccerror.InstanceNotFoundError{})
				})

				It("returns all warnings and the ProcessInstanceNotFoundError error", func() {
					Expect(executeErr).To(Equal(actionerror.ProcessInstanceNotFoundError{ProcessType: "some-process-type", InstanceIndex: 666}))
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-warning"))
				})
			})

			Context("when deleting the instance returns other error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationProcessInstanceReturns(ccv3.Warnings{"some-delete-warning"}, errors.New("some-delete-error"))
				})

				It("returns all warnings and the error", func() {
					Expect(executeErr).To(MatchError("some-delete-error"))
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-warning"))
				})
			})

			Context("when deleting the instance succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteApplicationProcessInstanceReturns(ccv3.Warnings{"some-delete-warning"}, nil)
				})

				It("returns all warnings and no error", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-get-app-warning", "some-delete-warning"))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					))

					Expect(fakeCloudControllerClient.DeleteApplicationProcessInstanceCallCount()).To(Equal(1))
					appGUID, processType, instanceIndex := fakeCloudControllerClient.DeleteApplicationProcessInstanceArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(processType).To(Equal("some-process-type"))
					Expect(instanceIndex).To(Equal(666))
				})
			})
		})
	})
})
