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

	Describe("GetServicesWithPlansForBroker", func() {
		When("the broker has no services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, nil, nil)
			})

			It("returns no services", func() {
				servicesWithPlans, _, err := actor.GetServicesWithPlansForBroker("some-broker-guid")
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
				servicesWithPlans, _, err := actor.GetServicesWithPlansForBroker("some-broker-guid")
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

			It("returns a single service with associated plans and warnings", func() {
				servicesWithPlans, warnings, err := actor.GetServicesWithPlansForBroker("some-broker-guid")
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
		})

		When("fetching services returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-service-warning"}, errors.New("EXPLODE"))
			})

			It("propagates the error and warnings", func() {
				_, warnings, err := actor.GetServicesWithPlansForBroker("some-broker-guid")
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
				_, warnings, err := actor.GetServicesWithPlansForBroker("some-broker-guid")
				Expect(err).To(MatchError("EXPLODE"))

				Expect(warnings).To(ConsistOf("get-service-warning", "get-plan-warning"))
			})
		})
	})
})
