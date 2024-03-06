package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Binding Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("ServiceBinding", func() {
		DescribeTable("IsInProgress",
			func(state constant.LastOperationState, expected bool) {
				serviceBinding := ServiceBinding{
					LastOperation: ccv2.LastOperation{
						State: state,
					},
				}
				Expect(serviceBinding.IsInProgress()).To(Equal(expected))
			},

			Entry("returns true for 'in progress'", constant.LastOperationInProgress, true),
			Entry("returns false for 'succeeded'", constant.LastOperationSucceeded, false),
			Entry("returns false for 'failed'", constant.LastOperationFailed, false),
			Entry("returns false for empty", constant.LastOperationState(""), false),
		)
	})

	Describe("BindServiceByApplicationAndServiceInstance", func() {
		var (
			applicationGUID     string
			serviceInstanceGUID string

			executeErr error
			warnings   Warnings
		)

		BeforeEach(func() {
			applicationGUID = "some-app-guid"
			serviceInstanceGUID = "some-service-instance-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.BindServiceByApplicationAndServiceInstance(applicationGUID, serviceInstanceGUID)
		})

		When("the binding is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBindingReturns(ccv2.ServiceBinding{}, ccv2.Warnings{"some-warnings"}, nil)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warnings"))

				Expect(fakeCloudControllerClient.CreateServiceBindingCallCount()).To(Equal(1))
				inputAppGUID, inputServiceInstanceGUID, inputBindingName, inputAcceptsIncomplete, inputParameters := fakeCloudControllerClient.CreateServiceBindingArgsForCall(0)
				Expect(inputAppGUID).To(Equal(applicationGUID))
				Expect(inputServiceInstanceGUID).To(Equal(serviceInstanceGUID))
				Expect(inputBindingName).To(Equal(""))
				Expect(inputAcceptsIncomplete).To(BeFalse())
				Expect(inputParameters).To(BeNil())
			})
		})

		When("the binding fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateServiceBindingReturns(ccv2.ServiceBinding{}, ccv2.Warnings{"some-warnings"}, errors.New("some-error"))
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warnings"))
			})
		})
	})

	Describe("BindServiceBySpace", func() {
		var (
			executeErr     error
			warnings       Warnings
			serviceBinding ServiceBinding
		)

		JustBeforeEach(func() {
			serviceBinding, warnings, executeErr = actor.BindServiceBySpace("some-app-name", "some-service-instance-name", "some-space-guid", "some-binding-name", map[string]interface{}{"some-parameter": "some-value"})
		})

		When("getting the application errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv2.Warnings{"foo-1"},
					errors.New("some-error"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(warnings).To(ConsistOf("foo-1"))
			})
		})

		When("getting the application succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{{GUID: "some-app-guid"}},
					ccv2.Warnings{"foo-1"},
					nil,
				)
			})

			When("getting the service instance errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{},
						ccv2.Warnings{"foo-2"},
						errors.New("some-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(errors.New("some-error")))
					Expect(warnings).To(ConsistOf("foo-1", "foo-2"))
				})
			})

			When("getting the service instance succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{{GUID: "some-service-instance-guid"}},
						ccv2.Warnings{"foo-2"},
						nil,
					)
				})

				When("binding the service instance to the application errors", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CreateServiceBindingReturns(
							ccv2.ServiceBinding{},
							ccv2.Warnings{"foo-3"},
							errors.New("some-error"),
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(errors.New("some-error")))
						Expect(warnings).To(ConsistOf("foo-1", "foo-2", "foo-3"))
					})
				})

				When("binding the service instance to the application succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.CreateServiceBindingReturns(
							ccv2.ServiceBinding{GUID: "some-service-binding-guid"},
							ccv2.Warnings{"foo-3"},
							nil,
						)
					})

					It("returns all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("foo-1", "foo-2", "foo-3"))
						Expect(serviceBinding).To(Equal(ServiceBinding{GUID: "some-service-binding-guid"}))

						Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))

						Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))

						Expect(fakeCloudControllerClient.CreateServiceBindingCallCount()).To(Equal(1))
						appGUID, serviceInstanceGUID, bindingName, acceptsIncomplete, parameters := fakeCloudControllerClient.CreateServiceBindingArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(serviceInstanceGUID).To(Equal("some-service-instance-guid"))
						Expect(bindingName).To(Equal("some-binding-name"))
						Expect(acceptsIncomplete).To(BeTrue())
						Expect(parameters).To(Equal(map[string]interface{}{"some-parameter": "some-value"}))
					})
				})
			})
		})
	})

	Describe("GetServiceBindingByApplicationAndServiceInstance", func() {
		When("the service binding exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBindingsReturns(
					[]ccv2.ServiceBinding{
						{
							GUID: "some-service-binding-guid",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service binding and warnings", func() {
				serviceBinding, warnings, err := actor.GetServiceBindingByApplicationAndServiceInstance("some-app-guid", "some-service-instance-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceBinding).To(Equal(ServiceBinding{
					GUID: "some-service-binding-guid",
				}))
				Expect(warnings).To(ConsistOf("foo"))

				Expect(fakeCloudControllerClient.GetServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceBindingsArgsForCall(0)).To(ConsistOf([]ccv2.Filter{
					ccv2.Filter{
						Type:     constant.AppGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-app-guid"},
					},
					ccv2.Filter{
						Type:     constant.ServiceInstanceGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-service-instance-guid"},
					},
				}))
			})
		})

		When("the service binding does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBindingsReturns([]ccv2.ServiceBinding{}, nil, nil)
			})

			It("returns a ServiceBindingNotFoundError", func() {
				_, _, err := actor.GetServiceBindingByApplicationAndServiceInstance("some-app-guid", "some-service-instance-guid")
				Expect(err).To(MatchError(actionerror.ServiceBindingNotFoundError{
					AppGUID:             "some-app-guid",
					ServiceInstanceGUID: "some-service-instance-guid",
				}))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetServiceBindingsReturns([]ccv2.ServiceBinding{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetServiceBindingByApplicationAndServiceInstance("some-app-guid", "some-service-instance-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("UnbindServiceBySpace", func() {
		var (
			executeErr     error
			warnings       Warnings
			serviceBinding ServiceBinding
		)

		JustBeforeEach(func() {
			serviceBinding, warnings, executeErr = actor.UnbindServiceBySpace("some-app", "some-service-instance", "some-space-guid")
		})

		When("the service binding exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "some-app-guid",
							Name: "some-app",
						},
					},
					ccv2.Warnings{"foo-1"},
					nil,
				)
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					ccv2.Warnings{"foo-2"},
					nil,
				)
				fakeCloudControllerClient.GetServiceBindingsReturns(
					[]ccv2.ServiceBinding{
						{
							GUID: "some-service-binding-guid",
						},
					},
					ccv2.Warnings{"foo-3"},
					nil,
				)

				fakeCloudControllerClient.DeleteServiceBindingReturns(
					ccv2.ServiceBinding{GUID: "deleted-service-binding-guid"},
					ccv2.Warnings{"foo-4", "foo-5"},
					nil,
				)
			})

			It("deletes the service binding", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"foo-1", "foo-2", "foo-3", "foo-4", "foo-5"}))
				Expect(serviceBinding).To(Equal(ServiceBinding{GUID: "deleted-service-binding-guid"}))

				Expect(fakeCloudControllerClient.DeleteServiceBindingCallCount()).To(Equal(1))
				passedGUID, acceptsIncomplete := fakeCloudControllerClient.DeleteServiceBindingArgsForCall(0)
				Expect(passedGUID).To(Equal("some-service-binding-guid"))
				Expect(acceptsIncomplete).To(BeTrue())
			})

			When("the cloud controller API returns warnings and an error", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = errors.New("I am a CC error")
					fakeCloudControllerClient.DeleteServiceBindingReturns(ccv2.ServiceBinding{}, ccv2.Warnings{"foo-4", "foo-5"}, expectedError)
				})

				It("returns the warnings and the error", func() {
					Expect(executeErr).To(MatchError(expectedError))
					Expect(warnings).To(ConsistOf(Warnings{"foo-1", "foo-2", "foo-3", "foo-4", "foo-5"}))
				})
			})
		})
	})

	Describe("GetServiceBindingsByServiceInstance", func() {
		var (
			serviceBindings         []ServiceBinding
			serviceBindingsWarnings Warnings
			serviceBindingsErr      error
		)

		JustBeforeEach(func() {
			serviceBindings, serviceBindingsWarnings, serviceBindingsErr = actor.GetServiceBindingsByServiceInstance("some-service-instance-guid")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
					[]ccv2.ServiceBinding{
						{GUID: "some-service-binding-1-guid"},
						{GUID: "some-service-binding-2-guid"},
					},
					ccv2.Warnings{"get-service-bindings-warning"},
					nil,
				)
			})

			It("returns service bindings and all warnings", func() {
				Expect(serviceBindingsErr).ToNot(HaveOccurred())
				Expect(serviceBindings).To(ConsistOf(
					ServiceBinding{GUID: "some-service-binding-1-guid"},
					ServiceBinding{GUID: "some-service-binding-2-guid"},
				))
				Expect(serviceBindingsWarnings).To(ConsistOf("get-service-bindings-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})

		When("an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get service bindings error")
				fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
					[]ccv2.ServiceBinding{},
					ccv2.Warnings{"get-service-bindings-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(serviceBindingsErr).To(MatchError(expectedErr))
				Expect(serviceBindingsWarnings).To(ConsistOf("get-service-bindings-warning"))

				Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal("some-service-instance-guid"))
			})
		})
	})

	Describe("GetServiceBindingsByUserProvidedServiceInstance", func() {
		var (
			serviceBindings         []ServiceBinding
			serviceBindingsWarnings Warnings
			serviceBindingsErr      error
		)

		JustBeforeEach(func() {
			serviceBindings, serviceBindingsWarnings, serviceBindingsErr = actor.GetServiceBindingsByUserProvidedServiceInstance("some-user-provided-service-instance-guid")
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
					[]ccv2.ServiceBinding{
						{GUID: "some-service-binding-1-guid"},
						{GUID: "some-service-binding-2-guid"},
					},
					ccv2.Warnings{"get-service-bindings-warning"},
					nil,
				)
			})

			It("returns service bindings and all warnings", func() {
				Expect(serviceBindingsErr).ToNot(HaveOccurred())
				Expect(serviceBindings).To(ConsistOf(
					ServiceBinding{GUID: "some-service-binding-1-guid"},
					ServiceBinding{GUID: "some-service-binding-2-guid"},
				))
				Expect(serviceBindingsWarnings).To(ConsistOf("get-service-bindings-warning"))

				Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsArgsForCall(0)).To(Equal("some-user-provided-service-instance-guid"))
			})
		})

		When("an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get service bindings error")
				fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
					[]ccv2.ServiceBinding{},
					ccv2.Warnings{"get-service-bindings-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(serviceBindingsErr).To(MatchError(expectedErr))
				Expect(serviceBindingsWarnings).To(ConsistOf("get-service-bindings-warning"))

				Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsArgsForCall(0)).To(Equal("some-user-provided-service-instance-guid"))
			})
		})
	})
})
