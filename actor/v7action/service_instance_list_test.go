package v7action_test

import (
	"errors"
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/batcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance List Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)
	const spaceGUID = "some-source-space-guid"

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
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
		fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
			[]resources.ServiceCredentialBinding{
				{Type: "app", ServiceInstanceGUID: "fake-guid-1", AppGUID: "app-1", AppName: "great-app-1", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-1", AppGUID: "app-2", AppName: "great-app-2", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-2", AppGUID: "app-3", AppName: "great-app-3", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-2", AppGUID: "app-4", AppName: "great-app-4", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-3", AppGUID: "app-5", AppName: "great-app-5", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-4", AppGUID: "app-6", AppName: "great-app-6", AppSpaceGUID: spaceGUID},
				{Type: "app", ServiceInstanceGUID: "fake-guid-5", AppGUID: "app-6", AppName: "great-app-6", AppSpaceGUID: spaceGUID},
				{Type: "key", ServiceInstanceGUID: "fake-guid-1", AppSpaceGUID: spaceGUID},
				{Type: "key", ServiceInstanceGUID: "fake-guid-2", AppSpaceGUID: spaceGUID},
			},
			ccv3.Warnings{"bindings warning"},
			nil,
		)
	})

	Describe("GetServiceInstancesForSpace", func() {
		var (
			serviceInstances []ServiceInstance
			warnings         Warnings
			executionError   error
			omitApps         bool
		)

		BeforeEach(func() {
			omitApps = false
		})

		JustBeforeEach(func() {
			serviceInstances, warnings, executionError = actor.GetServiceInstancesForSpace(spaceGUID, omitApps)
		})

		It("makes the correct call to get service instances", func() {
			Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
				ccv3.Query{Key: ccv3.FieldsServicePlan, Values: []string{"guid", "name", "relationships.service_offering"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"guid", "name", "relationships.service_broker"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"guid", "name"}},
				ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
			))
		})

		It("makes the correct call to get service credential bindings", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{
					"fake-guid-1",
					"fake-guid-2",
					"fake-guid-3",
					"fake-guid-4",
					"fake-guid-5",
					"fake-guid-6",
				}},
				ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
			))
		})

		When("omit apps is set to true", func() {
			BeforeEach(func() {
				omitApps = true
			})

			It("does not get service credential bindings", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
			})
		})

		When("the list of service instances is long", func() {
			BeforeEach(func() {
				var longList []resources.ServiceInstance
				for i := 0; i < batcher.BatchSize*10; i++ {
					longList = append(longList, resources.ServiceInstance{GUID: fmt.Sprintf("fake-guid-%d", i)})
				}
				fakeCloudControllerClient.GetServiceInstancesReturns(
					longList,
					ccv3.IncludedResources{},
					ccv3.Warnings{},
					nil,
				)
			})

			It("makes multiple different requests to get the service credential bindings", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(10))
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).
					NotTo(Equal(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(1)))
			})
		})

		When("the cloud controller request is successful", func() {
			It("returns a list of service instances and warnings", func() {
				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("a warning", "bindings warning"))

				Expect(serviceInstances).To(Equal([]ServiceInstance{
					{
						Name:                "msi1",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-1",
						ServiceOfferingName: "fake-offering-1",
						ServiceBrokerName:   "fake-broker-1",
						UpgradeAvailable:    types.NewOptionalBoolean(true),
						BoundApps:           []string{"great-app-1", "great-app-2"},
						LastOperation:       "create succeeded",
					},
					{
						Name:                "msi2",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-2",
						ServiceOfferingName: "fake-offering-2",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(false),
						BoundApps:           []string{"great-app-3", "great-app-4"},
						LastOperation:       "update succeeded",
					},
					{
						Name:                "msi3",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-3",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						BoundApps:           []string{"great-app-5"},
						LastOperation:       "create in progress",
					},
					{
						Name:                "msi4",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-4",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(true),
						BoundApps:           []string{"great-app-6"},
						LastOperation:       "create failed",
					},
					{
						Name:                "msi5",
						Type:                resources.ManagedServiceInstance,
						ServicePlanName:     "fake-plan-4",
						ServiceOfferingName: "fake-offering-3",
						ServiceBrokerName:   "fake-broker-2",
						UpgradeAvailable:    types.NewOptionalBoolean(false),
						BoundApps:           []string{"great-app-6"},
						LastOperation:       "delete in progress",
					},
					{
						Name:      "upsi",
						Type:      resources.UserProvidedServiceInstance,
						BoundApps: nil,
					},
				}))
			})
		})

		When("the getting the service instances returns an error", func() {
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

		When("the getting the service credential bindings returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"some-service-credential-binding-warning"},
					errors.New("something really REALLY awful"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("something really REALLY awful"))
				Expect(warnings).To(ConsistOf("a warning", "some-service-credential-binding-warning"))
			})
		})

		When("the service instance is shared", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					[]resources.ServiceInstance{{
						GUID:             "service-instance-guid",
						Type:             resources.ManagedServiceInstance,
						Name:             "msi1",
						ServicePlanGUID:  "fake-plan-guid-1",
						UpgradeAvailable: types.NewOptionalBoolean(true),
						LastOperation: resources.LastOperation{
							Type:  resources.CreateOperation,
							State: resources.OperationSucceeded,
						},
					}},
					ccv3.IncludedResources{
						ServicePlans: []resources.ServicePlan{{
							GUID:                "fake-plan-guid-1",
							Name:                "fake-plan-1",
							ServiceOfferingGUID: "fake-offering-guid-1",
						}},
						ServiceOfferings: []resources.ServiceOffering{{
							GUID:              "fake-offering-guid-1",
							Name:              "fake-offering-1",
							ServiceBrokerGUID: "fake-broker-guid-1",
						}},
						ServiceBrokers: []resources.ServiceBroker{{
							GUID: "fake-broker-guid-1", Name: "fake-broker-1",
						}},
					},
					ccv3.Warnings{"a warning"},
					nil,
				)

				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{{
						Type:                "app",
						GUID:                "some-binding-guid",
						ServiceInstanceGUID: "service-instance-guid",
						AppGUID:             "app-guid",
						AppName:             "app-on-space-1",
						AppSpaceGUID:        spaceGUID,
					}, {
						Type:                "app",
						GUID:                "some-binding-guid",
						ServiceInstanceGUID: "service-instance-guid",
						AppGUID:             "app-guid-2",
						AppName:             "app-on-space-2",
						AppSpaceGUID:        "another-space-guid",
					}},
					nil,
					nil,
				)
			})

			It("should only include bound apps in the provided space", func() {
				Expect(serviceInstances).To(HaveLen(1))
				serviceInstance := serviceInstances[0]
				Expect(serviceInstance.BoundApps).To(ConsistOf("app-on-space-1"))
			})
		})
	})
})
