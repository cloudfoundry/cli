package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceSharedFrom Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServiceInstanceSharedFromByServiceInstance", func() {
		var (
			serviceInstanceGUID string

			sharedFrom ServiceInstanceSharedFrom
			warnings   Warnings
			getErr     error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
		})

		JustBeforeEach(func() {
			sharedFrom, warnings, getErr = actor.GetServiceInstanceSharedFromByServiceInstance(serviceInstanceGUID)
		})

		Context("when no errors are encountered getting the service instance shared_from object", func() {
			var returnedSharedFrom ccv2.ServiceInstanceSharedFrom
			BeforeEach(func() {
				returnedSharedFrom = ccv2.ServiceInstanceSharedFrom{
					SpaceGUID:        "some-space-guid",
					SpaceName:        "some-space-name",
					OrganizationName: "some-org-name",
				}
				fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
					returnedSharedFrom,
					ccv2.Warnings{"get-service-instance-shared-from-warning"},
					nil,
				)
			})

			It("returns the service instance shared_from object and all warnings", func() {
				Expect(getErr).ToNot(HaveOccurred())
				Expect(sharedFrom).To(Equal(ServiceInstanceSharedFrom(returnedSharedFrom)))
				Expect(warnings).To(ConsistOf("get-service-instance-shared-from-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})

		Context("when an error is encountered getting the service instance shared_from object", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
					ccv2.ServiceInstanceSharedFrom{},
					ccv2.Warnings{"get-service-instance-shared-from-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(getErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-service-instance-shared-from-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})
	})
})
