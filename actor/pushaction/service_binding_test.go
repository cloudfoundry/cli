package pushaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Binding Services", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor, nil)
	})

	Describe("BindServices", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			boundServices  bool
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			config = ApplicationConfig{}
			config.DesiredApplication.GUID = "some-app-guid"
			config.CurrentServices = map[string]v2action.ServiceInstance{"service_instance_1": {GUID: "instance_1_guid"}}
			config.DesiredServices = map[string]v2action.ServiceInstance{
				"service_instance_1": {GUID: "instance_1_guid"},
				"service_instance_2": {GUID: "instance_2_guid"},
				"service_instance_3": {GUID: "instance_3_guid"},
			}
		})

		JustBeforeEach(func() {
			returnedConfig, boundServices, warnings, executeErr = actor.BindServices(config)
		})

		Context("when binding services is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturnsOnCall(0, v2action.Warnings{"service-instance-warning-1"}, nil)
				fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturnsOnCall(1, v2action.Warnings{"service-instance-warning-2"}, nil)
			})

			It("it updates CurrentServices to match DesiredServices", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("service-instance-warning-1", "service-instance-warning-2"))
				Expect(boundServices).To(BeTrue())
				Expect(returnedConfig.CurrentServices).To(Equal(map[string]v2action.ServiceInstance{
					"service_instance_1": {GUID: "instance_1_guid"},
					"service_instance_2": {GUID: "instance_2_guid"},
					"service_instance_3": {GUID: "instance_3_guid"},
				}))

				var serviceInstanceGUIDs []string
				Expect(fakeV2Actor.BindServiceByApplicationAndServiceInstanceCallCount()).To(Equal(2))
				appGUID, serviceInstanceGUID := fakeV2Actor.BindServiceByApplicationAndServiceInstanceArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				serviceInstanceGUIDs = append(serviceInstanceGUIDs, serviceInstanceGUID)

				appGUID, serviceInstanceGUID = fakeV2Actor.BindServiceByApplicationAndServiceInstanceArgsForCall(1)
				Expect(appGUID).To(Equal("some-app-guid"))
				serviceInstanceGUIDs = append(serviceInstanceGUIDs, serviceInstanceGUID)

				Expect(serviceInstanceGUIDs).To(ConsistOf("instance_2_guid", "instance_3_guid"))
			})
		})

		Context("when binding services fails", func() {
			BeforeEach(func() {
				fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"service-instance-warning-1"}, errors.New("some-error"))
			})

			It("it returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("service-instance-warning-1"))
				Expect(boundServices).To(BeFalse())
			})
		})
	})
})
