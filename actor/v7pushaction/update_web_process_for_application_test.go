package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("UpdateWebProcessForApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			_, warnings, executeErr = actor.UpdateWebProcessForApplication(paramPlan, eventStream, nil)
		})
	})

	When("process configuration is provided", func() {
		var startCommand types.FilteredString

		BeforeEach(func() {
			paramPlan.UpdateWebProcessNeedsUpdate = true

			startCommand = types.FilteredString{IsSet: true, Value: "some-start-command"}
			paramPlan.UpdateWebProcess = v7action.Process{
				Command: startCommand,
			}
		})

		When("the update is successful", func() {
			BeforeEach(func() {
				paramPlan.Application.GUID = "some-app-guid"

				fakeV7Actor.UpdateApplicationReturns(
					v7action.Application{
						Name: "some-app",
						GUID: paramPlan.Application.GUID,
					},
					v7action.Warnings{"some-app-update-warnings"},
					nil)

				fakeV7Actor.UpdateProcessByTypeAndApplicationReturns(v7action.Warnings{"health-check-warnings"}, nil)
			})

			It("sets the process config and returns warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("health-check-warnings"))
				Expect(events).To(ConsistOf(SetProcessConfiguration, SetProcessConfigurationComplete))

				Expect(fakeV7Actor.UpdateProcessByTypeAndApplicationCallCount()).To(Equal(1))
				passedProcessType, passedAppGUID, passedProcess := fakeV7Actor.UpdateProcessByTypeAndApplicationArgsForCall(0)
				Expect(passedProcessType).To(Equal(constant.ProcessTypeWeb))
				Expect(passedAppGUID).To(Equal("some-app-guid"))
				Expect(passedProcess).To(MatchFields(IgnoreExtras,
					Fields{
						"Command": Equal(startCommand),
					}))
			})
		})

		When("the update errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("nopes")
				fakeV7Actor.UpdateProcessByTypeAndApplicationReturns(v7action.Warnings{"health-check-warnings"}, expectedErr)
			})

			It("returns warnings and an error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("health-check-warnings"))
				Expect(events).To(ConsistOf(SetProcessConfiguration))
			})
		})
	})
})
