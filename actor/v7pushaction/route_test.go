package v7pushaction_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var (
		actor                   *Actor
		fakeV7Actor             *v7pushactionfakes.FakeV7Actor
		fakeRandomWordGenerator *v7pushactionfakes.FakeRandomWordGenerator
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		fakeRandomWordGenerator = new(v7pushactionfakes.FakeRandomWordGenerator)
		actor.RandomWordGenerator = fakeRandomWordGenerator
	})

	Describe("CreateAndMapRoute", func() {
		var (
			warnings   Warnings
			executeErr error

			orgGUID   string
			spaceGUID string
			app       v7action.Application
			gt        GenesisTechnique
		)

		BeforeEach(func() {
			orgGUID = "org-guid"
			spaceGUID = "space-guid"
			app = v7action.Application{Name: "app-name", GUID: "app-guid"}
			gt = DefaultRoute
			fakeV7Actor.GetDefaultDomainReturns(
				v7action.Domain{GUID: "domain-guid", Name: "domain-name"},
				v7action.Warnings{"get-default-domain-warning"},
				nil,
			)
			fakeV7Actor.GetRouteByAttributesReturns(
				v7action.Route{},
				v7action.Warnings{"get-route-by-attribute-warning"},
				actionerror.RouteNotFoundError{},
			)
			fakeV7Actor.CreateRouteStub = func(spaceGUID, domainName, host, path string) (v7action.Route, v7action.Warnings, error) {
				return v7action.Route{GUID: "route-guid", SpaceGUID: spaceGUID, DomainGUID: "domain-guid", DomainName: domainName, Host: host, Path: path},
					v7action.Warnings{"create-route-warning"},
					nil
			}
			fakeV7Actor.MapRouteReturns(
				v7action.Warnings{"map-route-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateAndMapRoute(orgGUID, spaceGUID, app, gt)
		})

		When("the route does **not** exist", func() {
			It("returns no error and all the warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-default-domain-warning", "get-route-by-attribute-warning", "create-route-warning", "map-route-warning"))
			})

			It("creates a route within the default domain", func() {
				Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
				actualSpaceGUID, actualDomainName, actualHost, actualPath := fakeV7Actor.CreateRouteArgsForCall(0)
				Expect(actualHost).To(Equal(app.Name))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualDomainName).To(Equal("domain-name"))
				Expect(actualPath).To(Equal(""))
			})

			When("creating the route fails", func() {
				BeforeEach(func() {
					fakeV7Actor.CreateRouteReturns(v7action.Route{}, v7action.Warnings{"create-route-warning"}, errors.New("create-route-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("create-route-error"))
					Expect(warnings).To(ConsistOf("get-default-domain-warning", "get-route-by-attribute-warning", "create-route-warning"))
					Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(0))
				})
			})

			It("maps the created route to the app", func() {
				Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
				actualRouteGUID, actualAppGUID := fakeV7Actor.MapRouteArgsForCall(0)
				Expect(actualRouteGUID).To(Equal("route-guid"))
				Expect(actualAppGUID).To(Equal(app.GUID))
			})
		})

		When("the route exists", func() {
			BeforeEach(func() {
				fakeV7Actor.GetRouteByAttributesReturns(v7action.Route{GUID: "route-guid"}, v7action.Warnings{"get-route-by-attribute-warning"}, nil)
			})

			It("returns no error and all the warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-default-domain-warning", "get-route-by-attribute-warning", "map-route-warning"))
			})

			It("does **not** create the route", func() {
				Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(0))
			})

			It("maps the route to the app", func() {
				Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
				actualRouteGUID, actualAppGUID := fakeV7Actor.MapRouteArgsForCall(0)
				Expect(actualRouteGUID).To(Equal("route-guid"))
				Expect(actualAppGUID).To(Equal(app.GUID))
			})
		})

		When("getting the default domain fails", func() {
			BeforeEach(func() {
				fakeV7Actor.GetDefaultDomainReturns(
					v7action.Domain{},
					v7action.Warnings{"get-default-domain-warning"},
					errors.New("get-default-domain-error"),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-default-domain-error"))
				Expect(warnings).To(ConsistOf("get-default-domain-warning"))
				Expect(fakeV7Actor.GetRouteByAttributesCallCount()).To(Equal(0))
				Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(0))
				Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(0))
			})
		})

		When("mapping the route fails", func() {
			BeforeEach(func() {
				fakeV7Actor.MapRouteReturns(
					v7action.Warnings{"map-route-warning"},
					errors.New("map-route-error"),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("map-route-error"))
				Expect(warnings).To(ConsistOf("get-default-domain-warning", "get-route-by-attribute-warning", "create-route-warning", "map-route-warning"))
			})
		})

		When("random is true", func() {
			BeforeEach(func() {
				gt = RandomRoute
				fakeRandomWordGenerator.RandomAdjectiveReturns("awesome")
				fakeRandomWordGenerator.RandomNounReturns("sauce")
			})

			It("creates and maps a route with a random host", func() {
				Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
				actualSpaceGUID, actualDomainName, actualHost, actualPath := fakeV7Actor.CreateRouteArgsForCall(0)
				Expect(actualHost).To(Equal(strings.Join([]string{app.Name, "awesome", "sauce"}, "-")))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualDomainName).To(Equal("domain-name"))
				Expect(actualPath).To(Equal(""))
			})
		})

		When("random is false", func() {
			BeforeEach(func() {
				gt = DefaultRoute
			})

			It("creates and maps a route with the app name as the host", func() {
				Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
				actualSpaceGUID, actualDomainName, actualHost, actualPath := fakeV7Actor.CreateRouteArgsForCall(0)
				Expect(actualHost).To(Equal(app.Name))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualDomainName).To(Equal("domain-name"))
				Expect(actualPath).To(Equal(""))
			})
		})
	})
})
