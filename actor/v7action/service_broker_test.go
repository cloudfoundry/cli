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
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
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

	Describe("CreateServiceBroker", func() {
		const (
			name      = "name"
			url       = "url"
			username  = "username"
			password  = "password"
			spaceGUID = "space-guid"
		)

		var (
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			warnings, executionError = actor.CreateServiceBroker(name, username, password, url, spaceGUID)
		})

		When("the client request is successful", func() {
			var (
				expectedJobURL = ccv3.JobURL("some-job-url")
			)

			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBrokerReturns(
					expectedJobURL, ccv3.Warnings{"some-creation-warning"}, nil,
				)
			})

			It("passes the service broker credentials to the client", func() {
				Expect(fakeCloudControllerClient.CreateServiceBrokerCallCount()).To(
					Equal(1), "Expected client.CreateServiceBroker to be called once",
				)
				// FIXME: nu pls, put me in a single object, pls :’(
				n, u, p, l, s := fakeCloudControllerClient.CreateServiceBrokerArgsForCall(0)
				Expect(n).To(Equal(name))
				Expect(u).To(Equal(username))
				Expect(p).To(Equal(password))
				Expect(l).To(Equal(url))
				Expect(s).To(Equal(spaceGUID))
			})

			It("passes the job url to the client for polling", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(
					Equal(1), "Expected client.PollJob to be called once",
				)

				jobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(jobURL).To(Equal(expectedJobURL))
			})

			When("async job succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, nil)
				})

				It("succeeds and returns warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(warnings).To(ConsistOf("some-creation-warning", "some-poll-warning"))
				})
			})

			When("async job fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(nil, errors.New("oopsie"))
				})

				It("succeeds and returns warnings", func() {
					Expect(executionError).To(MatchError("oopsie"))
				})
			})
		})

		When("the client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBrokerReturns(
					"", ccv3.Warnings{"some-other-warning"}, errors.New("invalid broker"),
				)
			})

			It("fails and returns warnings", func() {
				Expect(executionError).To(MatchError("invalid broker"))

				Expect(warnings).To(ConsistOf("some-other-warning"))
			})
		})
	})
})
