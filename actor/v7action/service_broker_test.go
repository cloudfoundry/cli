package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Broker Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetServiceBrokers", func() {
		var (
			serviceBrokers []ServiceBroker
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executionError = actor.GetServiceBrokers()
		})

		When("the cloud controller request is successful", func() {
			When("the cloud controller returns service brokers", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns([]ccv3.ServiceBroker{
						{
							GUID: "service-broker-guid-1",
							Name: "service-broker-1",
							URL:  "service-broker-url-1",
						},
						{
							GUID: "service-broker-guid-2",
							Name: "service-broker-2",
							URL:  "service-broker-url-2",
						},
					}, ccv3.Warnings{"some-service-broker-warning"}, nil)
				})

				It("returns the service brokers and warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(serviceBrokers).To(ConsistOf(
						ServiceBroker{Name: "service-broker-1", GUID: "service-broker-guid-1", URL: "service-broker-url-1"},
						ServiceBroker{Name: "service-broker-2", GUID: "service-broker-guid-2", URL: "service-broker-url-2"},
					))
					Expect(warnings).To(ConsistOf("some-service-broker-warning"))
					Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
				})
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					nil,
					ccv3.Warnings{"some-service-broker-warning"},
					errors.New("no service broker"))
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service broker"))
				Expect(warnings).To(ConsistOf("some-service-broker-warning"))
			})
		})
	})
})
