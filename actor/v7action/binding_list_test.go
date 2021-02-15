package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Binding List", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetBindingsByServiceInstance", func() {
		const (
			fakeSpaceGUID           = "fake-space-guid"
			fakeServiceInstanceName = "fake-service-instance-name"
			fakeServiceInstanceGUID = "fake-service-instance-guid"
		)

		var (
			params     BindingListParameters
			list       BindingList
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{GUID: fakeServiceInstanceGUID},
				ccv3.IncludedResources{},
				ccv3.Warnings{"si warning"},
				nil,
			)

			fakeCloudControllerClient.GetRouteBindingsReturns(
				[]resources.RouteBinding{
					{GUID: "rb-guid-1", RouteGUID: "r-guid-1", LastOperation: resources.LastOperation{Type: "create", State: "successful"}},
					{GUID: "rb-guid-2", RouteGUID: "r-guid-2", LastOperation: resources.LastOperation{Type: "update", State: "successful"}},
					{GUID: "rb-guid-3", RouteGUID: "r-guid-2", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				},
				ccv3.IncludedResources{
					Routes: []resources.Route{
						{GUID: "r-guid-1", URL: "https://foo.com/lala:8080"},
						{GUID: "r-guid-2", URL: "https://bar.org/fifi"},
					},
				},
				ccv3.Warnings{"rb warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{Type: resources.AppBinding, GUID: "a-1", AppName: "app-1", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
					{Type: resources.KeyBinding, GUID: "k-1", Name: "key-1", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
					{Type: resources.AppBinding, GUID: "a-2", AppName: "app-2", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
					{Type: resources.KeyBinding, GUID: "k-2", Name: "key-2", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
					{Type: resources.KeyBinding, GUID: "k-3", Name: "key-3", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
					{Type: resources.AppBinding, GUID: "a-3", AppName: "app-3", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				},
				ccv3.Warnings{"scb warning"},
				nil,
			)

			params = BindingListParameters{
				ServiceInstanceName: fakeServiceInstanceName,
				SpaceGUID:           fakeSpaceGUID,
				GetAppBindings:      true,
				GetServiceKeys:      true,
				GetRouteBindings:    true,
			}
		})

		JustBeforeEach(func() {
			list, warnings, executeErr = actor.GetBindingsByServiceInstance(params)
		})

		It("returns a list and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("si warning", "rb warning", "scb warning"))

			Expect(list).To(Equal(BindingList{
				App: []resources.ServiceCredentialBinding{
					{Type: resources.AppBinding, GUID: "a-1", AppName: "app-1", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
					{Type: resources.AppBinding, GUID: "a-2", AppName: "app-2", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
					{Type: resources.AppBinding, GUID: "a-3", AppName: "app-3", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				},
				Key: []resources.ServiceCredentialBinding{
					{Type: resources.KeyBinding, GUID: "k-1", Name: "key-1", LastOperation: resources.LastOperation{Type: "create", State: "succeeded"}},
					{Type: resources.KeyBinding, GUID: "k-2", Name: "key-2", LastOperation: resources.LastOperation{Type: "update", State: "succeeded"}},
					{Type: resources.KeyBinding, GUID: "k-3", Name: "key-3", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				},
				Route: []RouteBindingSummary{
					{URL: "https://foo.com/lala:8080", LastOperation: resources.LastOperation{Type: "create", State: "successful"}},
					{URL: "https://bar.org/fifi", LastOperation: resources.LastOperation{Type: "update", State: "successful"}},
					{URL: "https://bar.org/fifi", LastOperation: resources.LastOperation{Type: "create", State: "failed"}},
				},
			}))
		})

		It("gets the service instance GUID", func() {
			Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			actualServiceInstanceName, actualSpaceGUID, query := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(actualServiceInstanceName).To(Equal(fakeServiceInstanceName))
			Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
			Expect(query).To(BeEmpty())
		})

		It("gets the route bindings and associated routes", func() {
			Expect(fakeCloudControllerClient.GetRouteBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetRouteBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{fakeServiceInstanceGUID}},
				ccv3.Query{Key: ccv3.Include, Values: []string{"route"}},
			))
		})

		Context("route bindings not requested", func() {
			BeforeEach(func() {
				params.GetRouteBindings = false
			})

			It("does not get the route bindings", func() {
				Expect(fakeCloudControllerClient.GetRouteBindingsCallCount()).To(Equal(0))
			})
		})

		It("gets the service credential bindings", func() {
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{fakeServiceInstanceGUID}},
			))
		})

		Context("service keys not requested", func() {
			BeforeEach(func() {
				params.GetServiceKeys = false
			})

			It("gets just the app bindings", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{fakeServiceInstanceGUID}},
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}},
				))
			})
		})

		Context("app bindings not requested", func() {
			BeforeEach(func() {
				params.GetAppBindings = false
			})

			It("gets just the service keys", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{fakeServiceInstanceGUID}},
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				))
			})
		})

		Context("service keys and app bindings not requested", func() {
			BeforeEach(func() {
				params.GetAppBindings = false
				params.GetServiceKeys = false
			})

			It("does not get credential bindings", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(0))
			})
		})

		Context("errors", func() {
			When("getting service instance", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"si warning"},
						errors.New("si error"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(list).To(Equal(BindingList{}))
					Expect(warnings).To(ContainElement("si warning"))
					Expect(executeErr).To(MatchError("si error"))
				})
			})

			When("service instance not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"si warning"},
						ccerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName},
					)
				})

				It("returns the error and warnings", func() {
					Expect(list).To(Equal(BindingList{}))
					Expect(warnings).To(ContainElement("si warning"))
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
						Name: fakeServiceInstanceName,
					}))
				})
			})

			When("getting route bindings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRouteBindingsReturns(
						[]resources.RouteBinding{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"rb warning"},
						errors.New("rb error"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(list).To(Equal(BindingList{}))
					Expect(warnings).To(ContainElement("rb warning"))
					Expect(executeErr).To(MatchError("rb error"))
				})
			})

			When("getting service credential bindings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"scb warning"},
						errors.New("scb error"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(list).To(Equal(BindingList{}))
					Expect(warnings).To(ContainElement("scb warning"))
					Expect(executeErr).To(MatchError("scb error"))
				})
			})
		})
	})
})
