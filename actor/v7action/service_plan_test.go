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

var _ = Describe("Service Plan Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
	})

	Describe("GetServicePlanByNameOfferingAndBroker", func() {
		const (
			servicePlanName = "myServicePlan"
		)

		var (
			servicePlan         ServicePlan
			warnings            Warnings
			executionError      error
			serviceBrokerName   string
			serviceOfferingName string
		)

		BeforeEach(func() {
			serviceBrokerName = ""
			serviceOfferingName = ""
		})

		JustBeforeEach(func() {
			servicePlan, warnings, executionError = actor.GetServicePlanByNameOfferingAndBroker(
				servicePlanName,
				serviceOfferingName,
				serviceBrokerName,
			)
		})

		When("the cloud controller request is successful", func() {
			When("the cloud controller returns one service plan", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{
						{
							Name: "some-service-plan",
							GUID: "some-service-plan-guid",
						},
					}, ccv3.Warnings{"some-service-plan-warning"}, nil)
				})

				It("returns a service plan and warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(servicePlan).To(Equal(ServicePlan{Name: "some-service-plan", GUID: "some-service-plan-guid"}))
					Expect(warnings).To(ConsistOf("some-service-plan-warning"))
					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{servicePlanName}},
					))
				})
			})

			When("the cloud controller returns no service plans", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						nil,
						ccv3.Warnings{"some-service-plan-warning"},
						nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServicePlanNotFoundError{
						PlanName:    servicePlanName,
						ServiceName: serviceOfferingName,
					}))
					Expect(warnings).To(ConsistOf("some-service-plan-warning"))
				})
			})

			When("the cloud controller returns more than one service plan", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{
						{
							Name: "some-service-plan-1",
							GUID: "some-service-plan-guid-1",
						},
						{
							Name: "some-service-plan-2",
							GUID: "some-service-plan-guid-2",
						},
					}, ccv3.Warnings{"some-service-plan-warning"}, nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.DuplicateServicePlanError{Name: servicePlanName}))
					Expect(warnings).To(ConsistOf("some-service-plan-warning"))
				})
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					nil,
					ccv3.Warnings{"some-service-plan-warning"},
					errors.New("no service plan"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service plan"))
				Expect(warnings).To(ConsistOf("some-service-plan-warning"))
			})
		})

		When("the offering name is provided", func() {
			BeforeEach(func() {
				serviceOfferingName = "some-offering-name"
				fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{
					{
						Name: "some-service-plan",
						GUID: "some-service-plan-guid",
					},
					{
						Name: "some-service-plan",
						GUID: "some-other-service-plan-guid",
					},
				}, ccv3.Warnings{"some-service-plan-warning"}, nil)
			})

			It("queries only on the service plan and offering name", func() {
				Expect(warnings).To(ConsistOf("some-service-plan-warning"))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{servicePlanName}},
					ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{serviceOfferingName}},
				))
			})

			When("It returns multiple results", func() {
				It("returns an error with the right hint", func() {
					Expect(executionError).To(MatchError(actionerror.DuplicateServicePlanError{Name: servicePlanName, ServiceOfferingName: serviceOfferingName}))
				})
			})

		})

		When("the broker name is provided", func() {
			BeforeEach(func() {
				serviceBrokerName = "some-broker-name"
				fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{
					{
						Name: "some-service-plan",
						GUID: "some-service-plan-guid",
					},
					{
						Name: "some-service-plan",
						GUID: "other-some-service-plan-guid",
					},
				}, ccv3.Warnings{"some-service-plan-warning"}, nil)
			})

			It("queries only on the service plan name", func() {
				Expect(warnings).To(ConsistOf("some-service-plan-warning"))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{servicePlanName}},
					ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}},
				))
			})

			When("It returns multiple results", func() {
				It("returns an error with the right hint", func() {
					Expect(executionError).To(MatchError(actionerror.DuplicateServicePlanError{Name: servicePlanName, ServiceBrokerName: serviceBrokerName}))
				})
			})

		})
	})
})
