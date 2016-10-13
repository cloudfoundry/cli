package v2actions_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/actors/v2actions/v2actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

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
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID: "some-service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf([]ccv2.Query{
					ccv2.Query{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-service-instance",
					},
					ccv2.Query{
						Filter:   ccv2.SpaceGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-space-guid",
					},
				}))
			})
		})

		Context("when the service instance does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, nil)
			})

			It("returns a ServiceInstanceNotFoundError", func() {
				_, _, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(ServiceInstanceNotFoundError{Name: "some-service-instance"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetSpaceServiceInstanceByName", func() {
		Context("when the service instance exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := actor.GetSpaceServiceInstanceByName("some-space-guid", "some-service-instance")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID: "some-service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))

				spaceGUID, includeUserProvidedServices, queries := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(includeUserProvidedServices).To(BeTrue())
				Expect(queries).To(ConsistOf([]ccv2.Query{
					ccv2.Query{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-service-instance",
					},
				}))
			})
		})

		Context("when the service instance does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, nil)
			})

			It("returns a ServiceInstanceNotFoundError", func() {
				_, _, err := actor.GetSpaceServiceInstanceByName("some-space-guid", "some-service-instance")
				Expect(err).To(MatchError(ServiceInstanceNotFoundError{Name: "some-service-instance"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetSpaceServiceInstanceByName("some-space-guid", "some-service-instance")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})
})
