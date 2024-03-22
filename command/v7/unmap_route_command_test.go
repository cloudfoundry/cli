package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unmap-route Command", func() {
	var (
		cmd             UnmapRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
		domain          string
		appName         string
		hostname        string
		path            string
		orgGUID         string
		orgName         string
		spaceGUID       string
		spaceName       string
		userName        string
		port            int
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "my-app"
		domain = "some-domain.com"
		hostname = "host"
		path = "/path"
		orgGUID = "some-org-guid"
		orgName = "some-org"
		spaceGUID = "some-space-guid"
		spaceName = "some-space"
		userName = "steve"
		port = 0

		cmd = UnmapRouteCommand{
			RequiredArgs: flag.AppDomain{App: appName, Domain: domain},
			Hostname:     hostname,
			Path:         flag.V7RoutePath{Path: path},
			Port:         port,
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: orgName,
			GUID: orgGUID,
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: userName}, nil)

		fakeActor.GetDomainByNameReturns(
			resources.Domain{Name: "some-domain.com", GUID: "domain-guid"},
			v7action.Warnings{"get-domain-warnings"},
			nil,
		)

		fakeActor.GetApplicationByNameAndSpaceReturns(
			resources.Application{Name: "app", GUID: "app-guid"},
			v7action.Warnings{"get-app-warnings"},
			nil,
		)

		fakeActor.GetRouteByAttributesReturns(
			resources.Route{GUID: "route-guid"},
			v7action.Warnings{"get-route-warnings"},
			nil,
		)

		fakeActor.GetRouteDestinationByAppGUIDReturns(
			resources.RouteDestination{GUID: "destination-guid"},
			nil,
		)

		fakeActor.UnmapRouteReturns(
			v7action.Warnings{"unmap-route-warnings"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("getting the domain errors", func() {
		BeforeEach(func() {
			fakeActor.GetDomainByNameReturns(resources.Domain{}, v7action.Warnings{"get-domain-warnings"}, errors.New("get-domain-error"))
		})

		It("returns the error and displays warnings", func() {
			Expect(testUI.Err).To(Say("get-domain-warnings"))
			Expect(executeErr).To(MatchError(errors.New("get-domain-error")))
			Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
		})
	})

	It("gets the domain and displays warnings", func() {
		Expect(testUI.Err).To(Say("get-domain-warnings"))

		Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
		Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))
	})

	When("getting the app errors", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturns(
				resources.Application{},
				v7action.Warnings{"get-app-warnings"},
				errors.New("get-app-error"),
			)
		})

		It("returns the error and displays warnings", func() {
			Expect(testUI.Err).To(Say("get-app-warnings"))
			Expect(executeErr).To(MatchError(errors.New("get-app-error")))
			Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
		})
	})

	It("gets the app and displays the warnings", func() {
		Expect(testUI.Err).To(Say("get-app-warnings"))

		Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
		actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
		Expect(actualAppName).To(Equal(appName))
		Expect(actualSpaceGUID).To(Equal(spaceGUID))
	})

	When("getting the route errors", func() {
		BeforeEach(func() {
			fakeActor.GetRouteByAttributesReturns(
				resources.Route{},
				v7action.Warnings{"get-route-warnings"},
				errors.New("get-route-error"),
			)
		})

		It("returns the error and displays warnings", func() {
			Expect(testUI.Err).To(Say("get-route-warnings"))
			Expect(executeErr).To(MatchError(errors.New("get-route-error")))

			Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
		})
	})

	When("the route is TCP", func() {
		BeforeEach(func() {
			cmd.Hostname = ""
			cmd.Path = flag.V7RoutePath{Path: ""}
			cmd.Port = 1024
		})

		It("gets the routes and displays warnings", func() {
			Expect(testUI.Err).To(Say("get-route-warnings"))

			Expect(testUI.Out).To(Say(
				`Removing route %s from app %s in org %s / space %s as %s\.\.\.`,
				"some-domain.com:1024",
				appName,
				orgName,
				spaceName,
				userName,
			))

			Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
			actualDomain, _, _, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
			Expect(actualDomain.Name).To(Equal("some-domain.com"))
			Expect(actualDomain.GUID).To(Equal("domain-guid"))
			Expect(actualPort).To(Equal(1024))
		})
	})

	It("gets the routes and displays warnings", func() {
		Expect(testUI.Err).To(Say("get-route-warnings"))

		Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
		actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
		Expect(actualDomain.Name).To(Equal("some-domain.com"))
		Expect(actualDomain.GUID).To(Equal("domain-guid"))
		Expect(actualHostname).To(Equal(hostname))
		Expect(actualPath).To(Equal(path))
		Expect(actualPort).To(Equal(0))
	})

	It("prints flavor text", func() {
		Expect(testUI.Out).To(Say(
			`Removing route %s from app %s in org %s / space %s as %s\.\.\.`,
			"host.some-domain.com/path",
			appName,
			orgName,
			spaceName,
			userName,
		))
	})

	When("getting the route destination fails because the app is not mapped", func() {
		BeforeEach(func() {
			fakeActor.GetRouteDestinationByAppGUIDReturns(
				resources.RouteDestination{},
				actionerror.RouteDestinationNotFoundError{},
			)
		})

		It("prints a message and returns without an error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Route to be unmapped is not currently mapped to the application."))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	It("gets the route destination and prints the warnings", func() {
		Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
		givenRoute, givenAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
		Expect(givenRoute.GUID).To(Equal("route-guid"))
		Expect(givenAppGUID).To(Equal("app-guid"))
	})

	When("unmapping the route fails", func() {
		BeforeEach(func() {
			fakeActor.UnmapRouteReturns(
				v7action.Warnings{"unmap-route-warnings"},
				errors.New("failed to unmap route"),
			)
		})

		It("prints warnings and returns the error", func() {
			Expect(fakeActor.UnmapRouteCallCount()).To(Equal(1))
			givenRouteGUID, givenDestinationGUID := fakeActor.UnmapRouteArgsForCall(0)
			Expect(givenRouteGUID).To(Equal("route-guid"))
			Expect(givenDestinationGUID).To(Equal("destination-guid"))

			Expect(executeErr).To(MatchError("failed to unmap route"))
		})
	})

	It("prints warnings and does not return an error", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say("OK"))

		Expect(fakeActor.UnmapRouteCallCount()).To(Equal(1))
		givenRouteGUID, givenDestinationGUID := fakeActor.UnmapRouteArgsForCall(0)
		Expect(givenRouteGUID).To(Equal("route-guid"))
		Expect(givenDestinationGUID).To(Equal("destination-guid"))
	})
})
