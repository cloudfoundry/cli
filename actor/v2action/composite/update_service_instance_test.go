package composite_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2action/composite"
	"code.cloudfoundry.org/cli/actor/v2action/composite/compositefakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateServiceInstanceCompositeActor", func() {
	var (
		composite                                     *UpdateServiceInstanceCompositeActor
		fakeGetServiceInstanceActor                   *compositefakes.FakeGetServiceInstanceActor
		fakeGetServicePlanActor                       *compositefakes.FakeGetServicePlanActor
		fakeUpdateServiceInstanceMaintenanceInfoActor *compositefakes.FakeUpdateServiceInstanceMaintenanceInfoActor
		err                                           error
		warnings                                      v2action.Warnings
	)

	BeforeEach(func() {
		fakeGetServiceInstanceActor = new(compositefakes.FakeGetServiceInstanceActor)
		fakeGetServicePlanActor = new(compositefakes.FakeGetServicePlanActor)
		fakeUpdateServiceInstanceMaintenanceInfoActor = new(compositefakes.FakeUpdateServiceInstanceMaintenanceInfoActor)
		composite = &UpdateServiceInstanceCompositeActor{
			GetServiceInstanceActor:                   fakeGetServiceInstanceActor,
			GetServicePlanActor:                       fakeGetServicePlanActor,
			UpdateServiceInstanceMaintenanceInfoActor: fakeUpdateServiceInstanceMaintenanceInfoActor,
		}
	})

	Describe("UpgradeServiceInstance", func() {
		const (
			serviceInstanceGUID = "service-instance-guid"
			servicePlanGUID     = "service-plan-guid"
		)

		JustBeforeEach(func() {
			warnings, err = composite.UpgradeServiceInstance(serviceInstanceGUID, servicePlanGUID)
		})

		When("the plan exists", func() {
			var maintenanceInfo v2action.MaintenanceInfo

			BeforeEach(func() {
				maintenanceInfo = v2action.MaintenanceInfo{
					Version: "1.2.3-abc",
				}
				servicePlan := v2action.ServicePlan{
					MaintenanceInfo: ccv2.MaintenanceInfo(maintenanceInfo),
				}
				fakeGetServicePlanActor.GetServicePlanReturns(servicePlan, v2action.Warnings{"plan-lookup-warning"}, nil)
				fakeUpdateServiceInstanceMaintenanceInfoActor.UpdateServiceInstanceMaintenanceInfoReturns(v2action.Warnings{"update-service-instance-warning"}, nil)
			})

			It("updates the service instance with the latest maintenanceInfo on the plan", func() {
				Expect(err).To(BeNil())
				Expect(fakeUpdateServiceInstanceMaintenanceInfoActor.UpdateServiceInstanceMaintenanceInfoCallCount()).To(Equal(1))
				guid, minfo := fakeUpdateServiceInstanceMaintenanceInfoActor.UpdateServiceInstanceMaintenanceInfoArgsForCall(0)
				Expect(guid).To(Equal(serviceInstanceGUID))
				Expect(minfo).To(Equal(maintenanceInfo))

				Expect(fakeGetServicePlanActor.GetServicePlanCallCount()).To(Equal(1))
				planGUID := fakeGetServicePlanActor.GetServicePlanArgsForCall(0)
				Expect(planGUID).To(Equal(servicePlanGUID))
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf("plan-lookup-warning", "update-service-instance-warning"))
			})

			When("updating the service instance fails", func() {
				BeforeEach(func() {
					fakeUpdateServiceInstanceMaintenanceInfoActor.UpdateServiceInstanceMaintenanceInfoReturns(
						v2action.Warnings{"update-service-instance-warning"},
						errors.New("something really bad happened"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(MatchError("something really bad happened"))
					Expect(warnings).To(ConsistOf("plan-lookup-warning", "update-service-instance-warning"))
				})
			})
		})

		When("fetching the plan fails", func() {
			BeforeEach(func() {
				fakeGetServicePlanActor.GetServicePlanReturns(
					v2action.ServicePlan{},
					v2action.Warnings{"plan-lookup-warning"},
					errors.New("something really bad happened"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(err).To(MatchError("something really bad happened"))
				Expect(warnings).To(ConsistOf("plan-lookup-warning"))
			})
		})
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		var serviceInstance v2action.ServiceInstance

		JustBeforeEach(func() {
			serviceInstance, warnings, err = composite.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
		})

		When("the service instance exists", func() {
			BeforeEach(func() {
				fakeGetServiceInstanceActor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID: "some-service-instance-guid",
						Name: "some-service-instance",
					},
					v2action.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service instance and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(v2action.ServiceInstance{
					GUID: "some-service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(ConsistOf("foo"))

				Expect(fakeGetServiceInstanceActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))

				serviceInstanceGUID, spaceGUID := fakeGetServiceInstanceActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(serviceInstanceGUID).To(Equal("some-service-instance"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		When("there is an error getting the service instance", func() {
			BeforeEach(func() {
				fakeGetServiceInstanceActor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"foo"},
					errors.New("something really bad happened"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(err).To(MatchError("something really bad happened"))
				Expect(warnings).To(ConsistOf("foo"))
			})
		})
	})
})
