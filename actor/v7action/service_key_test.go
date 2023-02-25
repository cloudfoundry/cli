package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Key Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("CreateServiceKey", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			serviceKeyName      = "fake-key-name"
			spaceGUID           = "fake-space-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			params         CreateServiceKeyParams
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
				fakeJobURL,
				ccv3.Warnings{"create key warning"},
				nil,
			)

			fakeStream := make(chan ccv3.PollJobEvent)
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobPolling,
					Warnings: ccv3.Warnings{"poll warning"},
				}
			}()

			params = CreateServiceKeyParams{
				SpaceGUID:           spaceGUID,
				ServiceInstanceName: serviceInstanceName,
				ServiceKeyName:      serviceKeyName,
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
			}
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.CreateServiceKey(params)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"create key warning",
			}))

			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Warnings: Warnings{"poll warning"},
				Err:      nil,
			})))
		})

		Describe("service instance lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						errors.New("boof"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError("boof"))
				})
			})
		})

		Describe("initiating the create", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.CreateServiceCredentialBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateServiceCredentialBindingArgsForCall(0)).To(Equal(resources.ServiceCredentialBinding{
					Type:                resources.KeyBinding,
					Name:                serviceKeyName,
					ServiceInstanceGUID: serviceInstanceGUID,
					Parameters: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
					}),
				}))
			})

			When("key already exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"create binding warning"},
						ccerror.ServiceKeyTakenError{
							Message: "The binding name is invalid. Key binding names must be unique. The service instance already has a key binding with name 'fake-key-name'.",
						},
					)
				})

				It("returns an actionerror and warnings", func() {
					Expect(warnings).To(ContainElement("create binding warning"))
					Expect(executionError).To(MatchError(actionerror.ResourceAlreadyExistsError{
						Message: "Service key fake-key-name already exists",
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"create binding warning"},
						errors.New("boop"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("create binding warning"))
					Expect(executionError).To(MatchError("boop"))
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

	Describe("GetServiceKeysByServiceInstance", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			spaceGUID           = "fake-space-guid"
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{GUID: "1", Name: "flopsy", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
					{GUID: "2", Name: "mopsy", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
					{GUID: "3", Name: "cottontail", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
					{GUID: "4", Name: "peter", LastOperation: resources.LastOperation{Type: "create", State: "in progress"}},
				},
				ccv3.Warnings{"get keys warning"},
				nil,
			)
		})

		var (
			keys           []resources.ServiceCredentialBinding
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			keys, warnings, executionError = actor.GetServiceKeysByServiceInstance(serviceInstanceName, spaceGUID)
		})

		It("makes the correct call to get the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualQuery).To(BeEmpty())
		})

		It("makes the correct call to get the service keys", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
			))
		})

		It("returns a list of keys, with warnings and no error", func() {
			Expect(executionError).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
			Expect(keys).To(Equal([]resources.ServiceCredentialBinding{
				{GUID: "1", Name: "flopsy", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
				{GUID: "2", Name: "mopsy", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				{GUID: "3", Name: "cottontail", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
				{GUID: "4", Name: "peter", LastOperation: resources.LastOperation{Type: "create", State: "in progress"}},
			}))
		})

		When("service instance not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
			})
		})

		When("get service instance fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					errors.New("boof"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError("boof"))
			})
		})

		When("get keys fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"get keys warning"},
					errors.New("boom"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
				Expect(executionError).To(MatchError("boom"))
			})
		})
	})

	Describe("GetServiceKeyByServiceInstanceAndName", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			serviceKeyName      = "fake-service-key-name"
			serviceKeyGUID      = "fake-service-key-guid"
			spaceGUID           = "fake-space-guid"
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{
						Name: serviceKeyName,
						GUID: serviceKeyGUID,
					},
				},
				ccv3.Warnings{"get keys warning"},
				nil,
			)
		})

		var (
			key            resources.ServiceCredentialBinding
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			key, warnings, executionError = actor.GetServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
		})

		It("makes the correct call to get the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualQuery).To(BeEmpty())
		})

		It("makes the correct call to get the service keys", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceKeyName}},
			))
		})

		It("returns a key, with warnings and no error", func() {
			Expect(executionError).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
			Expect(key.Name).To(Equal(serviceKeyName))
			Expect(key.GUID).To(Equal(serviceKeyGUID))
		})

		When("service instance not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
			})
		})

		When("get service instance fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					errors.New("boof"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError("boof"))
			})
		})

		When("key not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"get keys warning"},
					nil,
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceKeyNotFoundError{
					KeyName:             serviceKeyName,
					ServiceInstanceName: serviceInstanceName,
				}))
			})
		})

		When("get keys fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"get keys warning"},
					errors.New("boom"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
				Expect(executionError).To(MatchError("boom"))
			})
		})
	})

	Describe("GetServiceKeyDetailsByServiceInstanceAndName", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			serviceKeyName      = "fake-service-key-name"
			serviceKeyGUID      = "fake-service-key-guid"
			spaceGUID           = "fake-space-guid"
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{
						Name: serviceKeyName,
						GUID: serviceKeyGUID,
					},
				},
				ccv3.Warnings{"get keys warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingDetailsReturns(
				resources.ServiceCredentialBindingDetails{
					Credentials: map[string]interface{}{"foo": "bar"},
				},
				ccv3.Warnings{"get details warning"},
				nil,
			)
		})

		var (
			details        resources.ServiceCredentialBindingDetails
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			details, warnings, executionError = actor.GetServiceKeyDetailsByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
		})

		It("makes the correct call to get the service instance", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualQuery).To(BeEmpty())
		})

		It("makes the correct call to get the service keys", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceKeyName}},
			))
		})

		It("makes the correct call to get the key details", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingDetailsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingDetailsArgsForCall(0)).To(Equal(serviceKeyGUID))
		})

		It("returns the details, with warnings and no error", func() {
			Expect(executionError).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElements("get instance warning", "get keys warning", "get details warning"))
			Expect(details.Credentials).To(Equal(map[string]interface{}{"foo": "bar"}))
		})

		When("service instance not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
			})
		})

		When("get service instance fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					errors.New("boof"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError("boof"))
			})
		})

		When("key not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"get keys warning"},
					nil,
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceKeyNotFoundError{
					KeyName:             serviceKeyName,
					ServiceInstanceName: serviceInstanceName,
				}))
			})
		})

		When("get keys fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
					[]resources.ServiceCredentialBinding{},
					ccv3.Warnings{"get keys warning"},
					errors.New("boom"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get instance warning", "get keys warning"))
				Expect(executionError).To(MatchError("boom"))
			})
		})

		When("get details fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceCredentialBindingDetailsReturns(
					resources.ServiceCredentialBindingDetails{},
					ccv3.Warnings{"get details warning"},
					errors.New("boom"),
				)
			})

			It("returns the error and warning", func() {
				Expect(warnings).To(ContainElements("get details warning", "get keys warning"))
				Expect(executionError).To(MatchError("boom"))
			})
		})
	})

	Describe("DeleteServiceKeyByServiceInstanceAndName", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			spaceGUID           = "fake-space-guid"
			serviceKeyName      = "fake-key-name"
			serviceKeyGUID      = "fake-key-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{GUID: serviceKeyGUID},
				},
				ccv3.Warnings{"get bindings warning"},
				nil,
			)

			fakeCloudControllerClient.DeleteServiceCredentialBindingReturns(
				fakeJobURL,
				ccv3.Warnings{"delete binding warning"},
				nil,
			)

			fakeStream := make(chan ccv3.PollJobEvent)
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobPolling,
					Warnings: ccv3.Warnings{"poll warning"},
				}
			}()
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.DeleteServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"get bindings warning",
				"delete binding warning",
			}))

			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Warnings: Warnings{"poll warning"},
				Err:      nil,
			})))
		})

		Describe("service instance lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						errors.New("boof"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError("boof"))
				})
			})
		})

		Describe("key lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceKeyName}},
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"get keys warning"},
						nil,
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get keys warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceKeyNotFoundError{
						KeyName:             serviceKeyName,
						ServiceInstanceName: serviceInstanceName,
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"get key warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get key warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("initiating the delete", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.DeleteServiceCredentialBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServiceCredentialBindingArgsForCall(0)).To(Equal(serviceKeyGUID))
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"delete binding warning"},
						errors.New("boop"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("delete binding warning"))
					Expect(executionError).To(MatchError("boop"))
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
})
