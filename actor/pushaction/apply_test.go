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

	Describe("when routes need to be created", func() {
		BeforeEach(func() {
			// This will skip the binding step
			config.CurrentRoutes = []v2action.Route{
				{GUID: ""},
				{GUID: "some-route-guid-2"},
			}

			config.DesiredRoutes = []v2action.Route{
				{GUID: "", Host: "some-route-1"},
				{GUID: "some-route-guid-2", Host: "some-route-2"},
				{GUID: "", Host: "some-route-3"},
			}

			fakeV2Actor.CreateApplicationReturns(
				v2action.Application{
					GUID: "some-app-guid",
				},
				v2action.Warnings{"create-app-warning"},
				nil)
		})

		Context("when the creation is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.CreateRouteReturns(
					v2action.Route{},
					v2action.Warnings{"create-route-warning"},
					nil)
			})

			It("only creates the routes that do not exist", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-app-warning")))
				Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warning")))

				Eventually(eventStream).Should(Receive(Equal(RouteCreated)))
				Eventually(eventStream).Should(Receive(Equal(Complete)))

				Expect(fakeV2Actor.CreateRouteCallCount()).To(Equal(2))
				Expect(fakeV2Actor.CreateRouteArgsForCall(0)).To(Equal(v2action.Route{Host: "some-route-1"}))
				Expect(fakeV2Actor.CreateRouteArgsForCall(1)).To(Equal(v2action.Route{Host: "some-route-3"}))
			})
		})

		Context("when the creation errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("oh my")
				fakeV2Actor.CreateRouteReturns(
					v2action.Route{},
					v2action.Warnings{"create-route-warning"},
					expectedErr)
			})

			It("returns warnings and error and stops", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-app-warning")))
				Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warning")))

				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				Consistently(eventStream).ShouldNot(Receive(Equal(RouteCreated)))
			})
		})
	})

	Context("when no routes are created", func() {
		BeforeEach(func() {
			fakeV2Actor.CreateRouteReturns(
				v2action.Route{},
				v2action.Warnings{"create-route-warning"},
				nil)
		})

		It("returns warnings and error and stops", func() {
			Eventually(warningsStream).Should(Receive())
			Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
			Consistently(eventStream).ShouldNot(Receive(Equal(RouteCreated)))
		})
	})

	Context("when routes need to be bound to the application", func() {
		BeforeEach(func() {
			config.CurrentRoutes = []v2action.Route{
				{GUID: "some-route-guid-2", Host: "some-route-2"},
			}
			config.DesiredRoutes = []v2action.Route{
				{GUID: "some-route-guid-1", Host: "some-route-1", Domain: v2action.Domain{Name: "some-domain.com"}},
				{GUID: "some-route-guid-2", Host: "some-route-2"},
				{GUID: "some-route-guid-3", Host: "some-route-3"},
			}

			fakeV2Actor.CreateApplicationReturns(
				v2action.Application{
					GUID: "some-app-guid",
				},
				v2action.Warnings{"create-app-warning"},
				nil)
		})

		Context("when the binding is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, nil)
			})

			It("only creates the routes that do not exist", func() {
				Eventually(warningsStream).Should(Receive())
				Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
				Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))

				Eventually(eventStream).Should(Receive(Equal(RouteBound)))
				Eventually(eventStream).Should(Receive(Equal(Complete)))

				Expect(fakeV2Actor.BindRouteToApplicationCallCount()).To(Equal(2))

				routeGUID, appGUID := fakeV2Actor.BindRouteToApplicationArgsForCall(0)
				Expect(routeGUID).To(Equal("some-route-guid-1"))
				Expect(appGUID).To(Equal("some-app-guid"))

				routeGUID, appGUID = fakeV2Actor.BindRouteToApplicationArgsForCall(1)
				Expect(routeGUID).To(Equal("some-route-guid-3"))
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when the creation errors", func() {
			Context("when the route is bound in another space", func() {
				BeforeEach(func() {
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, v2action.RouteInDifferentSpaceError{})
				})

				It("stops and returns the RouteInDifferentSpaceError (with a guid set) and warnings", func() {
					Eventually(warningsStream).Should(Receive())
					Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))

					Eventually(errorStream).Should(Receive(MatchError(
						v2action.RouteInDifferentSpaceError{Route: "some-route-1.some-domain.com"},
					)))
					Consistently(eventStream).ShouldNot(Receive(Equal(RouteBound)))
				})
			})
			Context("generic error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, expectedErr)
				})

				It("returns warnings and error and stops", func() {
					Eventually(warningsStream).Should(Receive())
					Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))

					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
					Consistently(eventStream).ShouldNot(Receive(Equal(RouteBound)))
				})
			})
		})
	})

	Context("when no routes need to be bound", func() {
		It("returns warnings and error and stops", func() {
			Eventually(warningsStream).Should(Receive())
			Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
			Consistently(eventStream).ShouldNot(Receive(Equal(RouteBound)))
		})
	})
})
