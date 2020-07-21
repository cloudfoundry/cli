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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		const (
			serviceInstanceName = "some-service-instance"
			spaceGUID           = "some-source-space-guid"
		)

		var (
			serviceInstance resources.ServiceInstance
			warnings        Warnings
			executionError  error
		)

		JustBeforeEach(func() {
			serviceInstance, warnings, executionError = actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
		})

		When("the cloud controller request is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(resources.ServiceInstance{
					Name: "some-service-instance",
					GUID: "some-service-instance-guid",
				}, ccv3.IncludedResources{}, ccv3.Warnings{"some-service-instance-warning"}, nil)
			})

			It("returns a service instance and warnings", func() {
				Expect(executionError).NotTo(HaveOccurred())

				Expect(serviceInstance).To(Equal(resources.ServiceInstance{Name: "some-service-instance", GUID: "some-service-instance-guid"}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})
		})

		When("the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName, SpaceGUID: spaceGUID},
				)
			})

			It("returns an actor error and warnings", func() {
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("no service instance"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service instance"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})

	Describe("GetServiceInstanceDetails", func() {
		const (
			serviceInstanceName          = "some-service-instance-name"
			serviceInstanceGUID          = "some-service-instance-guid"
			spaceGUID                    = "some-source-space-guid"
			servicePlanName              = "fake-service-plan-name"
			serviceOfferingName          = "fake-service-offering-name"
			serviceOfferingDescription   = "some-service-description"
			serviceOfferingDocumentation = "some-service-documentation-url"
			serviceBrokerName            = "fake-service-broker-name"
		)

		var (
			serviceInstance ServiceInstanceDetails
			warnings        Warnings
			executionError  error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetFeatureFlagReturns(
				resources.FeatureFlag{Name: "service_instance_sharing", Enabled: true},
				ccv3.Warnings{},
				nil,
			)

			fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
				resources.ServiceOffering{Name: "offering-name", AllowsInstanceSharing: true},
				ccv3.Warnings{},
				nil,
			)
		})

		JustBeforeEach(func() {
			serviceInstance, warnings, executionError = actor.GetServiceInstanceDetails(serviceInstanceName, spaceGUID)
		})

		When("the cloud controller request is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.ManagedServiceInstance,
						Name: serviceInstanceName,
						GUID: serviceInstanceGUID,
					},
					ccv3.IncludedResources{
						ServicePlans: []resources.ServicePlan{{Name: servicePlanName}},
						ServiceOfferings: []resources.ServiceOffering{{
							Name:             serviceOfferingName,
							Description:      serviceOfferingDescription,
							DocumentationURL: serviceOfferingDocumentation,
							Tags:             types.NewOptionalStringSlice("foo", "bar"),
						}},
						ServiceBrokers: []resources.ServiceBroker{{Name: serviceBrokerName}},
					},
					ccv3.Warnings{"some-service-instance-warning"},
					nil,
				)

				fakeCloudControllerClient.GetServiceInstanceParametersReturns(
					types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
					ccv3.Warnings{"some-parameters-warning"},
					nil,
				)
			})

			It("returns warnings and no errors", func() {
				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"some-service-instance-warning",
					"some-parameters-warning",
				))
			})

			It("returns the correct service instance details object", func() {
				Expect(serviceInstance).To(Equal(
					ServiceInstanceDetails{
						ServiceInstance: resources.ServiceInstance{
							Type: resources.ManagedServiceInstance,
							Name: serviceInstanceName,
							GUID: serviceInstanceGUID,
						},
						ServiceOffering: resources.ServiceOffering{
							Name:             serviceOfferingName,
							Description:      serviceOfferingDescription,
							DocumentationURL: serviceOfferingDocumentation,
							Tags:             types.NewOptionalStringSlice("foo", "bar"),
						},
						ServicePlanName:         servicePlanName,
						ServiceBrokerName:       serviceBrokerName,
						SharedStatus:            SharedStatus{},
						Parameters:              types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
						ParametersMissingReason: "",
					},
				))
			})

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
				))
			})

			It("makes the correct call to get the parameters", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersArgsForCall(0)).To(Equal(serviceInstanceGUID))
			})

			When("the service instance is managed", func() {
				When("the service instance has shared spaces", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]resources.Space{{GUID: "some-other-space-guid"}},
							ccv3.Warnings{},
							nil,
						)
					})
					It("returns a service with a SharedStatus of IsShared: true", func() {
						Expect(serviceInstance.SharedStatus.IsShared).To(BeTrue())
					})
				})

				When("the service instance does not have shared spaces", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]resources.Space{},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns a service with a SharedStatus of IsShared: false", func() {
						Expect(serviceInstance.SharedStatus.FeatureFlagIsDisabled).To(BeFalse())
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
						fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
							resources.ServiceOffering{Name: serviceOfferingName, AllowsInstanceSharing: false},
							ccv3.Warnings{},
							nil,
						)
					})

					It("returns service is not shared and feature flag info", func() {
						Expect(serviceInstance.SharedStatus.OfferingDisablesSharing).To(BeTrue())

						actualOfferingName, actualBrokerName := fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerArgsForCall(0)
						Expect(actualOfferingName).To(Equal(serviceOfferingName))
						Expect(actualBrokerName).To(Equal(serviceBrokerName))
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

						actualOfferingName, actualBrokerName := fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerArgsForCall(0)
						Expect(actualOfferingName).To(Equal(serviceOfferingName))
						Expect(actualBrokerName).To(Equal(serviceBrokerName))
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
						Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-parameters-warning", warningMessage))
					})
				})

				When("getting offering errors", func() {
					const warningMessage = "some-offering-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceOfferingByNameAndBrokerReturns(
							resources.ServiceOffering{},
							ccv3.Warnings{warningMessage},
							errors.New("error getting offering"),
						)
					})

					It("returns an empty service instance, warnings, and the error", func() {
						Expect(serviceInstance).To(Equal(ServiceInstanceDetails{}))
						Expect(executionError).To(MatchError("error getting offering"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-parameters-warning", warningMessage))
					})
				})

				When("the fetching spaces returns new warnings", func() {
					const warningMessage = "some-shared-spaces-warning"

					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceInstanceSharedSpacesReturns(
							[]resources.Space{},
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
						Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-parameters-warning", warningMessage))
					})
				})
			})
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

			It("does not attempt to get the parameters", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
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

			It("does not attempt to get the parameters", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceParametersCallCount()).To(Equal(0))
			})
		})

		When("getting the parameters fails with an expected error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceParametersReturns(
					types.OptionalObject{},
					ccv3.Warnings{"some-parameters-warning"},
					ccerror.V3UnexpectedResponseError{
						V3ErrorResponse: ccerror.V3ErrorResponse{
							Errors: []ccerror.V3Error{{
								Code:   1234,
								Detail: "cannot get parameters reason",
								Title:  "CF-SomeRandomError",
							}},
						},
					},
				)
			})

			It("does not return an error, but returns warnings and the reason", func() {
				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElement("some-parameters-warning"))
				Expect(serviceInstance.ParametersMissingReason).To(Equal("cannot get parameters reason"))
			})
		})

		When("getting the parameters fails with an another error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceParametersReturns(
					types.OptionalObject{},
					ccv3.Warnings{"some-parameters-warning"},
					errors.New("not expected"),
				)
			})

			It("does not return an error, but returns warnings and the reason", func() {
				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElement("some-parameters-warning"))
				Expect(serviceInstance.ParametersMissingReason).To(Equal("not expected"))
			})
		})
	})

	Describe("UnshareServiceInstanceByServiceInstanceAndSpace", func() {
		var (
			serviceInstanceGUID string
			sharedToSpaceGUID   string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
			sharedToSpaceGUID = "some-other-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID, sharedToSpaceGUID)
		})

		When("no errors occur deleting the service instance share relationship", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceReturns(
					ccv3.Warnings{"delete-share-relationship-warning"},
					nil)
			})

			It("returns no errors and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("delete-share-relationship-warning"))

				Expect(fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceCallCount()).To(Equal(1))
				serviceInstanceGUIDArg, sharedToSpaceGUIDArg := fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceArgsForCall(0)
				Expect(serviceInstanceGUIDArg).To(Equal(serviceInstanceGUID))
				Expect(sharedToSpaceGUIDArg).To(Equal(sharedToSpaceGUID))
			})
		})

		When("an error occurs deleting the service instance share relationship", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("delete share relationship error")
				fakeCloudControllerClient.DeleteServiceInstanceRelationshipsSharedSpaceReturns(
					ccv3.Warnings{"delete-share-relationship-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("delete-share-relationship-warning"))
			})
		})
	})

	Describe("CreateUserProvidedServiceInstance", func() {
		When("the service instance is created successfully", func() {
			It("returns warnings", func() {
				fakeCloudControllerClient.CreateServiceInstanceReturns("", ccv3.Warnings{"fake-warning"}, nil)

				warnings, err := actor.CreateUserProvidedServiceInstance(resources.ServiceInstance{
					Name:            "fake-upsi-name",
					SpaceGUID:       "fake-space-guid",
					Tags:            types.NewOptionalStringSlice("foo", "bar"),
					RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
					SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
					Credentials: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
						"baz": 42,
					}),
				})
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateServiceInstanceArgsForCall(0)).To(Equal(resources.ServiceInstance{
					Type:            "user-provided",
					Name:            "fake-upsi-name",
					SpaceGUID:       "fake-space-guid",
					Tags:            types.NewOptionalStringSlice("foo", "bar"),
					RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
					SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
					Credentials: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
						"baz": 42,
					}),
				}))
			})
		})

		When("there is an error creating the service instance", func() {
			It("returns warnings and an error", func() {
				fakeCloudControllerClient.CreateServiceInstanceReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))

				warnings, err := actor.CreateUserProvidedServiceInstance(resources.ServiceInstance{
					Name:      "fake-upsi-name",
					SpaceGUID: "fake-space-guid",
				})
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).To(MatchError("bang"))
			})
		})
	})

	Describe("UpdateUserProvidedServiceInstance", func() {
		const (
			originalName = "fake-service-instance-name"
			guid         = "fake-service-instance-guid"
			spaceGUID    = "fake-space-guid"
		)

		When("the service instance is updated successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.UserProvidedServiceInstance,
						GUID: guid,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					"",
					ccv3.Warnings{"warning from update"},
					nil,
				)
			})

			It("returns all warnings", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					originalName,
					spaceGUID,
					resources.ServiceInstance{
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					},
				)
				Expect(warnings).To(ConsistOf("warning from get", "warning from update"))
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(originalName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())

				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
				actualGUID, actualServiceInstance := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
				Expect(actualGUID).To(Equal(guid))
				Expect(actualServiceInstance).To(Equal(resources.ServiceInstance{
					SpaceGUID:       "fake-space-guid",
					Tags:            types.NewOptionalStringSlice("foo", "bar"),
					RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
					SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
					Credentials: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
						"baz": 42,
					}),
				}))
			})
		})

		When("the service instance is not user-provided", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.ManagedServiceInstance,
						GUID: guid,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
			})

			It("fails with warnings", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					originalName,
					spaceGUID,
					resources.ServiceInstance{
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					},
				)
				Expect(warnings).To(ConsistOf("warning from get"))

				Expect(err).To(MatchError(actionerror.ServiceInstanceTypeError{
					Name:         originalName,
					RequiredType: resources.UserProvidedServiceInstance,
				}))
			})
		})

		When("there is an error getting the service instance", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					errors.New("bang"),
				)
			})

			It("returns warnings and an error", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					originalName,
					spaceGUID,
					resources.ServiceInstance{
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					},
				)
				Expect(warnings).To(ConsistOf("warning from get"))
				Expect(err).To(MatchError("bang"))
			})
		})

		When("there is an error updating the service instance", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.UserProvidedServiceInstance,
						GUID: guid,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					"",
					ccv3.Warnings{"warning from update"},
					errors.New("boom"),
				)
			})

			It("returns warnings and an error", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					originalName,
					spaceGUID,
					resources.ServiceInstance{
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					},
				)
				Expect(warnings).To(ConsistOf("warning from get", "warning from update"))
				Expect(err).To(MatchError("boom"))
			})
		})
	})

	Describe("RenameServiceInstance", func() {
		const (
			currentServiceInstanceName = "current-service-instance-name"
			currentServiceInstanceGUID = "current-service-instance-guid"
			newServiceInstanceName     = "new-service-instance-name"
			spaceGUID                  = "some-source-space-guid"
		)

		var (
			warnings       Warnings
			executionError error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: currentServiceInstanceName,
					GUID: currentServiceInstanceGUID,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"some-get-service-instance-warning"},
				nil,
			)

			fakeCloudControllerClient.UpdateServiceInstanceReturns(
				"",
				ccv3.Warnings{"some-update-service-instance-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.RenameServiceInstance(currentServiceInstanceName, spaceGUID, newServiceInstanceName)
		})

		It("gets the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualName).To(Equal(currentServiceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualQuery).To(BeEmpty())
		})

		It("updates the service instance", func() {
			Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
			actualGUID, actualUpdates := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
			Expect(actualGUID).To(Equal(currentServiceInstanceGUID))
			Expect(actualUpdates).To(Equal(resources.ServiceInstance{Name: newServiceInstanceName}))
		})

		It("returns warnings and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("some-get-service-instance-warning", "some-update-service-instance-warning"))
		})

		When("the update is synchronous", func() {
			It("does not wait on the job", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(0))
			})
		})

		When("the update is asynchronous", func() {
			const job = "fake-job-url"

			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					job,
					ccv3.Warnings{"some-update-service-instance-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"some-poll-job-warning"},
					nil,
				)
			})

			It("waits on the job", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(BeEquivalentTo(job))
				Expect(warnings).To(ContainElement("some-poll-job-warning"))
			})

			When("polling the job returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(
						ccv3.Warnings{"some-poll-job-warning"},
						errors.New("bad polling issue"),
					)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError("bad polling issue"))
					Expect(warnings).To(ContainElement("some-poll-job-warning"))
				})
			})
		})

		When("the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					ccerror.ServiceInstanceNotFoundError{Name: currentServiceInstanceName, SpaceGUID: spaceGUID},
				)
			})

			It("returns an actor error and warnings", func() {
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: currentServiceInstanceName}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})

		When("getting the service instance returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("no service instance"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service instance"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})

			It("does not attempt to update the service instance", func() {
				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(0))
			})
		})

		When("updating the service instance returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					"",
					ccv3.Warnings{"some-update-service-instance-warning"},
					errors.New("something awful"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("something awful"))
				Expect(warnings).To(ContainElement("some-update-service-instance-warning"))
			})
		})
	})
})
