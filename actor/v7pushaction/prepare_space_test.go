package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrepareSpace", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		pushPlans          []PushPlan
		fakeManifestParser *v7pushactionfakes.FakeManifestParser

		spaceGUID string

		appNames    []string
		eventStream <-chan *PushEvent
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		spaceGUID = "space"

		fakeManifestParser = new(v7pushactionfakes.FakeManifestParser)
	})

	AfterEach(func() {
		Eventually(streamsDrainedAndClosed(eventStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		appNames, eventStream = actor.PrepareSpace(pushPlans, fakeManifestParser)
	})

	When("there is a single push state and no manifest", func() {
		var appName = "app-name"

		When("Creating the app succeeds", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{Name: appName},
					v7action.Warnings{"create-app-warnings"},
					nil,
				)
			})

			It("returns the app names from the push plans", func() {
				Expect(appNames).To(Equal([]string{appName}))
			})

			It("creates the app using the API", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: CreatingApplication, Plan: pushPlans[0]})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGUID := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp).To(Equal(v7action.Application{Name: appName}))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Event:    CreatedApplication,
					Warnings: Warnings{"create-app-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}},
				})))
			})
		})

		When("the app already exists", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"create-app-warnings"},
					actionerror.ApplicationAlreadyExistsError{},
				)
			})
			It("Sends already exists events", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: SkippingApplicationCreation, Plan: pushPlans[0]})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGUID := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp).To(Equal(v7action.Application{Name: appName}))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Event:    ApplicationAlreadyExists,
					Warnings: Warnings{"create-app-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}},
				})))
			})
		})

		When("creating the app fails", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"create-app-warnings"},
					errors.New("some-create-error"),
				)
			})

			It("Returns the error", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: CreatingApplication, Plan: pushPlans[0]})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGuid := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp.Name).To(Equal(appName))
				Expect(actualSpaceGuid).To(Equal(spaceGUID))

				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Warnings: Warnings{"create-app-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}},
					Err:      errors.New("some-create-error"),
				})))
			})
		})
	})

	When("There is a a manifest", func() {
		var (
			manifest = []byte("app manifest")
			appName1 = "app-name1"
			appName2 = "app-name2"
		)
		When("applying the manifest fails", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}}}
				fakeManifestParser.ContainsManifestReturns(true)
				fakeManifestParser.RawAppManifestReturns(manifest, nil)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, errors.New("some-error"))
			})

			It("returns the error and exits", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: ApplyManifest})))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGuid, actualManifest, _ := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualSpaceGuid).To(Equal(spaceGUID))
				Expect(actualManifest).To(Equal(manifest))

				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Warnings: Warnings{"apply-manifest-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
					Err:      errors.New("some-error"),
				})))
			})
		})

		When("There is a single pushPlan", func() {

			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}, NoRouteFlag: true}}
				fakeManifestParser.ContainsManifestReturns(true)
				fakeManifestParser.RawAppManifestReturns(manifest, nil)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
			})

			It("applies the app specific manifest", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: ApplyManifest})))
				Eventually(fakeManifestParser.RawAppManifestCallCount).Should(Equal(1))
				actualAppName := fakeManifestParser.RawAppManifestArgsForCall(0)
				Expect(actualAppName).To(Equal(appName1))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGUID, actualManifest, actualNoRoute := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualManifest).To(Equal(manifest))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualNoRoute).To(BeTrue())

				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Event:    ApplyManifestComplete,
					Warnings: Warnings{"apply-manifest-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}, NoRouteFlag: true},
				})))
			})
		})

		When("There are multiple push states", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{
					{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
					{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName2}},
				}
				fakeManifestParser.ContainsManifestReturns(true)
				fakeManifestParser.FullRawManifestReturns(manifest)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
			})

			It("returns the app names from the push plans", func() {
				Expect(appNames).To(Equal([]string{appName1, appName2}))
			})

			It("Applies the entire manifest", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Consistently(fakeManifestParser.RawAppManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Event: ApplyManifest})))
				Eventually(fakeManifestParser.FullRawManifestCallCount).Should(Equal(1))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGUID, actualManifest, actualNoRoute := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualManifest).To(Equal(manifest))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualNoRoute).To(BeFalse())

				Eventually(eventStream).Should(Receive(Equal(&PushEvent{
					Event:    ApplyManifestComplete,
					Warnings: Warnings{"apply-manifest-warnings"},
					Plan:     PushPlan{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
				})))
			})
		})
	})
})
