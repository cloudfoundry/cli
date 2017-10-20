package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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

		Context("when the service plan exists", func() {
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

		Context("when an error is encountered getting the service plan", func() {
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
})
