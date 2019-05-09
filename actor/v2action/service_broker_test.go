package v2action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
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

	Describe("GetServiceBrokers", func() {
		var (
			executeErr     error
			warnings       Warnings
			serviceBrokers []ServiceBroker
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executeErr = actor.GetServiceBrokers()
		})

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceBrokersReturns([]ccv2.ServiceBroker{
				{GUID: "some-broker-guid"}},
				[]string{"a-warning", "another-warning"}, nil)
		})

		It("retrieves all service brokers", func() {
			Expect(serviceBrokers).To(HaveLen(1))
			Expect(serviceBrokers[0].GUID).To(Equal("some-broker-guid"))
			Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
		})

		When("there is an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns([]ccv2.ServiceBroker{}, []string{"one-warning", "two-warnings"}, errors.New("error getting brokers"))
			})

			It("returns the errors and warnings", func() {
				Expect(warnings).To(Equal(Warnings{"one-warning", "two-warnings"}))
				Expect(executeErr).To(MatchError("error getting brokers"))
			})
		})
	})

	Describe("GetServiceBrokerByName", func() {
		var (
			executeErr    error
			warnings      Warnings
			serviceBroker ServiceBroker
		)

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = actor.GetServiceBrokerByName("broker-name")
		})

		When("there are no errors", func() {
			When("a service broker exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns([]ccv2.ServiceBroker{{GUID: "some-broker-guid"}}, []string{"a-warning", "another-warning"}, nil)
				})

				It("gets the service broker", func() {
					Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
					filter := fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)
					Expect(filter).To(Equal([]ccv2.Filter{{Type: constant.NameFilter, Operator: constant.EqualOperator, Values: []string{"broker-name"}}}))

					Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(serviceBroker.GUID).To(Equal("some-broker-guid"))
				})
			})

		})

		When("there is an error", func() {
			When("calling the client", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns([]ccv2.ServiceBroker{{GUID: "some-broker-guid"}}, []string{"one-warning", "two-warnings"}, errors.New("error creating broker"))
				})

				It("returns the errors and warnings", func() {
					Expect(warnings).To(Equal(Warnings{"one-warning", "two-warnings"}))
					Expect(executeErr).To(MatchError("error creating broker"))
				})
			})

			When("a service broker does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns([]ccv2.ServiceBroker{}, []string{"a-warning", "another-warning"}, nil)
				})

				It("raises a service broker not found error", func() {
					Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
					filter := fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)
					Expect(filter).To(Equal([]ccv2.Filter{{Type: constant.NameFilter, Operator: constant.EqualOperator, Values: []string{"broker-name"}}}))

					Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
					Expect(executeErr).To(MatchError(actionerror.ServiceBrokerNotFoundError{Key: actionerror.KeyName, Value: "broker-name"}))
				})
			})
		})
	})

	Describe("GetServiceBroker", func() {
		var (
			executeErr    error
			warnings      Warnings
			serviceBroker ServiceBroker
		)

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = actor.GetServiceBroker("broker-guid")
		})

		When("there are no errors", func() {
			When("a service broker exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokerReturns(ccv2.ServiceBroker{GUID: "broker-guid"}, []string{"a-warning", "another-warning"}, nil)
				})

				It("gets the service broker", func() {
					Expect(fakeCloudControllerClient.GetServiceBrokerCallCount()).To(Equal(1))
					brokerGUID := fakeCloudControllerClient.GetServiceBrokerArgsForCall(0)
					Expect(brokerGUID).To(Equal("broker-guid"))

					Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(serviceBroker.GUID).To(Equal("broker-guid"))
				})
			})
		})

		When("there is an error", func() {
			When("calling the client", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokerReturns(ccv2.ServiceBroker{GUID: "broker-guid"}, []string{"one-warning", "two-warnings"}, errors.New("error fetching broker"))
				})

				It("returns the errors and warnings", func() {
					Expect(warnings).To(Equal(Warnings{"one-warning", "two-warnings"}))
					Expect(executeErr).To(MatchError("error fetching broker"))
				})
			})

			When("a service broker does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokerReturns(ccv2.ServiceBroker{}, []string{"a-warning", "another-warning"}, ccerror.ResourceNotFoundError{})
				})

				It("raises a service broker not found error", func() {
					Expect(fakeCloudControllerClient.GetServiceBrokerCallCount()).To(Equal(1))
					brokerGUID := fakeCloudControllerClient.GetServiceBrokerArgsForCall(0)
					Expect(brokerGUID).To(Equal("broker-guid"))

					Expect(warnings).To(Equal(Warnings{"a-warning", "another-warning"}))
					Expect(executeErr).To(MatchError(actionerror.ServiceBrokerNotFoundError{Key: actionerror.KeyGUID, Value: "broker-guid"}))
				})
			})
		})
	})
})
