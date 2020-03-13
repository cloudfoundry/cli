package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Offering Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
	})

	Describe("GetServiceOfferingByNameAndBroker", func() {
		const (
			serviceOfferingName = "myServiceOffering"
		)

		var (
			serviceOffering   ServiceOffering
			warnings          Warnings
			executionError    error
			serviceBrokerName string
		)

		BeforeEach(func() {
			serviceBrokerName = ""
		})

		JustBeforeEach(func() {
			serviceOffering, warnings, executionError = actor.GetServiceOfferingByNameAndBroker(
				serviceOfferingName,
				serviceBrokerName,
			)
		})

		When("the cloud controller request is successful", func() {
			BeforeEach(func() {
				serviceBrokerName = "myServiceBroker"
			})

			When("the cloud controller returns one service offering", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{
						{
							Name: "some-service-offering",
							GUID: "some-service-offering-guid",
						},
					}, ccv3.Warnings{"some-service-offering-warning"}, nil)
				})

				It("returns a service offering and warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(serviceOffering).To(Equal(ServiceOffering{Name: "some-service-offering", GUID: "some-service-offering-guid"}))
					Expect(warnings).To(ConsistOf("some-service-offering-warning"))
					Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceOfferingName}},
						ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{"myServiceBroker"}},
					))
				})
			})

			When("the cloud controller returns no service offerings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns(
						nil,
						ccv3.Warnings{"some-service-offering-warning"},
						nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceNotFoundError{
						Name:   serviceOfferingName,
						Broker: "myServiceBroker",
					}))
					Expect(warnings).To(ConsistOf("some-service-offering-warning"))
				})
			})

			When("the cloud controller returns more than one service offering", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{
						{
							Name:              "some-service-offering-1",
							GUID:              "some-service-offering-guid-1",
							ServiceBrokerName: "a-service-broker",
						},
						{
							Name:              "some-service-offering-2",
							GUID:              "some-service-offering-guid-2",
							ServiceBrokerName: "another-service-broker",
						},
					}, ccv3.Warnings{"some-service-offering-warning"}, nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.DuplicateServiceError{Name: serviceOfferingName, ServiceBrokers: []string{"a-service-broker", "another-service-broker"}}))
					Expect(warnings).To(ConsistOf("some-service-offering-warning"))
				})
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingsReturns(
					nil,
					ccv3.Warnings{"some-service-offering-warning"},
					errors.New("no service offering"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service offering"))
				Expect(warnings).To(ConsistOf("some-service-offering-warning"))
			})
		})

		When("the broker name is not provided", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{
					{
						Name: "some-service-offering",
						GUID: "some-service-offering-guid",
					},
				}, ccv3.Warnings{"some-service-offering-warning"}, nil)
			})

			It("queries only on the service offering name", func() {
				Expect(executionError).NotTo(HaveOccurred())

				Expect(serviceOffering).To(Equal(ServiceOffering{Name: "some-service-offering", GUID: "some-service-offering-guid"}))
				Expect(warnings).To(ConsistOf("some-service-offering-warning"))
				Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceOfferingName}},
				))
			})
		})
	})
})
