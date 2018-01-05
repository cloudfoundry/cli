package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceSharedTo Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServiceInstanceSharedTosByServiceInstance", func() {
		var (
			serviceInstanceGUID string

			sharedTos []ServiceInstanceSharedTo
			warnings  Warnings
			getErr    error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
		})

		JustBeforeEach(func() {
			sharedTos, warnings, getErr = actor.GetServiceInstanceSharedTosByServiceInstance(serviceInstanceGUID)
		})

		Context("when no errors are encountered getting the service instance shared_to list", func() {
			var returnedSharedTos []ccv2.ServiceInstanceSharedTo
			BeforeEach(func() {
				returnedSharedTos = []ccv2.ServiceInstanceSharedTo{
					{
						SpaceGUID:        "some-space-guid",
						SpaceName:        "some-space-name",
						OrganizationName: "some-org-name",
						BoundAppCount:    3,
					},
					{
						SpaceGUID:        "another-space-guid",
						SpaceName:        "another-space-name",
						OrganizationName: "another-org-name",
						BoundAppCount:    2,
					},
				}
				fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
					returnedSharedTos,
					ccv2.Warnings{"get-service-instance-shared-to-warning"},
					nil,
				)
			})

			It("returns the service instance shared_to list and all warnings", func() {
				Expect(getErr).ToNot(HaveOccurred())
				Expect(sharedTos).To(ConsistOf(ServiceInstanceSharedTo(returnedSharedTos[0]), ServiceInstanceSharedTo(returnedSharedTos[1])))
				Expect(warnings).To(ConsistOf("get-service-instance-shared-to-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})

		Context("when an error is encountered getting the service instance shared_to list", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
					nil,
					ccv2.Warnings{"get-service-instance-shared-to-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(getErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-service-instance-shared-to-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})
	})
})
