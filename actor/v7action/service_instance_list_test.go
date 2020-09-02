package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance List Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetServiceInstancesForSpace", func() {
		const spaceGUID = "some-source-space-guid"

		var (
			serviceInstances []ServiceInstance
			warnings         Warnings
			executionError   error
		)

		JustBeforeEach(func() {
			serviceInstances, warnings, executionError = actor.GetServiceInstancesForSpace(spaceGUID)
		})

		It("makes the correct call to get service instances", func() {
			Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
				ccv3.Query{Key: ccv3.FieldsServicePlan, Values: []string{"guid", "name", "relationships.service_offering"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"guid", "name", "relationships.service_broker"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"guid", "name"}},
				ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
			))
		})

		When("the cloud controller request is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					[]resources.ServiceInstance{
						{
							GUID:             "fake-guid-1",
							Type:             resources.ManagedServiceInstance,
							Name:             "msi1",
							ServicePlanGUID:  "fake-plan-guid-1",
							UpgradeAvailable: types.NewOptionalBoolean(true),
							LastOperation: resources.LastOperation{
								Type:  resources.CreateOperation,
								State: resources.OperationSucceeded,
							},
						},
						{
							GUID:             "fake-guid-2",
							Type:             resources.ManagedServiceInstance,
							Name:             "msi2",
							ServicePlanGUID:  "fake-plan-guid-2",
							UpgradeAvailable: types.NewOptionalBoolean(false),
							LastOperation: resources.LastOperation{
								Type:  resources.UpdateOperation,
								State: resources.OperationSucceeded,
							},
						},
						{
							GUID:            "fake-guid-3",
							Type:            resources.ManagedServiceInstance,
							Name:            "msi3",
							ServicePlanGUID: "fake-plan-guid-3",
							LastOperation: resources.LastOperation{
								Type:  resources.CreateOperation,
								State: resources.OperationInProgress,
							},
						},
						{
							GUID:             "fake-guid-4",
							Type:             resources.ManagedServiceInstance,
							Name:             "msi4",
							ServicePlanGUID:  "fake-plan-guid-4",
							UpgradeAvailable: types.NewOptionalBoolean(true),
							LastOperation: resources.LastOperation{
								Type:  resources.CreateOperation,
								State: resources.OperationFailed,
							},
						},
						{
							GUID:             "fake-guid-5",
							Type:             resources.ManagedServiceInstance,
							Name:             "msi5",
							ServicePlanGUID:  "fake-plan-guid-4",
							UpgradeAvailable: types.NewOptionalBoolean(false),
							LastOperation: resources.LastOperation{
								Type:  resources.DeleteOperation,
								State: resources.OperationInProgress,
							},
						},
						{
							GUID: "fake-guid-6",
							Type: resources.UserProvidedServiceInstance,
							Name: "upsi",
						},
					},
					ccv3.IncludedResources{
						ServicePlans: []resources.ServicePlan{
							{
								GUID:                "fake-plan-guid-1",
								Name:                "fake-plan-1",
								ServiceOfferingGUID: "fake-offering-guid-1",
							},
							{
								GUID:                "fake-plan-guid-2",
								Name:                "fake-plan-2",
								ServiceOfferingGUID: "fake-offering-guid-2",
							},
							{
								GUID:                "fake-plan-guid-3",
								Name:                "fake-plan-3",
								ServiceOfferingGUID: "fake-offering-guid-3",
							},
							{
								GUID:                "fake-plan-guid-4",
								Name:                "fake-plan-4",
								ServiceOfferingGUID: "fake-offering-guid-3",
							},
						},
						ServiceOfferings: []resources.ServiceOffering{
							{
								GUID:              "fake-offering-guid-1",
								Name:              "fake-offering-1",
								ServiceBrokerGUID: "fake-broker-guid-1",
							},
							{
								GUID:              "fake-offering-guid-2",
								Name:              "fake-offering-2",
								ServiceBrokerGUID: "fake-broker-guid-2",
							},
							{
								GUID:              "fake-offering-guid-3",
								Name:              "fake-offering-3",
								ServiceBrokerGUID: "fake-broker-guid-2",
							},
						},
						ServiceBrokers: []resources.ServiceBroker{
							{
								GUID: "fake-broker-guid-1",
								Name: "fake-broker-1",
							},
							{
								GUID: "fake-broker-guid-2",
								Name: "fake-broker-2",
							},
						},
					},
					ccv3.Warnings{"a warning"},
					nil,
				)
			})

			It("returns a list of service instances and warnings", func() {
				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("a warning"))

				Expect(serviceInstances).To(Equal([]ServiceInstance{
					{
						Name:                "msi1",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-1",
						ServiceOfferingName: "fake-offering-1",
						ServiceBrokerName:   "fake-broker-1",
						UpgradeAvailable:    types.NewOptionalBoolean(true),
						BoundApps:           []string{"foo", "bar"},
						LastOperation:       "create succeeded",
					},
					{
						Name:                "msi2",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-2",
						ServiceOfferingName: "fake-offering-2",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(false),
						BoundApps:           []string{"foo", "bar"},
						LastOperation:       "update succeeded",
					},
					{
						Name:                "msi3",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-3",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						BoundApps:           []string{"foo", "bar"},
						LastOperation:       "create in progress",
					},
					{
						Name:                "msi4",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-4",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(true),
						BoundApps:           []string{"foo", "bar"},
						LastOperation:       "create failed",
					},
					{
						Name:                "msi5",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-4",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(false),
						BoundApps:           []string{"foo", "bar"},
						LastOperation:       "delete in progress",
					},
					{
						Name:      "upsi",
						Type:      resources.UserProvidedServiceInstance,
						BoundApps: []string{"foo", "bar"},
					},
				}))
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					[]resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("something really awful"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("something really awful"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})
})
