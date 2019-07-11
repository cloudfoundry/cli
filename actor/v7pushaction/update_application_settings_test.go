package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateApplicationSettings", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlans []PushPlan

		returnedPlans []PushPlan
		warnings      Warnings
		executeErr    error

		spaceGUID string
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		spaceGUID = "space"

		paramPlans = []PushPlan{
			{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-0"}},
			{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-1"}},
			{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-2"}},
		}
	})

	JustBeforeEach(func() {
		returnedPlans, warnings, executeErr = actor.UpdateApplicationSettings(paramPlans)
	})

	Describe("application", func() {
		var apps []v7action.Application
		BeforeEach(func() {
			apps = []v7action.Application{
				{
					GUID:  "app-guid-0",
					Name:  "app-name-0",
					State: constant.ApplicationStarted,
				},
				{
					GUID:  "app-guid-1",
					Name:  "app-name-1",
					State: constant.ApplicationStarted,
				},
				{
					GUID:  "app-guid-2",
					Name:  "app-name-2",
					State: constant.ApplicationStopped,
				},
			}
			fakeV7Actor.GetApplicationsByNamesAndSpaceReturns(apps, v7action.Warnings{"plural-get-warning"}, nil)

			fakeV7Actor.GetApplicationRoutesReturnsOnCall(0,
				[]v7action.Route{{GUID: "app-0-route"}}, v7action.Warnings{"get-route-warning-0"}, nil,
			)
			fakeV7Actor.GetApplicationRoutesReturnsOnCall(1,
				[]v7action.Route{{GUID: "app-1-route"}}, v7action.Warnings{"get-route-warning-1"}, nil,
			)
			fakeV7Actor.GetApplicationRoutesReturnsOnCall(2,
				[]v7action.Route{{GUID: "app-2-route"}}, v7action.Warnings{"get-route-warning-2"}, nil,
			)

		})

		When("the applications exist", func() {
			It("returns multiple push returnedPlans", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("plural-get-warning", "get-route-warning-0", "get-route-warning-1", "get-route-warning-2"))
				Expect(returnedPlans).To(HaveLen(3))

				Expect(returnedPlans[0].Application.Name).To(Equal("app-name-0"))
				Expect(returnedPlans[1].Application.Name).To(Equal("app-name-1"))
				Expect(returnedPlans[2].Application.Name).To(Equal("app-name-2"))
				Expect(returnedPlans[0].Application.GUID).To(Equal("app-guid-0"))
				Expect(returnedPlans[1].Application.GUID).To(Equal("app-guid-1"))
				Expect(returnedPlans[2].Application.GUID).To(Equal("app-guid-2"))
				Expect(returnedPlans[0].Application.State).To(Equal(constant.ApplicationStarted))
				Expect(returnedPlans[1].Application.State).To(Equal(constant.ApplicationStarted))
				Expect(returnedPlans[2].Application.State).To(Equal(constant.ApplicationStopped))
				Expect(returnedPlans[0].ApplicationRoutes).To(Equal(
					[]v7action.Route{{GUID: "app-0-route"}},
				))
				Expect(returnedPlans[1].ApplicationRoutes).To(Equal(
					[]v7action.Route{{GUID: "app-1-route"}},
				))
				Expect(returnedPlans[2].ApplicationRoutes).To(Equal(
					[]v7action.Route{{GUID: "app-2-route"}},
				))

				Expect(fakeV7Actor.GetApplicationRoutesCallCount()).To(Equal(3))
				Expect(fakeV7Actor.GetApplicationRoutesArgsForCall(0)).To(Equal("app-guid-0"))
				Expect(fakeV7Actor.GetApplicationRoutesArgsForCall(1)).To(Equal("app-guid-1"))
				Expect(fakeV7Actor.GetApplicationRoutesArgsForCall(2)).To(Equal("app-guid-2"))

				Expect(fakeV7Actor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(1))
				argumentAppNames, argumentSpaceGUID := fakeV7Actor.GetApplicationsByNamesAndSpaceArgsForCall(0)
				Expect(argumentAppNames).To(ConsistOf("app-name-0", "app-name-1", "app-name-2"))
				Expect(argumentSpaceGUID).To(Equal(spaceGUID))
			})
		})

		When("getting the app errors", func() {
			BeforeEach(func() {
				fakeV7Actor.GetApplicationsByNamesAndSpaceReturns([]v7action.Application{}, v7action.Warnings{"some-app-warning"}, errors.New("bad-get-error"))
			})

			It("errors", func() {
				Expect(executeErr).To(MatchError(errors.New("bad-get-error")))
				Expect(warnings).To(ConsistOf("some-app-warning"))

				Expect(fakeV7Actor.GetApplicationRoutesCallCount()).To(Equal(0))
			})
		})

		When("getting the routes errors", func() {
			BeforeEach(func() {
				fakeV7Actor.GetApplicationRoutesReturnsOnCall(0,
					[]v7action.Route{{GUID: "app-0-route"}}, v7action.Warnings{"get-route-warning-0"}, nil,
				)
				fakeV7Actor.GetApplicationRoutesReturnsOnCall(1,
					[]v7action.Route{{GUID: "app-1-route"}}, v7action.Warnings{"get-route-warning-1"}, errors.New("get-routes-error"),
				)
			})

			It("errors", func() {
				Expect(executeErr).To(MatchError(errors.New("get-routes-error")))
				Expect(warnings).To(ConsistOf("plural-get-warning", "get-route-warning-0", "get-route-warning-1"))

				Expect(fakeV7Actor.GetApplicationRoutesCallCount()).To(Equal(2))
				Expect(fakeV7Actor.GetApplicationRoutesArgsForCall(0)).To(Equal("app-guid-0"))
				Expect(fakeV7Actor.GetApplicationRoutesArgsForCall(1)).To(Equal("app-guid-1"))
			})
		})
	})
})
