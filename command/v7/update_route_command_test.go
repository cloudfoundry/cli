package v7_test

import (
	"errors"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/flag"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("update-route Command", func() {
	var (
		cmd             UpdateRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
		domain          string
		hostname        string
		path            string
		orgGUID         string
		spaceGUID       string
		commandOptions  []string
		removeOptions   []string
		options         map[string]*string
		cCAPIOldVersion string
		routeGuid       string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig.APIVersionReturns(ccversion.MinVersionPerRouteOpts)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		domain = "some-domain.com"
		hostname = "host"
		path = `path`
		orgGUID = "some-org-guid"
		spaceGUID = "some-space-guid"
		commandOptions = []string{"loadbalancing=least-connection"}
		removeOptions = []string{"loadbalancing"}
		lbLCVal := "least-connection"
		lbLeastConnections := &lbLCVal
		options = map[string]*string{"loadbalancing": lbLeastConnections}
		routeGuid = "route-guid"

		cmd = UpdateRouteCommand{
			RequiredArgs:  flag.Domain{Domain: domain},
			Hostname:      hostname,
			Path:          flag.V7RoutePath{Path: path},
			Options:       commandOptions,
			RemoveOptions: removeOptions,
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: orgGUID,
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: spaceGUID,
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)

		fakeActor.GetRouteByAttributesReturns(
			resources.Route{GUID: routeGuid, URL: domain},
			v7action.Warnings{"get-route-warnings"},
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

	When("the user is logged in and targeted", func() {
		When("getting the domain errors", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(resources.Domain{}, v7action.Warnings{"get-domain-warnings"}, errors.New("get-domain-error"))
			})

			It("returns the error and displays warnings", func() {
				Expect(testUI.Err).To(Say("get-domain-warnings"))
				Expect(executeErr).To(MatchError(errors.New("get-domain-error")))

				Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
				Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))

				Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))

				Expect(fakeActor.CreateRouteCallCount()).To(Equal(0))

				Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
			})
		})

		When("getting the domain succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(
					resources.Domain{Name: "some-domain.com", GUID: "domain-guid"},
					v7action.Warnings{"get-domain-warnings"},
					nil,
				)
				fakeActor.UpdateRouteReturns(
					resources.Route{GUID: routeGuid, URL: domain, Options: options},
					nil,
					nil,
				)
			})
			When("updating the route fails when the CC API version is too old for route options", func() {
				BeforeEach(func() {
					cmd.Options = []string{}
					cCAPIOldVersion = strconv.Itoa(1)
					fakeConfig.APIVersionReturns(cCAPIOldVersion)
				})

				It("does not update a route giving the error message", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(fakeActor.UpdateRouteCallCount()).To(Equal(0))
					Expect(testUI.Err).To(Say("CC API version"))
					Expect(testUI.Err).To(Say("does not support per-route options"))
				})
			})

			When("the route options are not specified", func() {
				BeforeEach(func() {
					cmd.Options = nil
					cmd.RemoveOptions = nil
				})
				It("does not update a route giving the error message", func() {
					Expect(executeErr).To(MatchError(actionerror.RouteOptionSupportError{
						ErrorText: fmt.Sprintf("No options were specified for the update of the Route %s", domain)}))
					Expect(fakeActor.UpdateRouteCallCount()).To(Equal(0))
				})
			})

			When("the route options are specified incorrectly", func() {
				BeforeEach(func() {
					cmd.Options = []string{"loadbalancing"}
				})
				It("does not update a route giving the error message", func() {
					Expect(executeErr).To(MatchError(actionerror.RouteOptionError{Name: "loadbalancing", DomainName: domain, Path: path, Host: hostname}))
					Expect(fakeActor.UpdateRouteCallCount()).To(Equal(0))
				})
			})

			When("removing the options of the route succeeds", func() {
				BeforeEach(func() {
					cmd.RemoveOptions = []string{"loadbalancing"}
					fakeActor.GetRouteByAttributesReturns(
						resources.Route{GUID: routeGuid, URL: domain, Options: options},
						nil,
						nil,
					)
				})

				It("updates a given route", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					expectedRouteGuid, expectedOptions := fakeActor.UpdateRouteArgsForCall(0)
					Expect(expectedRouteGuid).To(Equal(routeGuid))
					Expect(expectedOptions).To(Equal(options))

					expectedRouteGuid, expectedOptions = fakeActor.UpdateRouteArgsForCall(1)
					Expect(expectedRouteGuid).To(Equal(routeGuid))
					Expect(expectedOptions).To(Equal(map[string]*string{"loadbalancing": nil}))
					Expect(fakeActor.UpdateRouteCallCount()).To(Equal(2))

					Expect(testUI.Out).To(Say("Updating route"))
					Expect(testUI.Out).To(Say("has been updated"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("a requested route exists", func() {
				BeforeEach(func() {
					fakeActor.GetRouteByAttributesReturns(
						resources.Route{GUID: "route-guid", URL: domain},
						nil,
						nil,
					)
				})

				It("calls update route passing the proper arguments", func() {
					By("passing the expected arguments to the actor ", func() {
						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal("some-domain.com"))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal("host"))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))

						Expect(fakeActor.UpdateRouteCallCount()).To(Equal(2))
						actualRouteGUID, actualOptions := fakeActor.UpdateRouteArgsForCall(0)
						Expect(actualRouteGUID).To(Equal("route-guid"))
						Expect(actualOptions).To(Equal(options))

						// Second update route call to remove the option
						actualRouteGUID, actualOptions = fakeActor.UpdateRouteArgsForCall(1)
						Expect(actualRouteGUID).To(Equal("route-guid"))
						options["loadbalancing"] = nil
						Expect(actualOptions).To(Equal(options))
					})
				})
			})
		})
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
