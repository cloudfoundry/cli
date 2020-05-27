package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Offering Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("PurgeServiceOfferingByNameAndBroker", func() {
		Describe("steps", func() {
			const fakeServiceOfferingGUID = "fake-service-offering-guid"
			var (
				warnings     Warnings
				executeError error
			)

			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					ccv3.ServiceOffering{GUID: fakeServiceOfferingGUID},
					ccv3.Warnings{"a warning"},
					nil,
				)

				fakeCloudControllerClient.PurgeServiceOfferingReturns(
					ccv3.Warnings{"another warning"},
					nil,
				)

				warnings, executeError = actor.PurgeServiceOfferingByNameAndBroker("fake-service-offering", "fake-service-broker")
			})

			It("requests the service offering guid", func() {
				Expect(fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerCallCount()).To(Equal(1))
				actualOffering, actualBroker := fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerArgsForCall(0)
				Expect(actualOffering).To(Equal("fake-service-offering"))
				Expect(actualBroker).To(Equal("fake-service-broker"))
			})

			It("requests the purge of the service offering", func() {
				Expect(fakeCloudControllerClient.PurgeServiceOfferingCallCount()).To(Equal(1))
				actualGUID := fakeCloudControllerClient.PurgeServiceOfferingArgsForCall(0)
				Expect(actualGUID).To(Equal(fakeServiceOfferingGUID))
			})

			It("return all warinings and no errors", func() {
				Expect(executeError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("a warning", "another warning"))
			})
		})

		DescribeTable(
			"when getting the service offering fails ",
			func(clientError, expectedError error) {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(ccv3.ServiceOffering{}, ccv3.Warnings{"a warning"}, clientError)

				warnings, err := actor.PurgeServiceOfferingByNameAndBroker("fake-service-offering", "fake-service-broker")
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("a warning"))
			},
			Entry(
				"ServiceOfferingNameAmbiguityError",
				ccerror.ServiceOfferingNameAmbiguityError{
					ServiceOfferingName: "foo",
					ServiceBrokerNames:  []string{"bar", "baz"},
				},
				actionerror.ServiceOfferingNameAmbiguityError{
					ServiceOfferingNameAmbiguityError: ccerror.ServiceOfferingNameAmbiguityError{
						ServiceOfferingName: "foo",
						ServiceBrokerNames:  []string{"bar", "baz"},
					},
				},
			),
			Entry(
				"other error",
				errors.New("boom"),
				errors.New("boom"),
			),
		)

		When("purging the service offering fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
					ccv3.ServiceOffering{GUID: "fake-service-offering-guid"},
					ccv3.Warnings{"a warning"},
					nil,
				)

				fakeCloudControllerClient.PurgeServiceOfferingReturns(
					ccv3.Warnings{"another warning"},
					errors.New("ouch"),
				)
			})

			It("return all warinings and errors", func() {
				warnings, err := actor.PurgeServiceOfferingByNameAndBroker("fake-service-offering", "fake-service-broker")

				Expect(err).To(MatchError(errors.New("ouch")))
				Expect(warnings).To(ConsistOf("a warning", "another warning"))
			})
		})
	})
})
