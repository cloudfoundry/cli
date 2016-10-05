package v2actions_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/actors/v2actions/v2actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionsfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionsfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("GetServiceInstanceBySpace", func() {
		Context("when the service instance exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					[]cloudcontrollerv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					cloudcontrollerv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := actor.GetServiceInstanceBySpace("some-service-instance", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID: "some-service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf([]cloudcontrollerv2.Query{
					cloudcontrollerv2.Query{
						Filter:   cloudcontrollerv2.NameFilter,
						Operator: cloudcontrollerv2.EqualOperator,
						Value:    "some-service-instance",
					},
					cloudcontrollerv2.Query{
						Filter:   cloudcontrollerv2.SpaceGUIDFilter,
						Operator: cloudcontrollerv2.EqualOperator,
						Value:    "some-space-guid",
					},
				}))
			})
		})

		Context("when the service instance does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]cloudcontrollerv2.ServiceInstance{}, nil, nil)
			})

			It("returns a ServiceInstanceNotFoundError", func() {
				_, _, err := actor.GetServiceInstanceBySpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(ServiceInstanceNotFoundError{Name: "some-service-instance"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetServiceInstancesReturns([]cloudcontrollerv2.ServiceInstance{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetServiceInstanceBySpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})
})
