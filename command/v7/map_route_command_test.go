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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("map-route Command", func() {
	var (
		cmd             MapRouteCommand
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
		spaceGUID       string
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
		path = `path`
		orgGUID = "some-org-guid"
		spaceGUID = "some-space-guid"

		cmd = MapRouteCommand{
			RequiredArgs: flag.AppDomain{App: appName, Domain: domain},
			Hostname:     hostname,
			Path:         flag.V7RoutePath{Path: path},
			AppProtocol:  "http2",
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
					Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

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
						resources.Application{Name: "app", GUID: "app-guid"},
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
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(actualAppName).To(Equal(appName))
						Expect(actualSpaceGUID).To(Equal(spaceGUID))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal("some-domain.com"))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(cmd.Port))

						Expect(fakeActor.CreateRouteCallCount()).To(Equal(0))

						Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
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

					It("creates the route", func() {
						Expect(testUI.Err).To(Say("get-domain-warnings"))
						Expect(testUI.Err).To(Say("get-app-warnings"))
						Expect(testUI.Err).To(Say("get-route-warnings"))
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(actualAppName).To(Equal(appName))
						Expect(actualSpaceGUID).To(Equal(spaceGUID))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal("some-domain.com"))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(cmd.Port))

						Expect(fakeActor.CreateRouteCallCount()).To(Equal(1))
						actualSpaceGUID, actualDomainName, actualHostname, actualPath, actualPort := fakeActor.CreateRouteArgsForCall(0)
						Expect(actualSpaceGUID).To(Equal(spaceGUID))
						Expect(actualDomainName).To(Equal("some-domain.com"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(cmd.Port))
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
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal("some-domain.com"))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(cmd.Port))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))

							Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
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
						It("exits 0 with a helpful message that the route is already mapped to the app", func() {
							Expect(executeErr).ShouldNot(HaveOccurred())

							Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal("some-domain.com"))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(cmd.Port))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))
							Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
						})

					})
					When("the destination is not found", func() {
						When("mapping the route errors", func() {
							BeforeEach(func() {
								fakeActor.MapRouteReturns(v7action.Warnings{"map-route-warnings"}, errors.New("map-route-error"))
							})

							It("returns the error and displays warnings", func() {
								Expect(testUI.Err).To(Say("get-domain-warnings"))
								Expect(testUI.Err).To(Say("get-app-warnings"))
								Expect(testUI.Err).To(Say("get-route-warnings"))
								Expect(testUI.Err).To(Say("map-route-warnings"))
								Expect(executeErr).To(MatchError(errors.New("map-route-error")))

								Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
								Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

								Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
								actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
								Expect(actualAppName).To(Equal(appName))
								Expect(actualSpaceGUID).To(Equal(spaceGUID))

								Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
								actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
								Expect(actualDomain.Name).To(Equal("some-domain.com"))
								Expect(actualDomain.GUID).To(Equal("domain-guid"))
								Expect(actualHostname).To(Equal(hostname))
								Expect(actualPath).To(Equal(path))
								Expect(actualPort).To(Equal(cmd.Port))

								Expect(fakeActor.MapRouteCallCount()).To(Equal(1))
								actualRouteGUID, actualAppGUID, actualAppProtocol := fakeActor.MapRouteArgsForCall(0)
								Expect(actualRouteGUID).To(Equal("route-guid"))
								Expect(actualAppGUID).To(Equal("app-guid"))
								Expect(actualAppProtocol).To(Equal("http2"))
							})
						})

						When("mapping the route succeeds", func() {
							BeforeEach(func() {
								fakeActor.MapRouteReturns(v7action.Warnings{"map-route-warnings"}, nil)
							})

							It("returns the proper information and passing the proper arguments", func() {
								By("displaying warnings", func() {
									Expect(testUI.Err).To(Say("get-domain-warnings"))
									Expect(testUI.Err).To(Say("get-app-warnings"))
									Expect(testUI.Err).To(Say("get-route-warnings"))
									Expect(testUI.Err).To(Say("map-route-warnings"))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								By("passing the expected arguments to the actor", func() {
									Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
									Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

									Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
									actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
									Expect(actualAppName).To(Equal(appName))
									Expect(actualSpaceGUID).To(Equal(spaceGUID))

									Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
									actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
									Expect(actualDomain.Name).To(Equal("some-domain.com"))
									Expect(actualDomain.GUID).To(Equal("domain-guid"))
									Expect(actualHostname).To(Equal("host"))
									Expect(actualPath).To(Equal(path))
									Expect(actualPort).To(Equal(cmd.Port))

									Expect(fakeActor.MapRouteCallCount()).To(Equal(1))
									actualRouteGUID, actualAppGUID, actualAppProtocol := fakeActor.MapRouteArgsForCall(0)
									Expect(actualRouteGUID).To(Equal("route-guid"))
									Expect(actualAppGUID).To(Equal("app-guid"))
									Expect(actualAppProtocol).To(Equal("http2"))
								})
							})
						})
					})
				})

				When("a tcp route is requested without a port", func() {
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
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal("some-domain.com"))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(cmd.Port))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))

							Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
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
						It("exits 0 with a helpful message that the route is already mapped to the app", func() {
							Expect(executeErr).ShouldNot(HaveOccurred())

							Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
							Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(actualAppName).To(Equal(appName))
							Expect(actualSpaceGUID).To(Equal(spaceGUID))

							Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
							actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
							Expect(actualDomain.Name).To(Equal("some-domain.com"))
							Expect(actualDomain.GUID).To(Equal("domain-guid"))
							Expect(actualHostname).To(Equal(hostname))
							Expect(actualPath).To(Equal(path))
							Expect(actualPort).To(Equal(cmd.Port))

							Expect(fakeActor.GetRouteDestinationByAppGUIDCallCount()).To(Equal(1))
							actualRoute, actualAppGUID := fakeActor.GetRouteDestinationByAppGUIDArgsForCall(0)
							Expect(actualRoute.GUID).To(Equal("route-guid"))
							Expect(actualAppGUID).To(Equal("app-guid"))
							Expect(fakeActor.MapRouteCallCount()).To(Equal(0))
						})

					})
					When("the destination is not found", func() {
						When("mapping the route errors", func() {
							BeforeEach(func() {
								fakeActor.MapRouteReturns(v7action.Warnings{"map-route-warnings"}, errors.New("map-route-error"))
							})

							It("returns the error and displays warnings", func() {
								Expect(testUI.Err).To(Say("get-domain-warnings"))
								Expect(testUI.Err).To(Say("get-app-warnings"))
								Expect(testUI.Err).To(Say("get-route-warnings"))
								Expect(testUI.Err).To(Say("map-route-warnings"))
								Expect(executeErr).To(MatchError(errors.New("map-route-error")))

								Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
								Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

								Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
								actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
								Expect(actualAppName).To(Equal(appName))
								Expect(actualSpaceGUID).To(Equal(spaceGUID))

								Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
								actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
								Expect(actualDomain.Name).To(Equal("some-domain.com"))
								Expect(actualDomain.GUID).To(Equal("domain-guid"))
								Expect(actualHostname).To(Equal(hostname))
								Expect(actualPath).To(Equal(path))
								Expect(actualPort).To(Equal(cmd.Port))

								Expect(fakeActor.MapRouteCallCount()).To(Equal(1))
								actualRouteGUID, actualAppGUID, actualAppProtocol := fakeActor.MapRouteArgsForCall(0)
								Expect(actualRouteGUID).To(Equal("route-guid"))
								Expect(actualAppGUID).To(Equal("app-guid"))
								Expect(actualAppProtocol).To(Equal("http2"))
							})
						})

						When("mapping the route succeeds", func() {
							BeforeEach(func() {
								fakeActor.MapRouteReturns(v7action.Warnings{"map-route-warnings"}, nil)
							})

							It("returns the proper information and passing the proper arguments", func() {
								By("displaying warnings", func() {
									Expect(testUI.Err).To(Say("get-domain-warnings"))
									Expect(testUI.Err).To(Say("get-app-warnings"))
									Expect(testUI.Err).To(Say("get-route-warnings"))
									Expect(testUI.Err).To(Say("map-route-warnings"))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								By("passing the expected arguments to the actor ", func() {
									Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
									Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domain))

									Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
									actualAppName, actualSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
									Expect(actualAppName).To(Equal(appName))
									Expect(actualSpaceGUID).To(Equal(spaceGUID))

									Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
									actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
									Expect(actualDomain.Name).To(Equal("some-domain.com"))
									Expect(actualDomain.GUID).To(Equal("domain-guid"))
									Expect(actualHostname).To(Equal("host"))
									Expect(actualPath).To(Equal(path))
									Expect(actualPort).To(Equal(cmd.Port))

									Expect(fakeActor.MapRouteCallCount()).To(Equal(1))
									actualRouteGUID, actualAppGUID, actualAppProtocol := fakeActor.MapRouteArgsForCall(0)
									Expect(actualRouteGUID).To(Equal("route-guid"))
									Expect(actualAppGUID).To(Equal("app-guid"))
									Expect(actualAppProtocol).To(Equal("http2"))
								})
							})
						})
					})
				})
			})
		})
	})
})
