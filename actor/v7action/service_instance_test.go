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
	. "github.com/onsi/ginkgo/v2"
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
			serviceInstanceGUID = "fake-service-instance-guid"
			servicePlanGUID     = "fake-service-plan-guid"
			serviceOfferingName = "fake-service-offering-name"
			serviceOfferingGUID = "fake-service-offering-guid"
			serviceBrokerName   = "fake-service-broker-name"
			spaceGUID           = "fake-space-guid"
			newServicePlanGUID  = "fake-new-service-plan-guid"
			newServicePlanName  = "fake-new-service-plan-name"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			warnings      Warnings
			executeErr    error
			stream        chan PollJobEvent
			params        UpdateManagedServiceInstanceParams
			newTags       types.OptionalStringSlice
			newParameters types.OptionalObject
		)

		BeforeEach(func() {
			newTags = types.NewOptionalStringSlice("foo")
			newParameters = types.NewOptionalObject(map[string]interface{}{"foo": "bar"})
			params = UpdateManagedServiceInstanceParams{
				ServiceInstanceName: serviceInstanceName,
				SpaceGUID:           spaceGUID,
				ServicePlanName:     newServicePlanName,
				Tags:                newTags,
				Parameters:          newParameters,
			}

			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Type:            resources.ManagedServiceInstance,
					GUID:            serviceInstanceGUID,
					Name:            serviceInstanceName,
					ServicePlanGUID: servicePlanGUID,
				},
				ccv3.IncludedResources{
					ServiceBrokers: []resources.ServiceBroker{{
						Name: serviceBrokerName,
					}},
					ServiceOfferings: []resources.ServiceOffering{{
						Name: serviceOfferingName,
						GUID: serviceOfferingGUID,
					}},
				},
				ccv3.Warnings{"fake get service instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]resources.ServicePlan{{
					GUID: newServicePlanGUID,
					Name: newServicePlanName,
				}},
				ccv3.Warnings{"fake get service plan warning"},
				nil,
			)

			fakeCloudControllerClient.UpdateServiceInstanceReturns(
				fakeJobURL,
				ccv3.Warnings{"fake update service instance warning"},
				nil,
			)

			fakeStream := make(chan ccv3.PollJobEvent)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobProcessing,
					Warnings: ccv3.Warnings{"stream warning"},
				}
			}()
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
		})

		JustBeforeEach(func() {
			stream, warnings, executeErr = actor.UpdateManagedServiceInstance(params)
		})

		It("returns a stream, warnings, and no error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf(
				"fake get service instance warning",
				"fake get service plan warning",
				"fake update service instance warning",
			))
			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobProcessing,
				Warnings: Warnings{"stream warning"},
			})))
		})

		Describe("getting the service instance", func() {
			It("makes the right call", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(ConsistOf(
					[]ccv3.Query{
						{
							Key:    ccv3.FieldsServicePlanServiceOffering,
							Values: []string{"name", "guid"},
						},
						{
							Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
							Values: []string{"name"},
						},
					},
				))
			})

			When("not updating the plan", func() {
				BeforeEach(func() {
					params.ServicePlanName = ""
				})

				It("does not include the offering and broker", func() {
					Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
					_, _, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)

					Expect(actualQuery).To(BeEmpty())
				})
			})

			When("not a managed service instance", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{
							Type: resources.UserProvidedServiceInstance,
							Name: serviceInstanceName,
							GUID: serviceInstanceGUID,
						},
						ccv3.IncludedResources{},
						ccv3.Warnings{"warning from get"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ConsistOf("warning from get"))
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceTypeError{
						Name:         serviceInstanceName,
						RequiredType: resources.ManagedServiceInstance,
					}))
				})
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"warning from get"},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns warnings and an actionerror", func() {
					Expect(warnings).To(ConsistOf("warning from get"))
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
						Name: serviceInstanceName,
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"warning from get"},
						errors.New("bang"),
					)
				})

				It("returns warnings and an error", func() {
					Expect(warnings).To(ContainElement("warning from get"))
					Expect(executeErr).To(MatchError("bang"))
				})
			})
		})

		Describe("checking the new plan", func() {
			It("gets the plan", func() {
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.ServiceOfferingGUIDsFilter, Values: []string{serviceOfferingGUID}},
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{newServicePlanName}},
				))
			})

			When("no plan change requested", func() {
				BeforeEach(func() {
					params.ServicePlanName = ""
				})

				It("does not get the plan", func() {
					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(BeZero())
				})
			})

			When("no plan found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{},
						ccv3.Warnings{"fake get service plan warning"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.ServicePlanNotFoundError{
						PlanName:          newServicePlanName,
						OfferingName:      serviceOfferingName,
						ServiceBrokerName: serviceBrokerName,
					}))

					Expect(warnings).To(ConsistOf(
						"fake get service instance warning",
						"fake get service plan warning",
					))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{},
						ccv3.Warnings{"fake get service plan warning"},
						errors.New("bang"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("bang"))

					Expect(warnings).To(ConsistOf(
						"fake get service instance warning",
						"fake get service plan warning",
					))
				})
			})

		})

		Describe("detecting no-op updates", func() {
			When("no updates are requested", func() {
				BeforeEach(func() {
					params.ServicePlanName = ""
					params.Tags = types.OptionalStringSlice{}
					params.Parameters = types.OptionalObject{}
				})

				It("returns a no-op error", func() {
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceUpdateIsNoop{}))
				})
			})

			When("the new plan is the same as the old plan", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]resources.ServicePlan{{
							GUID: servicePlanGUID,
						}},
						ccv3.Warnings{"fake get service plan warning"},
						nil,
					)
				})

				Context("and no other updates are requested", func() {
					BeforeEach(func() {
						params.Tags = types.OptionalStringSlice{}
						params.Parameters = types.OptionalObject{}
					})

					It("returns a no-op error", func() {
						Expect(executeErr).To(MatchError(actionerror.ServiceInstanceUpdateIsNoop{}))
					})
				})

				Context("and a tag change is requested", func() {
					BeforeEach(func() {
						params.Parameters = types.OptionalObject{}
					})

					It("does not return a no-op error", func() {
						Expect(executeErr).NotTo(HaveOccurred())
					})
				})

				Context("and a parameter change is requested", func() {
					BeforeEach(func() {
						params.Tags = types.OptionalStringSlice{}
					})

					It("does not return a no-op error", func() {
						Expect(executeErr).NotTo(HaveOccurred())
					})
				})
			})
		})

		Describe("initiating the update", func() {
			It("makes the right call", func() {
				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
				actualServiceInstanceGUID, actualUpdateResource := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
				Expect(actualServiceInstanceGUID).To(Equal(serviceInstanceGUID))
				Expect(actualUpdateResource).To(Equal(resources.ServiceInstance{
					ServicePlanGUID: newServicePlanGUID,
					Tags:            newTags,
					Parameters:      newParameters,
				}))
			})

			When("just changing tags", func() {
				BeforeEach(func() {
					params.ServicePlanName = ""
					params.Parameters = types.OptionalObject{}
				})

				It("makes the right call", func() {
					Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
					actualServiceInstanceGUID, actualUpdateResource := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
					Expect(actualServiceInstanceGUID).To(Equal(serviceInstanceGUID))
					Expect(actualUpdateResource).To(Equal(resources.ServiceInstance{
						Tags: newTags,
					}))
				})
			})

			When("just changing parameters", func() {
				BeforeEach(func() {
					params.ServicePlanName = ""
					params.Tags = types.OptionalStringSlice{}
				})

				It("makes the right call", func() {
					Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
					actualServiceInstanceGUID, actualUpdateResource := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
					Expect(actualServiceInstanceGUID).To(Equal(serviceInstanceGUID))
					Expect(actualUpdateResource).To(Equal(resources.ServiceInstance{
						Parameters: newParameters,
					}))
				})

				When("just changing plan", func() {

				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateServiceInstanceReturns(
						"",
						ccv3.Warnings{"fake update service instance warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(stream).To(BeNil())
					Expect(warnings).To(ConsistOf(
						"fake get service instance warning",
						"fake get service plan warning",
						"fake update service instance warning",
					))
					Expect(executeErr).To(MatchError("boom"))
				})
			})
		})

		Describe("polling the job", func() {
			It("polls the job", func() {
				Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)).To(Equal(fakeJobURL))
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
			stream                chan PollJobEvent
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

			fakeStream := make(chan ccv3.PollJobEvent)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobPolling,
					Err:      errors.New("fake error"),
					Warnings: ccv3.Warnings{"fake warning"},
				}
			}()
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
		})

		JustBeforeEach(func() {
			params := CreateManagedServiceInstanceParams{
				ServiceOfferingName: fakeServiceOfferingName,
				ServicePlanName:     fakeServicePlanName,
				ServiceInstanceName: fakeServiceInstanceName,
				ServiceBrokerName:   fakeServiceBrokerName,
				SpaceGUID:           fakeSpaceGUID,
				Tags:                fakeTags,
				Parameters:          fakeParams,
			}
			stream, warnings, err = actor.CreateManagedServiceInstance(params)
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

		It("polls the job", func() {
			Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
			jobUrl := fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)
			Expect(jobUrl).To(Equal(fakeJobURL))
		})

		It("returns an event stream and warnings", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("plan-warning", "fake-warning"))
			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Err:      errors.New("fake error"),
				Warnings: Warnings{"fake warning"},
			})))
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
					Expect(stream).To(BeNil())
				})
			})

			When("more than one plan found", func() {
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
					Expect(err).To(MatchError(actionerror.ServiceBrokerNameRequiredError{
						ServiceOfferingName: fakeServiceOfferingName,
					}))
					Expect(stream).To(BeNil())
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
					Expect(stream).To(BeNil())
				})
			})

			When("there is an error creating the service instance", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceInstanceReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns warnings and an error", func() {
					Expect(fakeCloudControllerClient.PollJobForStateCallCount()).To(Equal(0))

					Expect(warnings).To(ConsistOf("plan-warning", "fake-warning"))
					Expect(err).To(MatchError("bang"))
					Expect(stream).To(BeNil())
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
			warnings Warnings
			err      error
			stream   chan PollJobEvent
		)

		JustBeforeEach(func() {
			stream, warnings, err = actor.DeleteServiceInstance(fakeServiceInstanceName, fakeSpaceGUID)
		})

		It("makes a request to get the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualName, actualSpace, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualName).To(Equal(fakeServiceInstanceName))
			Expect(actualSpace).To(Equal(fakeSpaceGUID))
			Expect(actualQuery).To(BeEmpty())
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

				It("returns a nil channel", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get warning", "delete warning"))
					Expect(stream).To(BeNil())
				})

				It("does not try to poll a job", func() {
					Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(BeZero())
				})
			})

			When("the delete response is asynchronous", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceInstanceReturns(
						"a fake job url",
						ccv3.Warnings{"delete warning"},
						nil,
					)

					fakeStream := make(chan ccv3.PollJobEvent)
					go func() {
						fakeStream <- ccv3.PollJobEvent{
							State:    constant.JobPolling,
							Err:      errors.New("fake error"),
							Warnings: ccv3.Warnings{"fake warning"},
						}
					}()
					fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
				})

				It("polls the job", func() {
					Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
					actualJobURL := fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)
					Expect(actualJobURL).To(BeEquivalentTo("a fake job url"))
				})

				It("returns an event stream and warnings", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get warning", "delete warning"))
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    JobPolling,
						Err:      errors.New("fake error"),
						Warnings: Warnings{"fake warning"},
					})))
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
					Expect(stream).To(BeNil())
				})
			})
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

			It("returns an error", func() {
				Expect(err).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName}))
				Expect(warnings).To(ConsistOf("get warning"))
				Expect(stream).To(BeNil())
			})

			It("does not try to delete", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(BeZero())
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
				Expect(stream).To(BeNil())
			})

			It("does not try to delete", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(BeZero())
			})
		})
	})

	Describe("PurgeServiceInstance", func() {
		const (
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeSpaceGUID           = "fake-space-GUID"
		)

		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = actor.PurgeServiceInstance(fakeServiceInstanceName, fakeSpaceGUID)
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

			It("returns the appropriate error", func() {
				Expect(err).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName}))
				Expect(warnings).To(ConsistOf("get warning"))
			})

			It("does not try to purge", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(BeZero())
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

				fakeCloudControllerClient.DeleteServiceInstanceReturns(
					"",
					ccv3.Warnings{"purge warning"},
					nil,
				)
			})

			It("makes the right call to purge the service instance", func() {
				Expect(fakeCloudControllerClient.DeleteServiceInstanceCallCount()).To(Equal(1))
				actualGUID, actualQuery := fakeCloudControllerClient.DeleteServiceInstanceArgsForCall(0)
				Expect(actualGUID).To(Equal(fakeServiceInstanceGUID))
				Expect(actualQuery).To(ConsistOf(ccv3.Query{Key: ccv3.Purge, Values: []string{"true"}}))
			})

			It("returns warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get warning", "purge warning"))
			})

			When("the purge responds with failure", func() {
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
			})
		})
	})

	Describe("UpgradeManagedServiceInstance", func() {
		const (
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeSpaceGUID           = "fake-space-GUID"
		)

		var (
			executeErr error
			warnings   Warnings
			stream     chan PollJobEvent
		)

		JustBeforeEach(func() {
			stream, warnings, executeErr = actor.UpgradeManagedServiceInstance(fakeServiceInstanceName, fakeSpaceGUID)
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
					ccv3.Warnings{"get SI warning"},
					ccerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName},
				)
			})

			It("does not try to upgrade", func() {
				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(BeZero())
			})

			It("returns the appropriate error", func() {
				Expect(stream).To(BeNil())
				Expect(warnings).To(ConsistOf("get SI warning"))
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
					Name: fakeServiceInstanceName,
				}))
			})
		})

		When("there's no upgrade available", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						GUID: "some-guid",
						Name: "some-name",
						UpgradeAvailable: types.OptionalBoolean{
							IsSet: true,
							Value: false,
						},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get SI warning"},
					nil,
				)
			})

			It("does not try to upgrade", func() {
				Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(BeZero())
			})

			It("returns the appropriate error", func() {
				Expect(stream).To(BeNil())
				Expect(warnings).To(ConsistOf("get SI warning"))
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceUpgradeNotAvailableError{}))
			})
		})

		When("there is an upgrade available", func() {
			const guid = "some-guid"
			const planGUID = "some-plan-guid"
			const jobURL = ccv3.JobURL("fake-job-url")

			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Type: resources.ManagedServiceInstance,
						GUID: guid,
						Name: "some-name",
						UpgradeAvailable: types.OptionalBoolean{
							IsSet: true,
							Value: true,
						},
						ServicePlanGUID: planGUID,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"warning from get"},
					nil,
				)
				fakeCloudControllerClient.GetServicePlanByGUIDReturns(
					resources.ServicePlan{
						MaintenanceInfoVersion: "9.1.2",
					},
					ccv3.Warnings{"warning from plan"},
					nil,
				)
				fakeCloudControllerClient.UpdateServiceInstanceReturns(
					jobURL,
					ccv3.Warnings{"warning from update"},
					nil,
				)

				fakeStream := make(chan ccv3.PollJobEvent)
				go func() {
					fakeStream <- ccv3.PollJobEvent{
						State:    constant.JobProcessing,
						Warnings: ccv3.Warnings{"stream warning"},
					}
				}()
				fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
			})

			It("makes the right calls and returns all warnings", func() {
				By("getting the service plan", func() {
					Expect(fakeCloudControllerClient.GetServicePlanByGUIDCallCount()).To(Equal(1))
					actualPlanGUID := fakeCloudControllerClient.GetServicePlanByGUIDArgsForCall(0)
					Expect(actualPlanGUID).To(Equal(planGUID))
				})

				By("updating the service instance", func() {
					Expect(fakeCloudControllerClient.UpdateServiceInstanceCallCount()).To(Equal(1))
					actualGUID, actualServiceInstance := fakeCloudControllerClient.UpdateServiceInstanceArgsForCall(0)
					Expect(actualGUID).To(Equal(guid))
					Expect(actualServiceInstance).To(Equal(resources.ServiceInstance{
						MaintenanceInfoVersion: "9.1.2",
					}))
				})

				By("polling the job", func() {
					Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)).To(Equal(jobURL))
				})

				By("returning a stream, warnings and no error", func() {
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    JobProcessing,
						Warnings: Warnings{"stream warning"},
					})))

					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning from get", "warning from plan", "warning from update"))
				})
			})
		})
	})
})
