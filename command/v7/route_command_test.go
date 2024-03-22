package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("route Command", func() {
	var (
		cmd             v7.RouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		domainName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		domainName = "some-domain.com"

		cmd = v7.RouteCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Domain{Domain: domainName},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)

		fakeActor.GetDomainByNameReturns(
			resources.Domain{Name: domainName, GUID: "domain-guid"},
			v7action.Warnings{"get-domain-warnings"},
			nil,
		)

		fakeActor.GetRouteByAttributesReturns(
			resources.Route{GUID: "route-guid"},
			v7action.Warnings{"get-route-warnings"},
			nil,
		)
		fakeActor.GetApplicationMapForRouteReturns(
			map[string]resources.Application{"app-guid": {GUID: "app-guid", Name: "app-name"}},
			v7action.Warnings{"get-route-warnings"},
			nil,
		)

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the target", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(checkTargetedOrg).To(BeTrue())
		Expect(checkTargetedSpace).To(BeFalse())
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
		})
	})

	It("checks if the user is logged in", func() {
		Expect(fakeActor.GetCurrentUserCallCount()).To(Equal(1))
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("no current user"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("no current user"))
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
		Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))
	})

	Describe("getting the routes", func() {
		It("calls GetRouteByAttributes and displaying warnings", func() {
			Expect(testUI.Err).To(Say("route-warning"))

			Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
			domain, host, path, port := fakeActor.GetRouteByAttributesArgsForCall(0)
			Expect(domain.Name).To(Equal(domainName))
			Expect(domain.GUID).To(Equal("domain-guid"))
			Expect(host).To(Equal(cmd.Hostname))
			Expect(path).To(Equal(cmd.Path.Path))
			Expect(port).To(Equal(0))
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
			})
		})
	})

	Describe("getting the apps", func() {
		It("calls GetApplicationMapForRoute and displaying warnings", func() {
			Expect(testUI.Err).To(Say("route-warning"))

			Expect(fakeActor.GetApplicationMapForRouteCallCount()).To(Equal(1))
			route := fakeActor.GetApplicationMapForRouteArgsForCall(0)
			Expect(route.GUID).To(Equal("route-guid"))
		})

		When("getting the Application mapping errors", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationMapForRouteReturns(
					map[string]resources.Application{},
					v7action.Warnings{"get-app-map-warnings"},
					errors.New("get-app-map-error"),
				)
			})

			It("returns the error and displays warnings", func() {
				Expect(testUI.Err).To(Say("get-app-map-warnings"))
				Expect(executeErr).To(MatchError(errors.New("get-app-map-error")))
			})
		})
	})

	When("passing hostname and path flags", func() {
		BeforeEach(func() {
			cmd.Path.Path = "/some-path"
			cmd.Hostname = "some-host"

			destAppA := resources.RouteDestinationApp{GUID: "abc", Process: struct{ Type string }{"web"}}
			destinationA := resources.RouteDestination{App: destAppA, Port: 8080, Protocol: "http1"}

			destAppB := resources.RouteDestinationApp{GUID: "123", Process: struct{ Type string }{"web"}}
			destinationB := resources.RouteDestination{App: destAppB, Port: 1337, Protocol: "http2"}

			destinations := []resources.RouteDestination{destinationA, destinationB}
			route := resources.Route{GUID: "route-guid", Host: cmd.Hostname, Path: cmd.Path.Path, Protocol: "http", Destinations: destinations}

			fakeActor.GetRouteByAttributesReturns(
				route,
				v7action.Warnings{"get-route-warnings"},
				nil,
			)

			appA := resources.Application{GUID: "abc", Name: "app-name"}
			appB := resources.Application{GUID: "123", Name: "other-app-name"}

			fakeActor.GetApplicationMapForRouteReturns(
				map[string]resources.Application{"abc": appA, "123": appB},
				v7action.Warnings{"get-apps-error"},
				nil,
			)
		})

		It("displays the summary", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say(`Showing route %s\.%s/some-path in org some-org / space some-space as some-user\.\.\.`, cmd.Hostname, domainName))
			Expect(testUI.Out).To(Say(`domain:\s+%s`, domainName))
			Expect(testUI.Out).To(Say(`host:\s+%s`, cmd.Hostname))
			Expect(testUI.Out).To(Say(`path:\s+%s`, cmd.Path.Path))
			Expect(testUI.Out).To(Say(`protocol:\s+http`))
			Expect(testUI.Out).To(Say(`\n`))
			Expect(testUI.Out).To(Say(`Destinations:`))
			Expect(testUI.Out).To(Say(`\s+app\s+process\s+port\s+app-protocol`))
			Expect(testUI.Out).To(Say(`\s+app-name\s+web\s+8080\s+http1`))
			Expect(testUI.Out).To(Say(`\s+other-app-name\s+web\s+1337\s+http2`))

			Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
			givenDomain, givenHostname, givenPath, givenPort := fakeActor.GetRouteByAttributesArgsForCall(0)
			Expect(givenDomain.Name).To(Equal(domainName))
			Expect(givenHostname).To(Equal("some-host"))
			Expect(givenPath).To(Equal("/some-path"))
			Expect(givenPort).To(Equal(0))
		})
	})
	Describe("RouteRetrieval display logic", func() {
		When("passing in just a domain", func() {
			BeforeEach(func() {
				cmd.Hostname = ""
				cmd.Path.Path = ""
			})
			It(" displays the right stuff", func() {
				Expect(testUI.Out).To(Say(`Showing route %s in org some-org / space some-space as some-user\.\.\.`, domainName))
			})
		})
		When("passing in a domain and hostname", func() {
			BeforeEach(func() {
				cmd.Hostname = "some-host"
				cmd.Path.Path = ""
			})
			It(" displays the right stuff", func() {
				Expect(testUI.Out).To(Say(`Showing route some-host\.%s in org some-org / space some-space as some-user\.\.\.`, domainName))
			})
		})

		When("passing in a domain, a hostname, and a path", func() {
			BeforeEach(func() {
				cmd.Hostname = "some-host"
				cmd.Path.Path = "/some-path"
			})
			It(" displays the right stuff", func() {
				Expect(testUI.Out).To(Say(`Showing route some-host\.%s\/some-path in org some-org / space some-space as some-user\.\.\.`, domainName))
			})
		})
		When("passing in a domain and a port", func() {
			BeforeEach(func() {
				cmd.Hostname = ""
				cmd.Path.Path = ""
				cmd.Port = 8080
			})
			It(" displays the right stuff", func() {
				Expect(testUI.Out).To(Say(`Showing route %s:8080 in org some-org / space some-space as some-user\.\.\.`, domainName))
			})
		})
	})
})
