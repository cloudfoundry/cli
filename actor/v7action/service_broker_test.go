package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Broker Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetServiceBrokers", func() {
		var (
			serviceBrokers []resources.ServiceBroker
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executionError = actor.GetServiceBrokers()
		})

		When("the cloud controller request is successful", func() {
			When("the cloud controller returns service brokers", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns([]resources.ServiceBroker{
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
						resources.ServiceBroker{Name: "service-broker-1", GUID: "service-broker-guid-1", URL: "service-broker-url-1"},
						resources.ServiceBroker{Name: "service-broker-2", GUID: "service-broker-guid-2", URL: "service-broker-url-2"},
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

	Describe("GetServiceBrokerByName", func() {
		const (
			serviceBroker1Name = "broker-name"
			serviceBroker1Guid = "broker-guid"
		)

		var (
			ccv3ServiceBrokers []resources.ServiceBroker
			serviceBroker      resources.ServiceBroker
			warnings           Warnings
			executeErr         error
		)

		BeforeEach(func() {
			ccv3ServiceBrokers = []resources.ServiceBroker{
				{Name: serviceBroker1Name, GUID: serviceBroker1Guid},
			}
		})

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = actor.GetServiceBrokerByName(serviceBroker1Name)
		})

		When("the service broker is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					ccv3ServiceBrokers,
					ccv3.Warnings{"some-service-broker-warning"},
					nil,
				)
			})

			It("returns the service broker and warnings", func() {
				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{serviceBroker1Name},
				}))

				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-service-broker-warning"))
				Expect(serviceBroker).To(Equal(
					resources.ServiceBroker{Name: serviceBroker1Name, GUID: serviceBroker1Guid},
				))
			})
		})

		When("the service broker is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{},
					ccv3.Warnings{"some-other-service-broker-warning"},
					nil,
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceBrokerNotFoundError{Name: serviceBroker1Name}))
				Expect(warnings).To(ConsistOf("some-other-service-broker-warning"))
				Expect(serviceBroker).To(Equal(resources.ServiceBroker{}))
			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]resources.ServiceBroker{},
					ccv3.Warnings{"some-service-broker-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-service-broker-warning"))
				Expect(serviceBroker).To(Equal(resources.ServiceBroker{}))

				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
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
			warnings, executionError = actor.CreateServiceBroker(
				resources.ServiceBroker{
					Name:      name,
					URL:       url,
					Username:  username,
					Password:  password,
					SpaceGUID: spaceGUID,
				},
			)
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

				serviceBroker := fakeCloudControllerClient.CreateServiceBrokerArgsForCall(0)
				Expect(serviceBroker.Name).To(Equal(name))
				Expect(serviceBroker.Username).To(Equal(username))
				Expect(serviceBroker.Password).To(Equal(password))
				Expect(serviceBroker.URL).To(Equal(url))
				Expect(serviceBroker.SpaceGUID).To(Equal(spaceGUID))
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

	Describe("UpdateServiceBroker", func() {
		const (
			guid     = "broker-guid"
			url      = "url"
			username = "username"
			password = "password"
		)

		var (
			expectedJobURL = ccv3.JobURL("some-job-url")
		)

		It("passes the service broker creds and url to the client", func() {
			_, executionError := actor.UpdateServiceBroker(
				guid,
				resources.ServiceBroker{
					Username: username,
					Password: password,
					URL:      url,
				},
			)
			Expect(executionError).ToNot(HaveOccurred())

			Expect(fakeCloudControllerClient.UpdateServiceBrokerCallCount()).To(Equal(1))
			guid, serviceBroker := fakeCloudControllerClient.UpdateServiceBrokerArgsForCall(0)
			Expect(guid).To(Equal(guid))
			Expect(serviceBroker.Name).To(BeEmpty())
			Expect(serviceBroker.Username).To(Equal(username))
			Expect(serviceBroker.Password).To(Equal(password))
			Expect(serviceBroker.URL).To(Equal(url))
		})

		It("passes the job url to the client for polling", func() {
			fakeCloudControllerClient.UpdateServiceBrokerReturns(
				expectedJobURL, ccv3.Warnings{"some-update-warning"}, nil,
			)

			_, executionError := actor.UpdateServiceBroker(
				guid,
				resources.ServiceBroker{
					Username: username,
					Password: password,
					URL:      url,
				},
			)
			Expect(executionError).ToNot(HaveOccurred())

			Expect(fakeCloudControllerClient.PollJobCallCount()).To(
				Equal(1), "Expected client.PollJob to be called once",
			)

			jobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
			Expect(jobURL).To(Equal(expectedJobURL))
		})

		When("async job succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServiceBrokerReturns(
					expectedJobURL, ccv3.Warnings{"some-update-warning"}, nil,
				)
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, nil)
			})

			It("succeeds and returns warnings", func() {
				warnings, executionError := actor.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{
						Username: username,
						Password: password,
						URL:      url,
					},
				)

				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-update-warning", "some-poll-warning"))
			})
		})

		When("async job fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServiceBrokerReturns(
					expectedJobURL, ccv3.Warnings{"some-update-warning"}, nil,
				)
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, errors.New("job-execution-failed"))
			})

			It("succeeds and returns warnings", func() {
				warnings, executionError := actor.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{
						Username: username,
						Password: password,
						URL:      url,
					},
				)

				Expect(executionError).To(MatchError("job-execution-failed"))
				Expect(warnings).To(ConsistOf("some-update-warning", "some-poll-warning"))
			})
		})

		When("the client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServiceBrokerReturns(
					"", ccv3.Warnings{"some-other-warning"}, errors.New("invalid broker"),
				)
			})

			It("fails and returns warnings", func() {
				warnings, executionError := actor.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{
						Username: username,
						Password: password,
						URL:      url,
					},
				)

				Expect(executionError).To(MatchError("invalid broker"))
				Expect(warnings).To(ConsistOf("some-other-warning"))
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(
					Equal(0), "Expected client.PollJob to not have been called",
				)
			})
		})
	})

	Describe("DeleteServiceBroker", func() {
		var (
			serviceBrokerGUID = "some-service-broker-guid"
			warnings          Warnings
			executionError    error
			expectedJobURL    = ccv3.JobURL("some-job-URL")
		)

		JustBeforeEach(func() {
			warnings, executionError = actor.DeleteServiceBroker(serviceBrokerGUID)
		})

		When("the client request is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteServiceBrokerReturns(expectedJobURL, ccv3.Warnings{"some-deletion-warning"}, nil)
			})

			It("passes the service broker credentials to the client", func() {
				Expect(fakeCloudControllerClient.DeleteServiceBrokerCallCount()).To(Equal(1))
				actualServiceBrokerGUID := fakeCloudControllerClient.DeleteServiceBrokerArgsForCall(0)
				Expect(actualServiceBrokerGUID).To(Equal(serviceBrokerGUID))
			})

			It("passes the job url to the client for polling", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(
					Equal(1), "Expected client.PollJob to be called once",
				)

				jobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(jobURL).To(Equal(expectedJobURL))
			})

			When("the delete service broker job completes successfully", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-poll-warning"}, nil)
				})

				It("succeeds and returns warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(warnings).To(ConsistOf("some-deletion-warning", "some-poll-warning"))
				})
			})

			When("the delete service broker job fails", func() {
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
				fakeCloudControllerClient.DeleteServiceBrokerReturns("", ccv3.Warnings{"some-other-warning"}, errors.New("invalid broker"))
			})

			It("fails and returns warnings", func() {
				Expect(executionError).To(MatchError("invalid broker"))

				Expect(warnings).To(ConsistOf("some-other-warning"))
			})
		})
	})
})
