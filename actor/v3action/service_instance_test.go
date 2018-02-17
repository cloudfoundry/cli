package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		var (
			serviceInstanceName string
			sourceSpaceGUID     string

			serviceInstance ServiceInstance
			warnings        Warnings
			executionError  error
		)

		BeforeEach(func() {
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
		})

		JustBeforeEach(func() {
			serviceInstance, warnings, executionError = actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)
		})

		Context("when the cloud controller request is successful", func() {
			Context("when the cloud controller returns one service instance", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{
						{
							Name: "some-service-instance",
							GUID: "some-service-instance-guid",
						},
					}, ccv3.Warnings{"some-service-instance-warning"}, nil)
				})

				It("returns a service instance and warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(serviceInstance).To(Equal(ServiceInstance{Name: "some-service-instance", GUID: "some-service-instance-guid"}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
					Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{sourceSpaceGUID}},
					))
				})
			})

			Context("when the cloud controller returns no service instances", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns(
						nil,
						ccv3.Warnings{"some-service-instance-warning"},
						nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))

					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				})
			})
		})

		Context("when the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					nil,
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("no service instance"))
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service instance"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})

	Describe("UnshareServiceInstanceByServiceInstanceAndSpace", func() {
		var (
			serviceInstanceGUID string
			sharedToSpaceGUID   string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
			sharedToSpaceGUID = "some-other-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID, sharedToSpaceGUID)
		})

		Context("when no errors occur deleting the service instance share relationship", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceReturns(
					ccv3.Warnings{"delete-share-relationship-warning"},
					nil)
			})

			It("returns no errors and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("delete-share-relationship-warning"))

				Expect(fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceCallCount()).To(Equal(1))
				serviceInstanceGUIDArg, sharedToSpaceGUIDArg := fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceArgsForCall(0)
				Expect(serviceInstanceGUIDArg).To(Equal(serviceInstanceGUID))
				Expect(sharedToSpaceGUIDArg).To(Equal(sharedToSpaceGUID))
			})
		})

		Context("when an error occurs deleting the service instance share relationship", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("delete share relationship error")
				fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceReturns(
					ccv3.Warnings{"delete-share-relationship-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("delete-share-relationship-warning"))
			})
		})
	})
})
