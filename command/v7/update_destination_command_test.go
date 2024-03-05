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

var _ = Describe("update Command", func() {
	var (
		cmd             v7.UpdateDestinationCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		domainName      string
		appName         string
		appProtocol     string
		orgGUID         string
		spaceGUID       string
		hostname        string
		path            string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		domainName = "some-domain.com"
		appName = "super-app"
		appProtocol = "http2"
		orgGUID = "some-org-guid"
		spaceGUID = "some-space-guid"
		hostname = "hostname"
		path = "path"

		cmd = v7.UpdateDestinationCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.AppDomain{App: appName, Domain: domainName},
			AppProtocol:  appProtocol,
			Hostname:     hostname,
			Path:         flag.V7RoutePath{Path: path},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: spaceGUID})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org", GUID: orgGUID})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the target", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(checkTargetedOrg).To(BeTrue())
		Expect(checkTargetedSpace).To(BeTrue())
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
				Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))

				Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))

				Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(0))

				Expect(fakeActor.UpdateDestinationCallCount()).To(Equal(0))
			})
		})

		When("getting the domain succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(
					resources.Domain{Name: domainName, GUID: "domain-guid"},
					v7action.Warnings{"get-domain-warnings"},
					nil,
				)
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
					Expect(testUI.Err).To(Say("get-domain-warnings"))
					Expect(testUI.Err).To(Say("get-app-warnings"))
					Expect(executeErr).To(MatchError(errors.New("get-app-error")))

					Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
					Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(actualAppName).To(Equal(appName))
					Expect(actualSpaceGUID).To(Equal(spaceGUID))

					Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))

					Expect(fakeActor.CreateRouteCallCount()).To(Equal(0))

					Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
				})
			})

			When("getting the app succeeds", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{GUID: "app-guid", Name: appName},
						v7action.Warnings{"get-app-warnings"},
						nil,
					)
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
						Expect(testUI.Err).To(Say("get-domain-warnings"))
						Expect(testUI.Err).To(Say("get-app-warnings"))
						Expect(testUI.Err).To(Say("get-route-warnings"))
						Expect(executeErr).To(MatchError(errors.New("get-route-error")))

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(actualAppName).To(Equal(appName))
						Expect(actualSpaceGUID).To(Equal(spaceGUID))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal(domainName))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))

						Expect(fakeActor.UpdateDestinationCallCount()).To(Equal(0))

					})
				})

				When("the requested route does not exist", func() {
					BeforeEach(func() {
						fakeActor.GetRouteByAttributesReturns(
							resources.Route{},
							v7action.Warnings{"get-route-warnings"},
							actionerror.RouteNotFoundError{},
						)
					})

					It("displays error message", func() {
						Expect(testUI.Err).To(Say("get-domain-warnings"))
						Expect(testUI.Err).To(Say("get-route-warnings"))
						Expect(executeErr).To(HaveOccurred())

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(actualAppName).To(Equal(appName))
						Expect(actualSpaceGUID).To(Equal(spaceGUID))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal(domainName))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))
					})
				})

				When("the requested route exists", func() {
					BeforeEach(func() {
						fakeActor.GetRouteByAttributesReturns(
							resources.Route{GUID: "route-guid"},
							v7action.Warnings{"get-route-warnings"},
							nil,
						)
					})
					When("getting the destination errors", func() {
						BeforeEach(func() {
							fakeActor.GetRouteDestinationByAppGUIDReturns(
								resources.RouteDestination{},
								errors.New("get-destination-error"),
							)
						})
						It("returns the error and warnings", func() {
							Expect(executeErr).To(MatchError(errors.New("get-destination-error")))

							Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal(domainName))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(0))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))

							Expect(fakeActor.UpdateDestinationCallCount()).To(Equal(0))
						})
					})
					When("the destination already exists", func() {
						BeforeEach(func() {
							fakeActor.GetRouteDestinationByAppGUIDReturns(
								resources.RouteDestination{
									GUID: "route-dst-guid",
									App: resources.RouteDestinationApp{
										GUID: "existing-app-guid",
									},
								},
								nil,
							)
						})

						It("exits 0 with a helpful message that the destination protocol was changed", func() {
							Expect(executeErr).ShouldNot(HaveOccurred())

							Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal(domainName))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(0))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))
							Expect(fakeActor.UpdateDestinationCallCount()).To(Equal(1))
						})
					})

					When("the destination is not found", func() {
						When("updating the destination errors", func() {
							BeforeEach(func() {
								fakeActor.UpdateDestinationReturns(v7action.Warnings{"update-dest-warnings"}, errors.New("map-route-error"))
								fakeActor.GetRouteDestinationByAppGUIDReturns(
									resources.RouteDestination{
										GUID: "route-dst-guid",
										App: resources.RouteDestinationApp{
											GUID: "existing-app-guid",
										},
									},
									nil,
								)
							})

							It("returns the error and displays warnings", func() {
								Expect(testUI.Err).To(Say("get-domain-warnings"))
								Expect(testUI.Err).To(Say("get-app-warnings"))
								Expect(testUI.Err).To(Say("get-route-warnings"))
								Expect(testUI.Err).To(Say("update-dest-warnings"))
								Expect(executeErr).To(MatchError(errors.New("map-route-error")))

								Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
								Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

								Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
								actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
								Expect(actualAppName).To(Equal(appName))
								Expect(actualSpaceGUID).To(Equal(spaceGUID))

								Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
								actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
								Expect(actualDomain.Name).To(Equal(domainName))
								Expect(actualDomain.GUID).To(Equal("domain-guid"))
								Expect(actualHostname).To(Equal(hostname))
								Expect(actualPath).To(Equal(path))
								Expect(actualPort).To(Equal(0))

								Expect(fakeActor.UpdateDestinationCallCount()).To(Equal(1))
								actualRouteGUID, destinationGUID, protocol := fakeActor.UpdateDestinationArgsForCall(0)
								Expect(actualRouteGUID).To(Equal("route-guid"))
								Expect(destinationGUID).To(Equal("route-dst-guid"))
								Expect(protocol).To(Equal("http2"))
							})
						})
					})
				})
			})
		})
	})
})
