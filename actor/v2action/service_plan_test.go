package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plan Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServicePlan", func() {
		var (
			servicePlan         ServicePlan
			servicePlanWarnings Warnings
			servicePlanErr      error
		)

		JustBeforeEach(func() {
			servicePlan, servicePlanWarnings, servicePlanErr = actor.GetServicePlan("some-service-plan-guid")
		})

		When("the service plan exists", func() {
			var returnedServicePlan ccv2.ServicePlan

			BeforeEach(func() {
				returnedServicePlan = ccv2.ServicePlan{
					GUID: "some-service-plan-guid",
					Name: "some-service-plan",
				}
				fakeCloudControllerClient.GetServicePlanReturns(
					returnedServicePlan,
					ccv2.Warnings{"get-service-plan-warning"},
					nil)
			})

			It("returns the service plan and all warnings", func() {
				Expect(servicePlanErr).ToNot(HaveOccurred())
				Expect(servicePlan).To(Equal(ServicePlan(returnedServicePlan)))
				Expect(servicePlanWarnings).To(ConsistOf("get-service-plan-warning"))

				Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlanArgsForCall(0)).To(Equal("some-service-plan-guid"))
			})
		})

		When("an error is encountered getting the service plan", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeCloudControllerClient.GetServicePlanReturns(
					ccv2.ServicePlan{},
					ccv2.Warnings{"get-service-plan-warning"},
					expectedErr)
			})

			It("returns the errors and all warnings", func() {
				Expect(servicePlanErr).To(MatchError(expectedErr))
				Expect(servicePlan).To(Equal(ServicePlan{}))
				Expect(servicePlanWarnings).To(ConsistOf("get-service-plan-warning"))

				Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlanArgsForCall(0)).To(Equal("some-service-plan-guid"))
			})
		})
	})

	Describe("GetServicePlansForService", func() {
		var (
			servicePlans        []ServicePlan
			servicePlanWarnings Warnings
			servicePlanErr      error
		)

		JustBeforeEach(func() {
			servicePlans, servicePlanWarnings, servicePlanErr = actor.GetServicePlansForService("some-service")
		})

		When("there is a service with plans", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:        "some-service-guid",
						Label:       "some-service",
						Description: "service-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a"},
					{Name: "plan-b"},
				}

				fakeCloudControllerClient.GetServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("returns all plans and warnings", func() {
				Expect(servicePlanErr).NotTo(HaveOccurred())
				Expect(servicePlans).To(Equal([]ServicePlan{
					{
						Name: "plan-a",
					},
					{
						Name: "plan-b",
					},
				},
				))
				Expect(servicePlanWarnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(Equal([]ccv2.Filter{{
					Type:     constant.ServiceGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service-guid"},
				}}))
			})
		})

		When("GetServices returns an error and warnings", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil, ccv2.Warnings{"get-service-warning"}, errors.New("service-error"))
			})

			It("propagates the error and all warnings", func() {
				Expect(servicePlanErr).To(MatchError(errors.New("service-error")))
				Expect(servicePlanWarnings).To(ConsistOf("get-service-warning"))
			})
		})

		When("GetServicePlans returns an error and warnings", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{{}}, ccv2.Warnings{"get-service-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(nil, ccv2.Warnings{"get-plans-warning"}, errors.New("plans-error"))
			})

			It("propagates the error and all warnings", func() {
				Expect(servicePlanErr).To(MatchError(errors.New("plans-error")))
				Expect(servicePlanWarnings).To(ConsistOf("get-service-warning", "get-plans-warning"))
			})
		})
	})
})
