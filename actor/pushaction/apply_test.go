package pushaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apply", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor

		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error

		config ApplicationConfig
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)

		config = ApplicationConfig{
			DesiredApplication: v2action.Application{
				Name:      "some-app-name",
				SpaceGUID: "some-space-guid",
			},
		}
	})

	JustBeforeEach(func() {
		eventStream, warningsStream, errorStream = actor.Apply(config)
	})

	AfterEach(func() {
		Eventually(warningsStream).Should(BeClosed())
		Eventually(eventStream).Should(BeClosed())
		Eventually(errorStream).Should(BeClosed())
	})

	Context("when the app exists", func() {
		BeforeEach(func() {
			config.CurrentApplication = v2action.Application{
				Name:      "some-app-name",
				GUID:      "some-app-guid",
				SpaceGUID: "some-space-guid",
				Buildpack: "java",
			}
			config.DesiredApplication = v2action.Application{
				Name:      "some-app-name",
				GUID:      "some-app-guid",
				SpaceGUID: "some-space-guid",
				Buildpack: "ruby",
			}
		})

		Context("when the update is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.UpdateApplicationReturns(v2action.Application{}, v2action.Warnings{"update-warning"}, nil)
			})

			It("updates the application", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("update-warning")))
				Eventually(eventStream).Should(Receive(Equal(ApplicationUpdated)))
				Eventually(eventStream).Should(Receive(Equal(Complete)))

				Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
				Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(Equal(v2action.Application{
					Name:      "some-app-name",
					GUID:      "some-app-guid",
					SpaceGUID: "some-space-guid",
					Buildpack: "ruby",
				}))
			})
		})

		Context("when the update errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("oh my")
				fakeV2Actor.UpdateApplicationReturns(v2action.Application{}, v2action.Warnings{"update-warning"}, expectedErr)
			})

			It("returns warnings and error and stops", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("update-warning")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				Consistently(eventStream).ShouldNot(Receive(Equal(ApplicationUpdated)))
			})
		})
	})

	Context("when the app does not exist", func() {
		Context("when the creation is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-warning"}, nil)
			})

			It("creates the application", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-warning")))
				Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
				Eventually(eventStream).Should(Receive(Equal(Complete)))

				Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
				Expect(fakeV2Actor.CreateApplicationArgsForCall(0)).To(Equal(v2action.Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}))
			})
		})

		Context("when the creation errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("oh my")
				fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-warning"}, expectedErr)
			})

			It("returns warnings and error and stops", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-warning")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				Consistently(eventStream).ShouldNot(Receive(Equal(ApplicationCreated)))
			})
		})
	})
})
