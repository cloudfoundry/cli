package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
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
				}, ccv3.Warnings{"some-service-instance-warning"}, nil)
			})

			It("returns a service instance and warnings", func() {
				Expect(executionError).NotTo(HaveOccurred())

				Expect(serviceInstance).To(Equal(resources.ServiceInstance{Name: "some-service-instance", GUID: "some-service-instance-guid"}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualName, actualSpaceGUID := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("no service instance"))
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
				actualName, actualSpaceGUID := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(originalName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))

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
})
