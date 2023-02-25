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

var _ = Describe("share-route Command", func() {
	var (
		cmd             v7.ShareRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		domainName      string
		orgName         string
		spaceName       string
		hostname        string
		path            string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "myBinaryBread"
		fakeConfig.BinaryNameReturns(binaryName)

		domainName = "some-domain.com"
		orgName = "org-name-a"
		spaceName = "space-name-a"
		hostname = "myHostname"
		path = "myPath"

		cmd = v7.ShareRouteCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequireArgs:      flag.Domain{Domain: domainName},
			Hostname:         hostname,
			Path:             flag.V7RoutePath{Path: path},
			DestinationOrg:   orgName,
			DestinationSpace: spaceName,
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks that target", func() {
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

				Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(0))

				Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(0))

				Expect(fakeActor.ShareRouteCallCount()).To(Equal(0))
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
				When("getting the target space errors", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							resources.Organization{GUID: "org-guid-a"},
							v7action.Warnings{"get-route-warnings"},
							nil,
						)
						fakeActor.GetSpaceByNameAndOrganizationReturns(
							resources.Space{},
							v7action.Warnings{"get-route-warnings"},
							actionerror.SpaceNotFoundError{},
						)
					})
					It("returns the error and warnings", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal(domainName))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))

						Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetOrganizationByNameArgsForCall(0)).To(Equal(orgName))
						Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
						spaceName, orgGuid := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
						Expect(spaceName).To(Equal("space-name-a"))
						Expect(orgGuid).To(Equal("org-guid-a"))

						Expect(fakeActor.ShareRouteCallCount()).To(Equal(0))
					})
				})
				When("getting the target org errors", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							resources.Organization{},
							v7action.Warnings{"get-route-warnings"},
							actionerror.OrganizationNotFoundError{},
						)
					})
					It("returns the error and warnings", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal(domainName))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))

						Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
						Expect(orgName).To(Equal("org-name-a"))

						Expect(fakeActor.ShareRouteCallCount()).To(Equal(0))
					})
				})
				When("getting the target space succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							resources.Organization{GUID: "org-guid-a"},
							v7action.Warnings{"get-route-warnings"},
							nil,
						)
						fakeActor.GetSpaceByNameAndOrganizationReturns(
							resources.Space{GUID: "space-guid-b"},
							v7action.Warnings{"get-route-warnings"},
							nil,
						)
					})
					It("exits 0 with helpful message that the route is now being shared", func() {
						Expect(executeErr).ShouldNot(HaveOccurred())

						Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
						Expect(fakeActor.GetDomainByNameArgsForCall(0)).To(Equal(domainName))

						Expect(fakeActor.GetRouteByAttributesCallCount()).To(Equal(1))
						actualDomain, actualHostname, actualPath, actualPort := fakeActor.GetRouteByAttributesArgsForCall(0)
						Expect(actualDomain.Name).To(Equal(domainName))
						Expect(actualDomain.GUID).To(Equal("domain-guid"))
						Expect(actualHostname).To(Equal(hostname))
						Expect(actualPath).To(Equal(path))
						Expect(actualPort).To(Equal(0))

						Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
						orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
						Expect(orgName).To(Equal("org-name-a"))

						Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
						spaceName, orgGuid := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
						Expect(spaceName).To(Equal("space-name-a"))
						Expect(orgGuid).To(Equal("org-guid-a"))
						Expect(fakeActor.ShareRouteCallCount()).To(Equal(1))
					})
				})
			})
		})
	})
})
