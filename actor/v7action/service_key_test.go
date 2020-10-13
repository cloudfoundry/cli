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

var _ = Describe("Service Key Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
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
					{Name: "flopsy"},
					{Name: "mopsy"},
					{Name: "cottontail"},
					{Name: "peter"},
				},
				ccv3.Warnings{"get keys warning"},
				nil,
			)
		})

		var (
			names          []string
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			names, warnings, executionError = actor.GetServiceKeysByServiceInstance(serviceInstanceName, spaceGUID)
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
			Expect(names).To(Equal([]string{"flopsy", "mopsy", "cottontail", "peter"}))
		})

		When("service instance is user-provided", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{
						Name: serviceInstanceName,
						GUID: serviceInstanceGUID,
						Type: resources.UserProvidedServiceInstance,
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get instance warning"},
					nil,
				)
			})

			It("returns an error and warning", func() {
				Expect(warnings).To(ContainElement("get instance warning"))
				Expect(executionError).To(MatchError(actionerror.ServiceInstanceTypeError{
					Name:         serviceInstanceName,
					RequiredType: resources.ManagedServiceInstance,
				}))
			})
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
})
