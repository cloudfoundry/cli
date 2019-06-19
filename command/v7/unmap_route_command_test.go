package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unmap-route Command", func() {
	var (
		cmd             UnmapRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeUnmapRouteActor
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
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeUnmapRouteActor)

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

		cmd = UnmapRouteCommand{
			RequiredArgs: flag.AppDomain{App: appName, Domain: domain},
			Hostname:     hostname,
			Path:         path,
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: orgName,
			GUID: orgGUID,
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)
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
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the user is logged in and targeted", func() {
		When("getting the domain errors", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(v7action.Domain{}, v7action.Warnings{"get-domain-warnings"}, errors.New("get-domain-error"))
			})

			It("returns the error and displays warnings", func() {
				Expect(testUI.Err).To(Say("get-domain-warnings"))
				Expect(executeErr).To(MatchError(errors.New("get-domain-error")))

				Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
				Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

				Expect(fakeActor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(0))
				Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))
				Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
			})
		})

		When("getting the domain succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(
					v7action.Domain{Name: "some-domain.com", GUID: "domain-guid"},
					v7action.Warnings{"get-domain-warnings"},
					nil,
				)
			})

			When("getting the app errors", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationsByNamesAndSpaceReturns(
						[]v7action.Application{},
						v7action.Warnings{"get-app-warnings"},
						errors.New("get-app-error"),
					)
				})

				It("returns the error and displays warnings", func() {
					Expect(testUI.Err).To(Say("get-domain-warnings"))
					Expect(testUI.Err).To(Say("get-app-warnings"))
					Expect(executeErr).To(MatchError(errors.New("get-app-error")))

					Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
					Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

					Expect(fakeActor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(1))
					actualAppNames, actualSpaceGUID := fakeActor.GetApplicationsByNamesAndSpaceArgsForCall(0)
					Expect(len(actualAppNames)).To(Equal(1))
					Expect(actualAppNames[0]).To(Equal(appName))
					Expect(actualSpaceGUID).To(Equal(spaceGUID))

					Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))
					Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
				})
			})

			When("getting the app succeeds", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationsByNamesAndSpaceReturns(
						[]v7action.Application{{Name: "app", GUID: "app-guid"}},
						v7action.Warnings{"get-app-warnings"},
						nil,
					)
				})

				When("getting the route errors", func() {
					BeforeEach(func() {
						fakeActor.GetRouteByAttributesReturns(
							v7action.Route{},
							v7action.Warnings{"get-route-warnings"},
							errors.New("get-route-error"),
						)
					})

					It("returns the error and displays warnings", func() {
						Expect(testUI.Err).To(Say("get-domain-warnings"))
						Expect(testUI.Err).To(Say("get-app-warnings"))
						Expect(testUI.Err).To(Say("get-route-warnings"))
						Expect(executeErr).To(MatchError(errors.New("get-route-error")))

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

						Expect(fakeActor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(1))
						actualAppNames, actualSpaceGUID := fakeActor.GetApplicationsByNamesAndSpaceArgsForCall(0)
						Expect(len(actualAppNames)).To(Equal(1))
						Expect(actualAppNames[0]).To(Equal(appName))
						Expect(actualSpaceGUID).To(Equal(spaceGUID))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomainName, actualDomainGUID, actualHostname, actualPath := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomainName).To(Equal("some-domain.com"))
						Expect(actualDomainGUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))

						Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
					})
				})

				When("getting the route succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetRouteByAttributesReturns(
							v7action.Route{GUID: "route-guid"},
							v7action.Warnings{"get-route-warnings"},
							nil,
						)
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
								v7action.RouteDestination{},
								v7action.Warnings{"get-destination-warning"},
								actionerror.RouteDestinationNotFoundError{},
							)
						})

						It("prints a message and returns without an error", func() {
							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							givenRouteGUID, givenAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(givenRouteGUID).To(Equal("route-guid"))
							Expect(givenAppGUID).To(Equal("app-guid"))

							Expect(executeErr).NotTo(HaveOccurred())
							Expect(testUI.Err).To(Say("get-destination-warning"))
							Expect(testUI.Out).To(Say("Route to be unmapped is not currently mapped to the application."))
							Expect(testUI.Out).To(Say("OK"))
						})
					})

					When("getting the route destination fails for another reason", func() {
						BeforeEach(func() {
							fakeActor.GetRouteDestinationByAppGUIDReturns(
								v7action.RouteDestination{},
								v7action.Warnings{"get-destination-warning"},
								errors.New("failed to get destination"),
							)
						})

						It("prints warnings and returns the error", func() {
							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							givenRouteGUID, givenAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(givenRouteGUID).To(Equal("route-guid"))
							Expect(givenAppGUID).To(Equal("app-guid"))

							Expect(testUI.Err).To(Say("get-destination-warning"))
							Expect(executeErr).To(MatchError("failed to get destination"))

							Expect(fakeActor.UnmapRouteCallCount()).To(Equal(0))
						})
					})

					When("getting the route destination succeeds", func() {
						BeforeEach(func() {
							fakeActor.GetRouteDestinationByAppGUIDReturns(
								v7action.RouteDestination{GUID: "destination-guid"},
								v7action.Warnings{"get-destination-warning"},
								nil,
							)
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

								Expect(testUI.Err).To(Say("get-destination-warning"))

								Expect(executeErr).To(MatchError("failed to unmap route"))
							})
						})

						When("unmapping the route succeeds", func() {
							BeforeEach(func() {
								fakeActor.UnmapRouteReturns(
									v7action.Warnings{"unmap-route-warnings"},
									nil,
								)
							})

							It("prints warnings and does not return an error", func() {
								Expect(fakeActor.UnmapRouteCallCount()).To(Equal(1))
								givenRouteGUID, givenDestinationGUID := fakeActor.UnmapRouteArgsForCall(0)
								Expect(givenRouteGUID).To(Equal("route-guid"))
								Expect(givenDestinationGUID).To(Equal("destination-guid"))

								Expect(testUI.Err).To(Say("get-destination-warning"))
								Expect(testUI.Out).To(Say("OK"))

								Expect(executeErr).NotTo(HaveOccurred())
							})
						})
					})
				})
			})
		})
	})
})
