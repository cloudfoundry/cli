package v7pushaction_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Push plan", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor
		pwd         string
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, _ = getTestPushActor()
	})

	Describe("Conceptualize", func() {
		var (
			paramPlans []PushPlan

			returnedPlans []PushPlan
			warnings      Warnings
			executeErr    error

			spaceGUID string
		)

		BeforeEach(func() {
			spaceGUID = "space"

			var err error
			pwd, err = os.Getwd()
			Expect(err).To(Not(HaveOccurred()))
			paramPlans = []PushPlan{
				{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-0"}, BitsPath: pwd},
				{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-1"}, BitsPath: pwd},
				{SpaceGUID: spaceGUID, Application: v7action.Application{Name: "app-name-2"}, BitsPath: pwd},
			}
		})

		JustBeforeEach(func() {
			returnedPlans, warnings, executeErr = actor.Conceptualize(paramPlans)
		})

		Describe("application", func() {
			When("the applications exist", func() {
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
				})

				It("returns multiple push returnedPlans", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("plural-get-warning"))
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
				})
			})
		})
	})
})
