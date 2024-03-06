package v2action_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo/v2"
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

	Describe("GetServiceByNameAndBrokerName", func() {
		var (
			service           Service
			serviceWarnings   Warnings
			serviceErr        error
			serviceBrokerName string
		)

		BeforeEach(func() {
			serviceBrokerName = ""
		})

		JustBeforeEach(func() {
			service, serviceWarnings, serviceErr = actor.GetServiceByNameAndBrokerName("some-service", serviceBrokerName)
		})

		When("broker name is empty", func() {
			It("should not fetch a broker", func() {
				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(0))
			})

			When("one service is returned from the client", func() {
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

			When("multiple services are returned from the client", func() {
				var returnedServices []ccv2.Service

				BeforeEach(func() {
					returnedServices = []ccv2.Service{
						{
							GUID:             "some-service-1-guid",
							Label:            "some-service-1",
							Description:      "some-description",
							DocumentationURL: "some-url",
						},
						{
							GUID:             "some-service-2-guid",
							Label:            "some-service-2",
							Description:      "some-description",
							DocumentationURL: "some-url",
						},
					}

					fakeCloudControllerClient.GetServicesReturns(
						returnedServices,
						ccv2.Warnings{"get-services-warning"},
						nil)
				})

				It("returns a DuplicateServiceError and all warnings", func() {
					Expect(serviceErr).To(MatchError(actionerror.DuplicateServiceError{Name: "some-service"}))
					Expect(serviceWarnings).To(ConsistOf("get-services-warning"))
				})

			})
		})

		When("broker name is provided", func() {
			var returnedBrokers []ccv2.ServiceBroker

			BeforeEach(func() {
				serviceBrokerName = "some-broker"

				returnedBrokers = []ccv2.ServiceBroker{
					{
						Name: "some-broker",
						GUID: "some-broker-guid",
					},
				}

				fakeCloudControllerClient.GetServiceBrokersReturns(
					returnedBrokers,
					ccv2.Warnings{"get-services-warning"},
					nil)
			})

			It("fetches the the broker by name", func() {
				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)).To(Equal([]ccv2.Filter{{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{serviceBrokerName},
				}}))
			})

			When("fetching the broker by name errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns(
						nil,
						ccv2.Warnings{"get-services-warning"},
						errors.New("failed-to-fetch-broker"))
				})

				It("propagates the error and all warnings", func() {
					Expect(serviceErr).To(MatchError("failed-to-fetch-broker"))
					Expect(serviceWarnings).To(ConsistOf("get-services-warning"))
				})
			})

			When("one service is returned from the client", func() {
				var returnedServices []ccv2.Service

				BeforeEach(func() {
					returnedServices = []ccv2.Service{
						{
							GUID:              "some-service-guid",
							Label:             "some-service",
							Description:       "some-description",
							DocumentationURL:  "some-url",
							ServiceBrokerName: "some-broker",
						},
					}

					fakeCloudControllerClient.GetServicesReturns(
						returnedServices,
						ccv2.Warnings{"get-services-warning"},
						nil)
				})

				It("returns the service filtered by label and broker guid and all warnings", func() {
					Expect(serviceErr).ToNot(HaveOccurred())
					Expect(service).To(Equal(Service(returnedServices[0])))
					Expect(serviceWarnings).To(ConsistOf("get-services-warning"))

					Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicesArgsForCall(0)).To(ConsistOf(
						ccv2.Filter{
							Type:     constant.LabelFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-service"},
						},
						ccv2.Filter{
							Type:     constant.ServiceBrokerGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-broker-guid"},
						},
					))
				})
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

	Describe("GetServicesWithPlans", func() {
		When("the broker has no services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, nil, nil)
			})

			It("returns no services", func() {
				servicesWithPlans, _, err := actor.GetServicesWithPlans(Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(servicesWithPlans).To(HaveLen(0))
			})
		})

		When("there is a service with no plans", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{
							GUID:  "some-service-guid-1",
							Label: "some-service-label-1",
						},
					},
					nil, nil)

				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, nil, nil)
			})

			It("returns a service with no plans", func() {
				servicesWithPlans, _, err := actor.GetServicesWithPlans(Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(servicesWithPlans).To(HaveLen(1))
				Expect(servicesWithPlans).To(HaveKeyWithValue(
					Service{GUID: "some-service-guid-1", Label: "some-service-label-1"},
					[]ServicePlan{},
				))
			})
		})

		When("there are services with plans", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{
							GUID:  "some-service-guid-1",
							Label: "some-service-label-1",
						},
						{
							GUID:  "some-service-guid-2",
							Label: "some-service-label-2",
						},
					},
					ccv2.Warnings{"get-service-warning"}, nil)

				fakeCloudControllerClient.GetServicePlansReturnsOnCall(0,
					[]ccv2.ServicePlan{
						{
							GUID: "some-plan-guid-1",
							Name: "some-plan-name-1",
						},
						{
							GUID: "some-plan-guid-2",
							Name: "some-plan-name-2",
						},
					},
					ccv2.Warnings{"get-plan-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturnsOnCall(1,
					[]ccv2.ServicePlan{
						{
							GUID: "some-plan-guid-3",
							Name: "some-plan-name-3",
						},
						{
							GUID: "some-plan-guid-4",
							Name: "some-plan-name-4",
						},
					},
					nil, nil)
			})

			It("returns all services with associated plans and warnings", func() {
				servicesWithPlans, warnings, err := actor.GetServicesWithPlans(Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				},
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(servicesWithPlans).To(HaveLen(2))
				Expect(servicesWithPlans).To(HaveKeyWithValue(
					Service{GUID: "some-service-guid-1", Label: "some-service-label-1"},
					[]ServicePlan{
						{GUID: "some-plan-guid-1", Name: "some-plan-name-1"},
						{GUID: "some-plan-guid-2", Name: "some-plan-name-2"},
					},
				))
				Expect(servicesWithPlans).To(HaveKeyWithValue(
					Service{GUID: "some-service-guid-2", Label: "some-service-label-2"},
					[]ServicePlan{
						{GUID: "some-plan-guid-3", Name: "some-plan-name-3"},
						{GUID: "some-plan-guid-4", Name: "some-plan-name-4"},
					},
				))
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicesArgsForCall(0)).To(ConsistOf(ccv2.Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				}))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(ccv2.Filter{
					Type:     constant.ServiceGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service-guid-1"},
				}))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(1)).To(ConsistOf(ccv2.Filter{
					Type:     constant.ServiceGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service-guid-2"},
				}))

				Expect(warnings).To(ConsistOf("get-service-warning", "get-plan-warning"))
			})

			When("a service name is provided", func() {
				It("filters by service label", func() {
					_, _, err := actor.GetServicesWithPlans(Filter{
						Type:     constant.ServiceBrokerGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-broker-guid"},
					},
						Filter{
							Type:     constant.LabelFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-service-name"},
						})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicesArgsForCall(0)).To(ConsistOf(
						ccv2.Filter{
							Type:     constant.ServiceBrokerGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-broker-guid"},
						},
						ccv2.Filter{
							Type:     constant.LabelFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-service-name"},
						},
					))
				})
			})
		})

		When("fetching services returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-service-warning"}, errors.New("EXPLODE"))
			})

			It("propagates the error and warnings", func() {
				_, warnings, err := actor.GetServicesWithPlans(Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				})
				Expect(err).To(MatchError("EXPLODE"))

				Expect(warnings).To(ConsistOf("get-service-warning"))
			})
		})

		When("fetching plans for a service returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{
							GUID:  "some-service-guid-1",
							Label: "some-service-label-1",
						},
					},
					ccv2.Warnings{"get-service-warning"}, nil)

				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, ccv2.Warnings{"get-plan-warning"}, errors.New("EXPLODE"))
			})

			It("propagates the error and warnings", func() {
				_, warnings, err := actor.GetServicesWithPlans(Filter{
					Type:     constant.ServiceBrokerGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-broker-guid"},
				})

				Expect(err).To(MatchError("EXPLODE"))

				Expect(warnings).To(ConsistOf("get-service-warning", "get-plan-warning"))
			})
		})
	})

	Describe("ServiceExistsWithName", func() {
		var (
			exists   bool
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			exists, warnings, err = actor.ServiceExistsWithName("some-service")
		})

		When("a service exists with that name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{
					{
						GUID:             "some-service-guid",
						Label:            "some-service",
						Description:      "some-description",
						DocumentationURL: "some-url",
					},
				}, ccv2.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("succeeds, returning true and warnings", func() {
				Expect(exists).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetServicesArgsForCall(0)
				Expect(filters).To(Equal(
					[]ccv2.Filter{{
						Type:     constant.LabelFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-service"},
					}}))
			})
		})

		When("no service exists with that name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("returns false and warnings", func() {
				Expect(exists).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("fetching services throws an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"warning-1", "warning-2"}, errors.New("boom"))
			})

			It("propagates the error and warnings", func() {
				Expect(err).To(MatchError("boom"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("PurgeServiceOffering", func() {
		var (
			warnings Warnings
			purgeErr error
		)

		JustBeforeEach(func() {
			warnings, purgeErr = actor.PurgeServiceOffering(Service{
				Label: "some-service",
				GUID:  "some-service-guid",
			})
		})

		When("purging the service succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteServiceReturns(
					ccv2.Warnings{"delete-service-warning"},
					nil,
				)
			})

			It("should purge the returned service instance and return any warnings", func() {
				Expect(purgeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("delete-service-warning"))

				Expect(fakeCloudControllerClient.DeleteServiceCallCount()).To(Equal(1))

				serviceOfferingBeingPurged, purge := fakeCloudControllerClient.DeleteServiceArgsForCall(0)
				Expect(serviceOfferingBeingPurged).To(Equal("some-service-guid"))
				Expect(purge).To(BeTrue())
			})
		})

		When("purging the service fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteServiceReturns(
					ccv2.Warnings{"delete-service-warning"},
					fmt.Errorf("it didn't work"),
				)
			})

			It("should return the error and any warnings", func() {
				Expect(purgeErr).To(MatchError(fmt.Errorf("it didn't work")))
				Expect(warnings).To(ConsistOf("delete-service-warning"))
			})
		})
	})
})
