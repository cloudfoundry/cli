package v2action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

var _ = Describe("Service Key", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateServiceKey", func() {
		var (
			executeErr error
			warnings   Warnings
			serviceKey ServiceKey
		)

		JustBeforeEach(func() {
			serviceKey, warnings, executeErr = actor.CreateServiceKey("some-service-instance-name", "some-key-name", "some-space-guid", map[string]interface{}{"some-parameter": "some-value"})
		})

		It("looks for the service instance with the given name", func() {
			Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
			spaceGuid, includedUserProvided, filters := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
			Expect(spaceGuid).To(Equal("some-space-guid"))
			Expect(includedUserProvided).To(BeTrue())
			Expect(filters).To(ConsistOf(
				ccv2.Filter{Type: constant.NameFilter, Operator: constant.EqualOperator, Values: []string{"some-service-instance-name"}},
			))
		})

		When("getting the service instance errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					nil,
					ccv2.Warnings{"warning-1"},
					errors.New("some-error"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(errors.New("some-error")))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("getting the service instance succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{{GUID: "some-service-instance-guid"}},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("creates the service key with the correct values", func() {
				Expect(fakeCloudControllerClient.CreateServiceKeyCallCount()).To(Equal(1))
				serviceInstanceGuid, keyName, parameters := fakeCloudControllerClient.CreateServiceKeyArgsForCall(0)
				Expect(keyName).To(Equal("some-key-name"))
				Expect(serviceInstanceGuid).To(Equal("some-service-instance-guid"))
				Expect(parameters).To(Equal(map[string]interface{}{"some-parameter": "some-value"}))
			})

			When("creating the service key errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceKeyReturns(
						ccv2.ServiceKey{},
						ccv2.Warnings{"warning-2"},
						errors.New("some-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(errors.New("some-error")))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			When("creating the service key succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceKeyReturns(
						ccv2.ServiceKey{GUID: "service-key-guid"},
						ccv2.Warnings{"warning-2"},
						errors.New("some-error"),
					)
				})

				It("returns the warnings", func() {
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})

				It("returns the created service key", func() {
					Expect(serviceKey.GUID).To(Equal("service-key-guid"))
				})
			})
		})

		When("getting the service instance doesn't return any service instances", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns a ServiceInstanceNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: "some-service-instance-name"}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})
})
