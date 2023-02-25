package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Details Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetServiceInstanceDetails", func() {
		const (
			serviceInstanceName          = "some-service-instance-name"
			serviceInstanceGUID          = "some-service-instance-guid"
			spaceGUID                    = "some-source-space-guid"
			spaceName                    = "some-source-space-name"
			orgName                      = "some-source-org-name"
			servicePlanName              = "fake-service-plan-name"
			servicePlanGUID              = "fake-service-plan-guid"
			serviceOfferingName          = "fake-service-offering-name"
			serviceOfferingGUID          = "fake-service-offering-guid"
			serviceOfferingDescription   = "some-service-description"
			serviceOfferingDocumentation = "some-service-documentation-url"
			serviceBrokerName            = "fake-service-broker-name"
		)

		var (
			serviceInstance ServiceInstanceDetails
			warnings        Warnings
			executionError  error
			omitApps        bool
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Type:      resources.ManagedServiceInstance,
					Name:      serviceInstanceName,
					GUID:      serviceInstanceGUID,
					SpaceGUID: spaceGUID,
				},
				ccv3.IncludedResources{
					ServicePlans: []resources.ServicePlan{{Name: servicePlanName}},
					ServiceOfferings: []resources.ServiceOffering{{
						Name:             serviceOfferingName,
						GUID:             serviceOfferingGUID,
						Description:      serviceOfferingDescription,
						DocumentationURL: serviceOfferingDocumentation,
						Tags:             types.NewOptionalStringSlice("foo", "bar"),
					}},
					ServiceBrokers: []resources.ServiceBroker{{Name: serviceBrokerName}},
					Spaces:         []resources.Space{{Name: spaceName}},
					Organizations:  []resources.Organization{{Name: orgName}},
				},
				ccv3.Warnings{"some-service-instance-warning"},
				nil,
			)

			fakeCloudControllerClient.GetFeatureFlagReturns(
				resources.FeatureFlag{Name: "service_instance_sharing", Enabled: true},
				ccv3.Warnings{},
				nil,
			)

			fakeCloudControllerClient.GetServiceOfferingByGUIDReturns(
				resources.ServiceOffering{Name: "offering-name", AllowsInstanceSharing: true},
				ccv3.Warnings{},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{
						Type:    resources.AppBinding,
						Name:    "binding-1",
						AppName: "app-1",
						LastOperation: resources.LastOperation{
							Type:        resources.CreateOperation,
							State:       resources.OperationSucceeded,
							Description: "all ok",
						},
					},
					{
						Type:    resources.AppBinding,
						Name:    "binding-2",
						AppName: "app-2",
						LastOperation: resources.LastOperation{
							Type:        resources.UpdateOperation,
							State:       resources.OperationInProgress,
							Description: "please wait",
						},
					},
				},
				ccv3.Warnings{"some-bindings-warning"},
				nil,
			)

			omitApps = false
		})

		JustBeforeEach(func() {
			serviceInstance, warnings, executionError = actor.GetServiceInstanceDetails(serviceInstanceName, spaceGUID, omitApps)
		})

		It("returns warnings and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf(
				"some-service-instance-warning",
				"some-bindings-warning",
			))
		})

		It("returns the correct service instance details object", func() {
			Expect(serviceInstance).To(Equal(
				ServiceInstanceDetails{
					ServiceInstance: resources.ServiceInstance{
						Type:      resources.ManagedServiceInstance,
						Name:      serviceInstanceName,
						GUID:      serviceInstanceGUID,
						SpaceGUID: spaceGUID,
					},
					SpaceName:        spaceName,
					OrganizationName: orgName,
					ServiceOffering: resources.ServiceOffering{
						Name:             serviceOfferingName,
						GUID:             serviceOfferingGUID,
						Description:      serviceOfferingDescription,
						DocumentationURL: serviceOfferingDocumentation,
						Tags:             types.NewOptionalStringSlice("foo", "bar"),
					},
					ServicePlan:       resources.ServicePlan{Name: servicePlanName},
					ServiceBrokerName: serviceBrokerName,
					SharedStatus:      SharedStatus{},
					BoundApps: []resources.ServiceCredentialBinding{
						{
							Type:    resources.AppBinding,
							Name:    "binding-1",
							AppName: "app-1",
							LastOperation: resources.LastOperation{
								Type:        resources.CreateOperation,
								State:       resources.OperationSucceeded,
								Description: "all ok",
							},
						},
						{
							Type:    resources.AppBinding,
							Name:    "binding-2",
							AppName: "app-2",
							LastOperation: resources.LastOperation{
								Type:        resources.UpdateOperation,
								State:       resources.OperationInProgress,
								Description: "please wait",
							},
						},
					},
				},
			))
		})

		Describe("getting the service instance", func() {
			It("makes the correct call to get the service instance", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.FieldsServicePlan,
						Values: []string{"name", "guid"},
					},
					ccv3.Query{
						Key:    ccv3.FieldsServicePlanServiceOffering,
						Values: []string{"name", "guid", "description", "tags", "documentation_url"},
					},
					ccv3.Query{
						Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
						Values: []string{"name", "guid"},
					},
					ccv3.Query{
						Key:    ccv3.FieldsSpace,
						Values: []string{"name", "guid"},
					},
					ccv3.Query{
						Key:    ccv3.FieldsSpaceOrganization,
						Values: []string{"name", "guid"},
					},
				))
			})

			When("the service instance cannot be found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{},
						ccerror.ServiceInstanceNotFoundError{},
					)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})

				It("does not attempt any other requests", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceOfferingByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
				})
			})

			When("getting the service instance returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"some-service-instance-warning"},
						errors.New("no service instance"))
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError("no service instance"))
					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				})

				It("does not attempt any other requests", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceOfferingByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
				})
			})
		})

		Describe("sharing", func() {
			When("targeting originating service instance space", func() {
				When("the service instance has shared spaces", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]ccv3.SpaceWithOrganization{{SpaceGUID: "some-other-space-guid", SpaceName: "some-other-space-name", OrganizationName: "some-org-name"}},
							ccv3.Warnings{},
							nil,
						)
						fakeCloudControllerClient.GetServiceInstanceUsageSummaryReturns(
							[]resources.ServiceInstanceUsageSummary{{SpaceGUID: "some-other-space-guid", BoundAppCount: 3}},
							ccv3.Warnings{},
							nil,
						)

					})
					It("Passes the right parameters to the client", func() {
						Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(Equal(1))
						serviceInstanceGUID := fakeCloudControllerClient.GetServiceInstanceSharedSpacesArgsForCall(0)
						Expect(serviceInstanceGUID).To(Equal(serviceInstance.GUID))

						Expect(fakeCloudControllerClient.GetServiceInstanceUsageSummaryCallCount()).To(Equal(1))
						serviceInstanceGUID = fakeCloudControllerClient.GetServiceInstanceUsageSummaryArgsForCall(0)
						Expect(serviceInstanceGUID).To(Equal(serviceInstance.GUID))
					})

					It("returns a service with a SharedStatus of IsSharedToOtherSpaces: true", func() {
						Expect(serviceInstance.SharedStatus.IsSharedToOtherSpaces).To(BeTrue())
					})

					It("returns a service instance with the usage summary", func() {
						Expect(serviceInstance.SharedStatus.UsageSummary).To(Equal([]UsageSummaryWithSpaceAndOrg{
							{
								SpaceName:        "some-other-space-name",
								OrganizationName: "some-org-name",
								BoundAppCount:    3,
							},
						}))
					})

				})

				When("the service instance does not have shared spaces", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]ccv3.SpaceWithOrganization{},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns a service with a SharedStatus of IsSharedToOtherSpaces: false", func() {
						Expect(serviceInstance.SharedStatus.FeatureFlagIsDisabled).To(BeFalse())
					})

					It("does not retrieve the usage summary", func() {
						Expect(fakeCloudControllerClient.GetServiceInstanceUsageSummaryCallCount()).To(Equal(0))
						Expect(serviceInstance.SharedStatus.UsageSummary).To(BeNil())
					})

				})

				When("the service sharing feature flag is disabled", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetFeatureFlagReturns(
							resources.FeatureFlag{Name: "service_instance_sharing", Enabled: false},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns service is not shared and feature flag info", func() {
						Expect(serviceInstance.SharedStatus.FeatureFlagIsDisabled).To(BeTrue())
					})
				})

				When("the service sharing feature flag is enabled", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetFeatureFlagReturns(
							resources.FeatureFlag{Name: "service_instance_sharing", Enabled: true},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns service is not shared and feature flag info", func() {
						Expect(serviceInstance.SharedStatus.FeatureFlagIsDisabled).To(BeFalse())
					})
				})

				When("the service offering does not allow sharing", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceOfferingByGUIDReturns(
							resources.ServiceOffering{Name: serviceOfferingName, AllowsInstanceSharing: false},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns service is not shared and feature flag info", func() {
						Expect(serviceInstance.SharedStatus.OfferingDisablesSharing).To(BeTrue())

						actualOfferingGUID := fakeCloudControllerClient.GetServiceOfferingByGUIDArgsForCall(0)
						Expect(actualOfferingGUID).To(Equal(serviceOfferingGUID))
					})
				})

				When("the service offering allows sharing", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
							resources.ServiceOffering{Name: serviceOfferingName, AllowsInstanceSharing: true},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns that the service broker does not disable sharing", func() {
						Expect(serviceInstance.SharedStatus.OfferingDisablesSharing).To(BeFalse())

						actualOfferingGUID := fakeCloudControllerClient.GetServiceOfferingByGUIDArgsForCall(0)
						Expect(actualOfferingGUID).To(Equal(serviceOfferingGUID))
					})
				})

				When("the service offering details can't be found", func() {
					const warningMessage = "some-offering-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
							resources.ServiceOffering{},
							ccv3.Warnings{warningMessage},
							ccerror.ServiceOfferingNotFoundError{},
						)
					})

					Context("because they can't actually share this service", func() {
						It("just returns that the service broker does not disable sharing and carries on", func() {
							Expect(serviceInstance.SharedStatus.OfferingDisablesSharing).To(BeFalse())
							Expect(executionError).NotTo(HaveOccurred())
						})
					})
				})

				When("getting feature flags errors", func() {
					const warningMessage = "some-feature-flag-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetFeatureFlagReturns(
							resources.FeatureFlag{},
							ccv3.Warnings{warningMessage},
							errors.New("error getting feature flag"),
						)
					})

					It("returns an empty service instance, warnings, and the error", func() {
						Expect(serviceInstance).To(Equal(ServiceInstanceDetails{}))
						Expect(executionError).To(MatchError("error getting feature flag"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", warningMessage))
					})
				})

				When("getting offering errors", func() {
					const warningMessage = "some-offering-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceOfferingByGUIDReturns(
							resources.ServiceOffering{},
							ccv3.Warnings{warningMessage},
							errors.New("error getting offering"),
						)
					})

					It("returns an empty service instance, warnings, and the error", func() {
						Expect(serviceInstance).To(Equal(ServiceInstanceDetails{}))
						Expect(executionError).To(MatchError("error getting offering"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", warningMessage))
					})
				})

				When("the fetching spaces returns new warnings", func() {
					const warningMessage = "some-shared-spaces-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]ccv3.SpaceWithOrganization{},
							ccv3.Warnings{warningMessage},
							nil,
						)
					})
					It("forwards those warnings on", func() {
						Expect(warnings).To(ContainElement(warningMessage))
					})
				})

				When("fetching shared spaces throws an error", func() {
					const warningMessage = "some-shared-spaces-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							nil,
							ccv3.Warnings{warningMessage},
							errors.New("no service instance"),
						)
					})

					It("returns an empty service instance, warnings, and the error", func() {
						Expect(serviceInstance).To(Equal(ServiceInstanceDetails{}))
						Expect(executionError).To(MatchError("no service instance"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", warningMessage))
					})
				})

				When("fetching usage summary throws an error", func() {
					const warningMessage = "some-usage-summary-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]ccv3.SpaceWithOrganization{{SpaceGUID: "some-other-space-guid", SpaceName: "some-other-space-name", OrganizationName: "some-org-name"}},
							ccv3.Warnings{},
							nil,
						)

						fakeCloudControllerClient.GetServiceInstanceUsageSummaryReturns(
							nil,
							ccv3.Warnings{warningMessage},
							errors.New("no service instance"),
						)
					})

					It("returns an empty service instance, warnings, and the error", func() {
						Expect(serviceInstance).To(Equal(ServiceInstanceDetails{}))
						Expect(executionError).To(MatchError("no service instance"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", warningMessage))
					})
				})

			})

			When("targeting space that service instance has been shared to", func() {
				BeforeEach(func() {
					originalSpace := "some-other-space"
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{
							Type:      resources.ManagedServiceInstance,
							Name:      serviceInstanceName,
							GUID:      serviceInstanceGUID,
							SpaceGUID: originalSpace,
						},
						ccv3.IncludedResources{
							ServicePlans: []resources.ServicePlan{{Name: servicePlanName}},
							ServiceOfferings: []resources.ServiceOffering{{
								Name:             serviceOfferingName,
								GUID:             serviceOfferingGUID,
								Description:      serviceOfferingDescription,
								DocumentationURL: serviceOfferingDocumentation,
								Tags:             types.NewOptionalStringSlice("foo", "bar"),
							}},
							ServiceBrokers: []resources.ServiceBroker{{Name: serviceBrokerName}},
							Spaces:         []resources.Space{{Name: spaceName}},
							Organizations:  []resources.Organization{{Name: orgName}},
						},
						ccv3.Warnings{"some-service-instance-warning"},
						nil,
					)
				})

				It("returns a service with a SharedStatus of IsSharedFromOriginalSpace: true", func() {
					Expect(serviceInstance.SharedStatus.IsSharedFromOriginalSpace).To(BeTrue())
				})

				It("does not retrieve more sharing information", func() {
					Expect(serviceInstance.SharedStatus.IsSharedToOtherSpaces).To(BeFalse())
					Expect(serviceInstance.SharedStatus.OfferingDisablesSharing).To(BeFalse())
					Expect(serviceInstance.SharedStatus.FeatureFlagIsDisabled).To(BeFalse())

					Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(BeZero())
					Expect(fakeCloudControllerClient.GetServiceOfferingByGUIDCallCount()).To(BeZero())
					Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(BeZero())
					Expect(fakeCloudControllerClient.GetServiceInstanceUsageSummaryCallCount()).To(BeZero())
				})
			})
		})

		Describe("upgrades", func() {
			When("upgrade is not available and the service does not have maintenance info", func() {
				It("says that upgrades are not supported", func() {
					Expect(serviceInstance.UpgradeStatus.State).To(Equal(ServiceInstanceUpgradeNotSupported))
				})

				It("does not get the service plan", func() {
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
				})
			})

			When("upgrade is not available but the service instance has maintenance info", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{
							Type:                   resources.ManagedServiceInstance,
							Name:                   serviceInstanceName,
							GUID:                   serviceInstanceGUID,
							MaintenanceInfoVersion: "1.2.3",
						},
						ccv3.IncludedResources{
							ServicePlans: []resources.ServicePlan{{Name: servicePlanName}},
							ServiceOfferings: []resources.ServiceOffering{{
								Name:             serviceOfferingName,
								GUID:             serviceOfferingGUID,
								Description:      serviceOfferingDescription,
								DocumentationURL: serviceOfferingDocumentation,
								Tags:             types.NewOptionalStringSlice("foo", "bar"),
							}},
							ServiceBrokers: []resources.ServiceBroker{{Name: serviceBrokerName}},
						},
						ccv3.Warnings{"some-service-instance-warning"},
						nil,
					)
				})

				It("says that an upgrade is not available", func() {
					Expect(serviceInstance.UpgradeStatus.State).To(Equal(ServiceInstanceUpgradeNotAvailable))
				})

				It("does not get the service plan", func() {
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
				})
			})

			When("an upgrade is available", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{
							Type:             resources.ManagedServiceInstance,
							Name:             serviceInstanceName,
							GUID:             serviceInstanceGUID,
							ServicePlanGUID:  servicePlanGUID,
							UpgradeAvailable: types.NewOptionalBoolean(true),
						},
						ccv3.IncludedResources{
							ServicePlans: []resources.ServicePlan{{Name: servicePlanName}},
							ServiceOfferings: []resources.ServiceOffering{{
								Name:             serviceOfferingName,
								GUID:             serviceOfferingGUID,
								Description:      serviceOfferingDescription,
								DocumentationURL: serviceOfferingDocumentation,
								Tags:             types.NewOptionalStringSlice("foo", "bar"),
							}},
							ServiceBrokers: []resources.ServiceBroker{{Name: serviceBrokerName}},
						},
						ccv3.Warnings{"some-service-instance-warning"},
						nil,
					)

					fakeCloudControllerClient.GetServicePlanByGUIDReturns(
						resources.ServicePlan{
							GUID:                       servicePlanGUID,
							Name:                       servicePlanName,
							MaintenanceInfoDescription: "requires downtime",
						},
						ccv3.Warnings{"service-plan-warning"},
						nil,
					)
				})

				It("gets the service plan", func() {
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDArgsForCall(0)).To(Equal(servicePlanGUID))

					Expect(warnings).To(ContainElement("service-plan-warning"))
				})

				It("says that an upgrade is available", func() {
					Expect(serviceInstance.UpgradeStatus).To(Equal(ServiceInstanceUpgradeStatus{
						State:       ServiceInstanceUpgradeAvailable,
						Description: "requires downtime",
					}))
				})

				When("the service plan cannot be found", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServicePlanByGUIDReturns(
							resources.ServicePlan{},
							ccv3.Warnings{"service-plan-warning"},
							ccerror.ServicePlanNotFound{},
						)
					})

					It("says that an upgrade is available without details", func() {
						Expect(warnings).To(ContainElement("service-plan-warning"))
						Expect(serviceInstance.UpgradeStatus).To(Equal(ServiceInstanceUpgradeStatus{
							State:       ServiceInstanceUpgradeAvailable,
							Description: "No upgrade details where found",
						}))
					})
				})

				When("getting the service plan fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServicePlanByGUIDReturns(
							resources.ServicePlan{},
							ccv3.Warnings{"service-plan-warning"},
							errors.New("boom"),
						)
					})

					It("says that an upgrade is available without details", func() {
						Expect(warnings).To(ContainElement("service-plan-warning"))
						Expect(executionError).To(MatchError("boom"))
					})
				})
			})
		})

		Describe("bindings", func() {
			It("makes the correct call to get bindings", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}},
					ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
				))
			})

			When("omitApps is true", func() {
				BeforeEach(func() {
					omitApps = true
				})

				It("does not get the bindings", func() {
					Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(BeZero())
				})

				It("does not include the bindings in the result", func() {
					Expect(serviceInstance.BoundApps).To(BeEmpty())
				})
			})

			When("there is an error getting the bindings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"some-bindings-warning"},
						errors.New("a bindings error"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(executionError).To(MatchError("a bindings error"))
					Expect(warnings).To(ContainElement("some-bindings-warning"))
				})
			})
		})
	})

	Describe("GetServiceInstanceParameters", func() {
		const (
			serviceInstanceName = "some-service-instance-name"
			serviceInstanceGUID = "some-service-instance-guid"
			spaceGUID           = "some-source-space-guid"
		)

		var (
			serviceInstanceParams ServiceInstanceParameters
			warnings              Warnings
			executionError        error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Type:      resources.ManagedServiceInstance,
					Name:      serviceInstanceName,
					GUID:      serviceInstanceGUID,
					SpaceGUID: spaceGUID,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"some-service-instance-warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceInstanceParametersReturns(
				map[string]interface{}{"foo": "bar"},
				ccv3.Warnings{"some-parameters-warning"},
				nil,
			)

		})

		JustBeforeEach(func() {
			serviceInstanceParams, warnings, executionError = actor.GetServiceInstanceParameters(serviceInstanceName, spaceGUID)
		})

		Describe("getting the service instance", func() {
			It("makes the correct call to get the service instance", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeNil())
			})

			When("the service instance cannot be found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})

				It("does not attempt any other requests", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceOfferingByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
				})
			})

			When("getting the service instance returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"some-service-instance-warning"},
						errors.New("no service instance"))
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError("no service instance"))
					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				})

				It("does not attempt any other requests", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetFeatureFlagCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceOfferingByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceInstanceSharedSpacesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
				})
			})
		})

		Describe("getting the parameters", func() {
			It("makes the correct call to get the parameters", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersArgsForCall(0)).To(Equal(serviceInstanceGUID))
			})

			It("returns parameters", func() {
				Expect(serviceInstanceParams).To(Equal(
					ServiceInstanceParameters(map[string]interface{}{"foo": "bar"})))
			})

			When("getting the parameters fails with a ServiceInstanceParametersFetchNotSupportedError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceParametersReturns(
						map[string]interface{}{},
						ccv3.Warnings{"some-parameters-warning"},
						ccerror.ServiceInstanceParametersFetchNotSupportedError{
							Message: "This service does not support fetching service instance parameters.",
						},
					)
				})

				It("does not return an error, but returns warnings and the reason", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceParamsFetchingNotSupportedError{}))
					Expect(warnings).To(ContainElement("some-parameters-warning"))
				})
			})

			When("getting the parameters fails with a ResourceNotFoundError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceParametersReturns(
						map[string]interface{}{},
						ccv3.Warnings{"some-parameters-warning"},
						ccerror.ResourceNotFoundError{
							Message: "Service instance not found",
						},
					)
				})

				It("converts the error to fetching not supported", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceParamsFetchingNotSupportedError{}))
					Expect(warnings).To(ContainElement("some-parameters-warning"))
				})
			})

			When("getting the parameters fails with an another error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceParametersReturns(
						map[string]interface{}{},
						ccv3.Warnings{"some-parameters-warning"},
						errors.New("not expected"),
					)
				})

				It("does not return an error, but returns warnings and the reason", func() {
					Expect(executionError).To(MatchError("not expected"))
					Expect(warnings).To(ContainElement("some-parameters-warning"))
				})
			})
		})
	})
})
