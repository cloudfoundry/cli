package pushaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Config", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)
	})

	Describe("ConvertToApplicationConfigs", func() {
		var (
			appName      string
			orgGUID      string
			spaceGUID    string
			domain       v2action.Domain
			manifestApps []manifest.Application

			configs    []ApplicationConfig
			warnings   Warnings
			executeErr error

			firstConfig ApplicationConfig
		)

		BeforeEach(func() {
			appName = "some-app"
			orgGUID = "some-org-guid"
			spaceGUID = "some-space-guid"
			manifestApps = []manifest.Application{{
				Name: appName,
				Path: "some-path",
			}}

			domain = v2action.Domain{
				Name: "private-domain.com",
				GUID: "some-private-domain-guid",
			}
			// Prevents NoDomainsFoundError
			fakeV2Actor.GetOrganizationDomainsReturns(
				[]v2action.Domain{domain},
				v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"},
				nil,
			)
		})

		JustBeforeEach(func() {
			configs, warnings, executeErr = actor.ConvertToApplicationConfigs(orgGUID, spaceGUID, manifestApps)
			if len(configs) > 0 {
				firstConfig = configs[0]
			}
		})

		Context("when the application exists", func() {
			var app v2action.Application
			var route v2action.Route

			BeforeEach(func() {
				app = v2action.Application{
					Name:      appName,
					GUID:      "some-app-guid",
					SpaceGUID: spaceGUID,
				}

				route = v2action.Route{
					Domain: v2action.Domain{
						Name: "some-domain.com",
						GUID: "some-domain-guid",
					},
					Host:      app.Name,
					GUID:      "route-guid",
					SpaceGUID: spaceGUID,
				}

				fakeV2Actor.GetApplicationByNameAndSpaceReturns(app, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, nil)
			})

			Context("when retrieving the application's routes is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{route}, v2action.Warnings{"app-route-warnings"}, nil)
				})

				It("sets the current and desired application to the current", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings", "private-domain-warnings", "shared-domain-warnings"))
					Expect(firstConfig.CurrentApplication).To(Equal(app))
					Expect(firstConfig.DesiredApplication).To(Equal(app))
					Expect(firstConfig.Path).To(Equal("some-path"))
					Expect(firstConfig.TargetedSpaceGUID).To(Equal(spaceGUID))

					Expect(fakeV2Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, passedSpaceGUID := fakeV2Actor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal(app.Name))
					Expect(passedSpaceGUID).To(Equal(spaceGUID))
				})

				It("sets the current routes", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings", "private-domain-warnings", "shared-domain-warnings"))
					Expect(firstConfig.CurrentRoutes).To(ConsistOf(route))
				})
			})

			Context("when retrieving the application's routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"app-route-warnings"}, expectedErr)
				})

				It("sets the current and desired application to the current", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings"))

					Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GetApplicationRoutesArgsForCall(0)).To(Equal(app.GUID))
				})
			})
		})

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, v2action.ApplicationNotFoundError{})
			})

			It("creates a new application and sets it to the desired application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "private-domain-warnings", "shared-domain-warnings"))
				Expect(firstConfig.CurrentApplication).To(Equal(v2action.Application{}))
				Expect(firstConfig.DesiredApplication).To(Equal(v2action.Application{
					Name:      "some-app",
					SpaceGUID: spaceGUID,
				}))
				Expect(firstConfig.Path).To(Equal("some-path"))
				Expect(firstConfig.TargetedSpaceGUID).To(Equal(spaceGUID))
			})
		})

		Context("when retrieving the application errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("dios mio")
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
			})
		})

		Context("when retrieving the default route is successful", func() {
			BeforeEach(func() {
				// Assumes new route
				fakeV2Actor.CheckRouteReturns(false, v2action.Warnings{"get-route-warnings"}, nil)
			})

			It("adds the route to desired routes", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))
				Expect(firstConfig.DesiredRoutes).To(ConsistOf(v2action.Route{
					Domain:    domain,
					Host:      appName,
					SpaceGUID: spaceGUID,
				}))
			})
		})

		Context("when retrieving the default route errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("dios mio")
				fakeV2Actor.CheckRouteReturns(false, v2action.Warnings{"get-route-warnings"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))
			})
		})

		Context("when given a directory", func() {
			Context("when scanning is successful", func() {
				var resources []v2action.Resource

				BeforeEach(func() {
					resources = []v2action.Resource{
						{Filename: "I am a file!"},
						{Filename: "I am not a file"},
					}
					fakeV2Actor.GatherResourcesReturns(resources, nil)
				})

				It("sets the full resource list on the config", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
					Expect(firstConfig.AllResources).To(Equal(resources))

					Expect(fakeV2Actor.GatherResourcesCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GatherResourcesArgsForCall(0)).To(Equal("some-path"))
				})
			})

			Context("when scanning errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					fakeV2Actor.GatherResourcesReturns(nil, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
				})
			})
		})
	})
})
