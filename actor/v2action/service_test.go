package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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

		Context("when no errors are encountered getting the service", func() {
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

		Context("when an error is encountered getting the service", func() {
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
})
