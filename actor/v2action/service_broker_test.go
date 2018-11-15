package v2action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

var _ = Describe("Service Broker", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateServiceBroker", func() {
		var (
			executeErr    error
			warnings      Warnings
			serviceBroker ServiceBroker
		)

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = actor.CreateServiceBroker("broker-name", "username", "password", "https://broker.com", "a-space-guid")
		})

		When("there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBrokerReturns(ccv2.ServiceBroker{GUID: "some-broker-guid"}, []string{"a-warning", "another-warning"}, nil)
			})

			It("creates the service broker with the correct values", func() {
				Expect(fakeCloudControllerClient.CreateServiceBrokerCallCount()).To(Equal(1))
				brokerName, username, password, url, space := fakeCloudControllerClient.CreateServiceBrokerArgsForCall(0)
				Expect(brokerName).To(Equal("broker-name"))
				Expect(username).To(Equal("username"))
				Expect(password).To(Equal("password"))
				Expect(url).To(Equal("https://broker.com"))
				Expect(space).To(Equal("a-space-guid"))

				Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBroker.GUID).To(Equal("some-broker-guid"))
			})
		})

		When("there is an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBrokerReturns(ccv2.ServiceBroker{}, []string{"one-warning", "two-warnings"}, errors.New("error creating broker"))
			})

			It("returns the errors and warnings", func() {
				Expect(warnings).To(Equal(Warnings{"one-warning", "two-warnings"}))
				Expect(executeErr).To(MatchError("error creating broker"))
			})
		})
	})
})
