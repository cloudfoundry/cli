package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetService", func() {
		var (
			service         Service
			serviceWarnings Warnings
			serviceErr      error
		)

		JustBeforeEach(func() {
			service, serviceWarnings, serviceErr = actor.GetService("some-service-guid")
		})

		When("no errors are encountered getting the service", func() {
			var returnedService ccv2.Service

			BeforeEach(func() {
				returnedService = ccv2.Service{
					GUID:             "some-service-guid",
					Label:            "some-service",
					Description:      "some-description",
					DocumentationURL: "some-url",
				}
				fakeCloudControllerClient.GetServiceReturns(
					returnedService,
					ccv2.Warnings{"get-service-warning"},
					nil)
			})

			It("returns the service and all warnings", func() {
				Expect(serviceErr).ToNot(HaveOccurred())
				Expect(service).To(Equal(Service(returnedService)))
				Expect(serviceWarnings).To(ConsistOf("get-service-warning"))

				Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal("some-service-guid"))
			})
		})

		When("an error is encountered getting the service", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetServiceReturns(
					ccv2.Service{},
					ccv2.Warnings{"get-service-warning"},
					expectedErr)
			})

			It("returns the errors and all warnings", func() {
				Expect(serviceErr).To(MatchError(expectedErr))
				Expect(serviceWarnings).To(ConsistOf("get-service-warning"))

				Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal("some-service-guid"))
			})
		})
	})

	Describe("GetServiceByName", func() {
		var (
			service         Service
			serviceWarnings Warnings
			serviceErr      error
		)

		JustBeforeEach(func() {
			service, serviceWarnings, serviceErr = actor.GetServiceByName("some-service")
		})

		When("services are returned from the client", func() {
			var returnedServices []ccv2.Service

			BeforeEach(func() {
				returnedServices = []ccv2.Service{
					{
						GUID:             "some-service-guid",
						Label:            "some-service",
						Description:      "some-description",
						DocumentationURL: "some-url",
					},
				}

				fakeCloudControllerClient.GetServicesReturns(
					returnedServices,
					ccv2.Warnings{"get-services-warning"},
					nil)
			})

			It("returns the service and all warnings", func() {
				Expect(serviceErr).ToNot(HaveOccurred())
				Expect(service).To(Equal(Service(returnedServices[0])))
				Expect(serviceWarnings).To(ConsistOf("get-services-warning"))

				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicesArgsForCall(0)).To(Equal([]ccv2.Filter{{
					Type:     constant.LabelFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service"},
				}}))
			})
		})

		When("there are no services returned by the client", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"get-services-warning"},
					nil)
			})

			It("returns a ServiceNotFoundError and all warnings", func() {
				Expect(serviceErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "some-service"}))
				Expect(serviceWarnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("the client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"get-services-warning"},
					errors.New("client-error"))
			})

			It("propagates the error and all warnings", func() {
				Expect(serviceErr).To(MatchError(errors.New("client-error")))
				Expect(serviceWarnings).To(ConsistOf("get-services-warning"))
			})
		})
	})
})
