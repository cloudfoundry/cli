package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
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
			serviceInstanceName = "fake-service-instance-name"
			guid                = "fake-service-instance-guid"
			spaceGUID           = "fake-space-guid"
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

			It("makes the right calls and returns all warnings", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					resources.ServiceInstance{
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
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())

				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
				actualGUID, actualServiceInstance := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
				Expect(actualGUID).To(Equal(guid))
				Expect(actualServiceInstance).To(Equal(resources.ServiceInstance{
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
						Name: serviceInstanceName,
						GUID: guid,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
			})

			It("fails with warnings", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					resources.ServiceInstance{
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
					Name:         serviceInstanceName,
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
					serviceInstanceName,
					spaceGUID,
					resources.ServiceInstance{
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
					serviceInstanceName,
					spaceGUID,
					resources.ServiceInstance{
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

	Describe("UpdateManagedServiceInstance", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			guid                = "fake-service-instance-guid"
			spaceGUID           = "fake-space-guid"
		)

		When("the service instance is successfully updated synchronously", func() {
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
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					"",
					ccv3.Warnings{"warning from update"},
					nil,
				)
			})

			It("makes the right calls and returns all warnings", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{
						Tags:       types.NewOptionalStringSlice("foo", "bar"),
						Parameters: types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
					},
				)

				By("returning warnings and no error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning from get", "warning from update"))
				})

				By("getting the service instance", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
					actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
					Expect(actualName).To(Equal(serviceInstanceName))
					Expect(actualSpaceGUID).To(Equal(spaceGUID))
					Expect(actualQuery).To(BeEmpty())
				})

				By("updating the service instance", func() {
					Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
					actualGUID, actualServiceInstance := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
					Expect(actualGUID).To(Equal(guid))
					Expect(actualServiceInstance).To(Equal(resources.ServiceInstance{
						Tags:       types.NewOptionalStringSlice("foo", "bar"),
						Parameters: types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
					}))
				})

				By("specifying an empty job URL", func() {
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(1))
					actualURL, _ := fakeCloudControllerClient.PollJobForStateArgsForCall(0)
					Expect(actualURL).To(BeEmpty())
				})
			})
		})

		When("the service instance is successfully updated asynchronously", func() {
			const jobURL = ccv3.JobURL("fake-job-url")

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
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					jobURL,
					ccv3.Warnings{"warning from update"},
					nil,
				)
				fakeCloudControllerClient.PollJobForStateReturns(
					ccv3.Warnings{"warning from poll"},
					nil,
				)
			})

			It("makes the right calls and returns all warnings", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{
						Tags:       types.NewOptionalStringSlice("foo", "bar"),
						Parameters: types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
					},
				)

				By("returning warnings and no error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning from get", "warning from update", "warning from poll"))
				})

				By("getting the service instance", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
					actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
					Expect(actualName).To(Equal(serviceInstanceName))
					Expect(actualSpaceGUID).To(Equal(spaceGUID))
					Expect(actualQuery).To(BeEmpty())
				})

				By("updating the service instance", func() {
					Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
					actualGUID, actualServiceInstance := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
					Expect(actualGUID).To(Equal(guid))
					Expect(actualServiceInstance).To(Equal(resources.ServiceInstance{
						Tags:       types.NewOptionalStringSlice("foo", "bar"),
						Parameters: types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
					}))
				})

				By("polling the job", func() {
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(1))
					actualURL, actualState := fakeCloudControllerClient.PollJobForStateArgsForCall(0)
					Expect(actualURL).To(Equal(jobURL))
					Expect(actualState).To(Equal(constant.JobPolling))
				})
			})
		})

		When("the service instance is not managed", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.UserProvidedServiceInstance,
						Name: serviceInstanceName,
						GUID: guid,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
			})

			It("fails with warnings", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{Tags: types.NewOptionalStringSlice("foo", "bar")},
				)
				Expect(warnings).To(ConsistOf("warning from get"))

				Expect(err).To(MatchError(actionerror.ServiceInstanceTypeError{
					Name:         serviceInstanceName,
					RequiredType: resources.ManagedServiceInstance,
				}))
			})
		})

		When("the service instance is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("returns warnings and an actionerror", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{Tags: types.NewOptionalStringSlice("foo", "bar")},
				)
				Expect(warnings).To(ConsistOf("warning from get"))
				Expect(err).To(MatchError(actionerror.ServiceInstanceNotFoundError{
					Name: serviceInstanceName,
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
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{Tags: types.NewOptionalStringSlice("foo", "bar")},
				)
				Expect(warnings).To(ConsistOf("warning from get"))
				Expect(err).To(MatchError("bang"))
			})
		})

		When("there is an error updating the service instance", func() {
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
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					"",
					ccv3.Warnings{"warning from update"},
					errors.New("boom"),
				)
			})

			It("returns warnings and an error", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{Tags: types.NewOptionalStringSlice("foo", "bar")},
				)
				Expect(warnings).To(ConsistOf("warning from get", "warning from update"))
				Expect(err).To(MatchError("boom"))
			})
		})

		Context("getting the plan", func() {
			var (
				warnings Warnings
				err      error
			)

			const (
				fakeServicePlanName     = "invalid-plan"
				fakeServiceOfferingName = "my-service-offering"
				fakeServiceBrokerName   = "my-broker"
			)

			JustBeforeEach(func() {
				params := ServiceInstanceUpdateManagedParams{
					ServicePlanName: types.NewOptionalString(fakeServicePlanName),
				}

				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type:                resources.ManagedServiceInstance,
						GUID:                guid,
						ServiceOfferingGUID: "something",
					},
					ccv3.IncludedResources{
						ServiceOfferings: []resources.ServiceOffering{{Name: fakeServiceOfferingName}},
						ServiceBrokers:   []resources.ServiceBroker{{Name: fakeServiceBrokerName}},
					},
					ccv3.Warnings{"warning from get"},
					nil,
				)
				warnings, err = actor.UpdateManagedServiceInstance(serviceInstanceName, spaceGUID, params)
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns([]resources.ServicePlan{}, ccv3.Warnings{"be warned"}, nil)
			})

			It("makes the right call to fetch the plan", func() {
				By("Passing the right query to fetch the instance")
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(ConsistOf(
					[]ccv3.Query{
						{
							Key:    ccv3.FieldsServicePlanServiceOffering,
							Values: []string{"name"},
						},
						{
							Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
							Values: []string{"name"},
						},
					},
				))

				By("Passing the right query to get the plan")
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetServicePlansArgsForCall(0)
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values[0]).To(Equal(fakeServicePlanName))
				Expect(query[1].Key).To(Equal(ccv3.ServiceBrokerNamesFilter))
				Expect(query[1].Values[0]).To(Equal(fakeServiceBrokerName))
				Expect(query[2].Key).To(Equal(ccv3.ServiceOfferingNamesFilter))
				Expect(query[2].Values[0]).To(Equal(fakeServiceOfferingName))

			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf("be warned", "warning from get"))

			})

			Context("errors getting the plan", func() {
				When("no plan found", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServicePlansReturns([]resources.ServicePlan{}, ccv3.Warnings{"be warned"}, nil)
					})

					It("returns with warnings and an error", func() {
						Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

						Expect(warnings).To(ConsistOf("warning from get", "be warned"))
						Expect(err).To(MatchError(actionerror.ServicePlanNotFoundError{
							PlanName:          fakeServicePlanName,
							ServiceBrokerName: fakeServiceBrokerName,
							OfferingName:      fakeServiceOfferingName,
						}))
					})
				})

				When("client error when getting the plan", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServicePlansReturns([]resources.ServicePlan{}, ccv3.Warnings{"be warned"}, errors.New("boom"))
					})

					It("returns warnings and an error", func() {
						Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

						Expect(warnings).To(ConsistOf("warning from get", "be warned"))
						Expect(err).To(MatchError("boom"))
					})
				})
			})
		})

		When("there is an error polling the job", func() {
			const jobURL = ccv3.JobURL("fake-job-url")

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
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					jobURL,
					ccv3.Warnings{"warning from update"},
					nil,
				)
				fakeCloudControllerClient.PollJobForStateReturns(
					ccv3.Warnings{"warning from poll"},
					errors.New("boom"),
				)
			})

			It("returns warnings and an error", func() {
				warnings, err := actor.UpdateManagedServiceInstance(
					serviceInstanceName,
					spaceGUID,
					ServiceInstanceUpdateManagedParams{Tags: types.NewOptionalStringSlice("foo", "bar")},
				)
				Expect(warnings).To(ConsistOf("warning from get", "warning from update", "warning from poll"))
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

				fakeCloudControllerClient.PollJobForStateReturns(
					ccv3.Warnings{"some-poll-job-warning"},
					nil,
				)
			})

			It("waits on the job", func() {
				Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(1))
				actualURL, actualState := fakeCloudControllerClient.PollJobForStateArgsForCall(0)
				Expect(actualURL).To(Equal(actualURL))
				Expect(actualState).To(Equal(constant.JobPolling))
				Expect(warnings).To(ContainElement("some-poll-job-warning"))
			})

			When("polling the job returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobForStateReturns(
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

	Describe("CreateManagedServiceInstance", func() {
		const (
			fakeServiceOfferingName = "fake-offering-name"
			fakeServicePlanName     = "fake-plan-name"
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeSpaceGUID           = "fake-space-GUID"
		)

		var (
			fakeServiceBrokerName string
			fakeTags              types.OptionalStringSlice
			warnings              Warnings
			err                   error
			fakeJobURL            ccv3.JobURL
			fakeParams            types.OptionalObject
		)

		BeforeEach(func() {
			fakeServiceBrokerName = "fake-broker-name"
			fakeTags = types.NewOptionalStringSlice("tag1", "tag2")
			fakeJobURL = "http://some-cc-api/v3/jobs/job-guid"
			fakeParams = types.NewOptionalObject(map[string]interface{}{"param1": "some-value", "param-2": "cool service"})

			fakeCloudControllerClient.GetServicePlansReturns(
				[]resources.ServicePlan{{GUID: "fake-plan-guid"}},
				ccv3.Warnings{"plan-warning"},
				nil,
			)
			fakeCloudControllerClient.CreateServiceInstanceReturns(fakeJobURL, ccv3.Warnings{"fake-warning"}, nil)
			fakeCloudControllerClient.PollJobForStateReturns(ccv3.Warnings{"job-warning"}, nil)
		})

		JustBeforeEach(func() {
			params := ManagedServiceInstanceParams{
				ServiceOfferingName: fakeServiceOfferingName,
				ServicePlanName:     fakeServicePlanName,
				ServiceInstanceName: fakeServiceInstanceName,
				ServiceBrokerName:   fakeServiceBrokerName,
				SpaceGUID:           fakeSpaceGUID,
				Tags:                fakeTags,
				Parameters:          fakeParams,
			}
			warnings, err = actor.CreateManagedServiceInstance(params)

		})

		It("gets the service plan", func() {
			Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
			query := fakeCloudControllerClient.GetServicePlansArgsForCall(0)
			Expect(query[0].Values).To(ConsistOf(fakeServicePlanName))
			Expect(query[0].Key).To(Equal(ccv3.NameFilter))
			Expect(query[1].Values).To(ConsistOf(fakeServiceBrokerName))
			Expect(query[1].Key).To(Equal(ccv3.ServiceBrokerNamesFilter))
			Expect(query[2].Values).To(ConsistOf(fakeServiceOfferingName))
			Expect(query[2].Key).To(Equal(ccv3.ServiceOfferingNamesFilter))
		})

		It("calls the client to create the instance", func() {
			Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.CreateServiceInstanceArgsForCall(0)).To(Equal(resources.ServiceInstance{
				Type:            "managed",
				Name:            fakeServiceInstanceName,
				ServicePlanGUID: "fake-plan-guid",
				SpaceGUID:       fakeSpaceGUID,
				Tags:            fakeTags,
				Parameters:      fakeParams,
			}))
		})

		It("polls the job until is in polling state", func() {
			Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(1))
			jobUrl, status := fakeCloudControllerClient.PollJobForStateArgsForCall(0)
			Expect(jobUrl).To(Equal(fakeJobURL))
			Expect(status).To(Equal(constant.JobPolling))
		})

		It("returns all warnings", func() {
			Expect(warnings).To(ConsistOf("plan-warning", "fake-warning", "job-warning"))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("error scenarios", func() {
			When("no plan found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]resources.ServicePlan{}, ccv3.Warnings{"be warned"}, nil)
				})

				It("returns with warnings and an error", func() {
					Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

					Expect(warnings).To(ConsistOf("be warned"))
					Expect(err).To(MatchError(actionerror.ServicePlanNotFoundError{PlanName: fakeServicePlanName, OfferingName: fakeServiceOfferingName, ServiceBrokerName: fakeServiceBrokerName}))
				})
			})

			When("more than a plan found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{{GUID: "a-guid"}, {GUID: "another-guid"}},
						ccv3.Warnings{"be warned"},
						nil,
					)
					fakeServiceBrokerName = ""
				})

				It("returns warnings and an error", func() {
					Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

					Expect(warnings).To(ConsistOf("be warned"))
					Expect(err).To(MatchError(actionerror.DuplicateServicePlanError{Name: fakeServicePlanName, ServiceOfferingName: fakeServiceOfferingName, ServiceBrokerName: ""}))
				})
			})

			When("client error when getting the plan", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]resources.ServicePlan{}, ccv3.Warnings{"be warned"}, errors.New("boom"))
					fakeServiceBrokerName = ""
				})

				It("returns warnings and an error", func() {
					Expect(fakeCloudControllerClient.CreateServiceInstanceCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

					Expect(warnings).To(ConsistOf("be warned"))
					Expect(err).To(MatchError("boom"))
				})
			})

			When("there is an error creating the job", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceInstanceReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns warnings and an error", func() {
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

					Expect(warnings).To(ConsistOf("plan-warning", "fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})

			When("there are job errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobForStateReturns(ccv3.Warnings{"job-warning"}, errors.New("bad job"))
				})

				It("returns warnings and an error", func() {
					Expect(warnings).To(ConsistOf("plan-warning", "fake-warning", "job-warning"))
					Expect(err).To(MatchError("bad job"))
				})
			})
		})
	})

	Describe("DeleteServiceInstance", func() {
		const (
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeSpaceGUID           = "fake-space-GUID"
		)

		var (
			wait     bool
			warnings Warnings
			err      error
			state    ServiceInstanceDeleteState
		)

		BeforeEach(func() {
			wait = false
		})

		JustBeforeEach(func() {
			state, warnings, err = actor.DeleteServiceInstance(fakeServiceInstanceName, fakeSpaceGUID, wait)
		})

		It("makes a request to get the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualName, actualSpace, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualName).To(Equal(fakeServiceInstanceName))
			Expect(actualSpace).To(Equal(fakeSpaceGUID))
			Expect(actualQuery).To(BeEmpty())
		})

		When("the service instance does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get warning"},
					ccerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName},
				)
			})

			It("does not try to delete", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(BeZero())
			})

			It("returns the appropriate state flag", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get warning"))
				Expect(state).To(Equal(ServiceInstanceDidNotExist))
			})
		})

		When("the service instance exists", func() {
			const fakeServiceInstanceGUID = "fake-si-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{GUID: fakeServiceInstanceGUID},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get warning"},
					nil,
				)
			})

			It("makes the right call to delete the service instance", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServiceInstanceArgsForCall(0)).To(Equal(fakeServiceInstanceGUID))
			})

			When("the delete response is synchronous", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceInstanceReturns(
						"",
						ccv3.Warnings{"delete warning"},
						nil,
					)
				})

				It("returns the appropriate state flag", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get warning", "delete warning"))
					Expect(state).To(Equal(ServiceInstanceGone))
				})
			})

			When("the delete response is asynchronous", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceInstanceReturns(
						"a fake job url",
						ccv3.Warnings{"delete warning"},
						nil,
					)

					fakeCloudControllerClient.PollJobForStateReturns(
						ccv3.Warnings{"poll warning"},
						nil,
					)
				})

				It("polls the job", func() {
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(1))
					actualJobURL, actualState := fakeCloudControllerClient.PollJobForStateArgsForCall(0)
					Expect(actualJobURL).To(BeEquivalentTo("a fake job url"))
					Expect(actualState).To(Equal(constant.JobPolling))
				})

				It("returns the appropriate state flag", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get warning", "delete warning", "poll warning"))
					Expect(state).To(Equal(ServiceInstanceDeleteInProgress))
				})

				When("the `wait` flag is specified", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(
							ccv3.Warnings{"poll job warning"},
							nil,
						)

						wait = true
					})

					It("polls the job until complete", func() {
						Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(BeEquivalentTo("a fake job url"))
					})

					It("returns the appropriate state flag", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("get warning", "delete warning", "poll job warning"))
						Expect(state).To(Equal(ServiceInstanceGone))
					})
				})

				When("polling the job fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobForStateReturns(
							ccv3.Warnings{"poll warning"},
							errors.New("bang"),
						)
					})

					It("return the error and warnings", func() {
						Expect(err).To(MatchError("bang"))
						Expect(warnings).To(ConsistOf("get warning", "delete warning", "poll warning"))
						Expect(state).To(Equal(ServiceInstanceUnknownState))
					})
				})
			})

			When("the delete responds with failure", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceInstanceReturns(
						"a fake job url",
						ccv3.Warnings{"delete warning"},
						errors.New("bong"),
					)
				})

				It("return the error and warnings", func() {
					Expect(err).To(MatchError("bong"))
					Expect(warnings).To(ConsistOf("get warning", "delete warning"))
					Expect(state).To(Equal(ServiceInstanceUnknownState))
				})
			})
		})

		When("getting the service instance fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get warning"},
					errors.New("boom"),
				)
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("boom"))
				Expect(warnings).To(ConsistOf("get warning"))
				Expect(state).To(Equal(ServiceInstanceUnknownState))
			})
		})
	})
})
