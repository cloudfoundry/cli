package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("UpdateApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		returnedPlan PushPlan
		warnings     Warnings
		executeErr   error
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		returnedPlan, warnings, executeErr = actor.UpdateApplication(paramPlan, nil, nil)
	})

	When("the apps needs an update", func() {
		BeforeEach(func() {
			paramPlan.ApplicationNeedsUpdate = true
		})

		When("updating is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.UpdateApplicationReturns(
					v7action.Application{
						Name:                "some-app",
						GUID:                "some-app-guid",
						LifecycleBuildpacks: []string{"some-buildpack-1"},
					},
					v7action.Warnings{"some-app-update-warnings"},
					nil)
			})

			It("puts the updated application in the stream", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-update-warnings"))

				Expect(returnedPlan).To(MatchFields(IgnoreExtras,
					Fields{
						"Application": Equal(v7action.Application{
							Name:                "some-app",
							GUID:                "some-app-guid",
							LifecycleBuildpacks: []string{"some-buildpack-1"},
						}),
					}))
			})
		})

		When("updating errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeV7Actor.UpdateApplicationReturns(
					v7action.Application{},
					v7action.Warnings{"some-app-update-warnings"},
					expectedErr)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-update-warnings"))
			})
		})
	})
})
