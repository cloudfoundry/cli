package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
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

	Describe("BindServiceBySpace", func() {
		var (
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.BindServiceBySpace("some-app-name", "some-service-instance-name", "some-space-guid", map[string]interface{}{"some-parameter": "some-value"})
		})

		Context("when getting the application errors", func() {
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

		Context("when getting the application succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{{GUID: "some-app-guid"}},
					ccv2.Warnings{"foo-1"},
					nil,
				)
			})

			Context("when getting the service instance errors", func() {
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

			Context("when getting the service instance succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{{GUID: "some-service-instance-guid"}},
						ccv2.Warnings{"foo-2"},
						nil,
					)
				})

				Context("when getting binding the service instance to the application errors", func() {
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
				Context("when getting binding the service instance to the application succeeds", func() {
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

						Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))

						Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))

						Expect(fakeCloudControllerClient.CreateServiceBindingCallCount()).To(Equal(1))
						appGUID, serviceInstanceGUID, parameters := fakeCloudControllerClient.CreateServiceBindingArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(serviceInstanceGUID).To(Equal("some-service-instance-guid"))
						Expect(parameters).To(Equal(map[string]interface{}{"some-parameter": "some-value"}))
					})
				})
			})
		})
	})

	Describe("GetServiceBindingByApplicationAndServiceInstance", func() {
		Context("when the service binding exists", func() {
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
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetServiceBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceBindingsArgsForCall(0)).To(ConsistOf([]ccv2.Query{
					ccv2.Query{
						Filter:   ccv2.AppGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-app-guid",
					},
					ccv2.Query{
						Filter:   ccv2.ServiceInstanceGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-service-instance-guid",
					},
				}))
			})
		})

		Context("when the service binding does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBindingsReturns([]ccv2.ServiceBinding{}, nil, nil)
			})

			It("returns a ServiceBindingNotFoundError", func() {
				_, _, err := actor.GetServiceBindingByApplicationAndServiceInstance("some-app-guid", "some-service-instance-guid")
				Expect(err).To(MatchError(ServiceBindingNotFoundError{
					AppGUID:             "some-app-guid",
					ServiceInstanceGUID: "some-service-instance-guid",
				}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
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
		Context("when the service binding exists", func() {
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
					ccv2.Warnings{"foo-4", "foo-5"},
					nil,
				)
			})

			It("deletes the service binding", func() {
				warnings, err := actor.UnbindServiceBySpace("some-app", "some-service-instance", "some-space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"foo-1", "foo-2", "foo-3", "foo-4", "foo-5"}))

				Expect(fakeCloudControllerClient.DeleteServiceBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServiceBindingArgsForCall(0)).To(Equal("some-service-binding-guid"))
			})

			Context("when the cloud controller API returns warnings and an error", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = errors.New("I am a CC error")
					fakeCloudControllerClient.DeleteServiceBindingReturns(ccv2.Warnings{"foo-4", "foo-5"}, expectedError)
				})

				It("returns the warnings and the error", func() {
					warnings, err := actor.UnbindServiceBySpace("some-app", "some-service-instance", "some-space-guid")
					Expect(err).To(MatchError(expectedError))
					Expect(warnings).To(ConsistOf(Warnings{"foo-1", "foo-2", "foo-3", "foo-4", "foo-5"}))
				})
			})
		})
	})
})
