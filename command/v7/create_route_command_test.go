package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
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

var _ = Describe("create-route Command", func() {
	var (
		cmd             CreateRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateRouteActor

		executeErr error

		binaryName string
		domainName string
		spaceName  string
		orgName    string
		hostname   string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateRouteActor)

		domainName = "example.com"
		spaceName = "space"
		orgName = "org"
		hostname = ""

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		cmd = CreateRouteCommand{
			RequiredArgs: flag.Domain{
				Domain: domainName,
			},
			Hostname:    hostname,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: spaceName,
				GUID: "some-space-guid",
			})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: orgName,
				GUID: "some-org-guid",
			})
		})

		It("should print text indicating it is creating a route", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating route %s for org %s / space %s as the-user\.\.\.`, domainName, orgName, spaceName))
		})

		When("passing in a hostname", func() {
			BeforeEach(func() {
				hostname = "flan"
			})

			It("should print text indicating it is creating a route", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(`Creating route %s\.%s for org %s / space %s as the-user\.\.\.`, hostname, domainName, orgName, spaceName))
			})
		})

		When("creating the route errors", func() {
			BeforeEach(func() {
				fakeActor.CreateRouteReturns(v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-route"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-route"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the route is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateRouteReturns(v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say(`Route %s has been created.`, domainName))
				Expect(testUI.Out).To(Say("OK"))
			})

			It("creates the route", func() {
				Expect(fakeActor.CreateRouteCallCount()).To(Equal(1))
				expectedOrgName, expectedSpaceName, expectedDomainName, expectedHostname := fakeActor.CreateRouteArgsForCall(0)
				Expect(expectedOrgName).To(Equal(orgName))
				Expect(expectedDomainName).To(Equal(domainName))
				Expect(expectedSpaceName).To(Equal(spaceName))
				Expect(expectedHostname).To(Equal(hostname))
			})

			When("passing in a hostname", func() {
				BeforeEach(func() {
					hostname = "flan"
				})

				It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("warnings-1"))
					Expect(testUI.Err).To(Say("warnings-2"))
					Expect(testUI.Out).To(Say(`Route %s\.%s has been created.`, hostname, domainName))
					Expect(testUI.Out).To(Say("OK"))
				})

				It("creates the route", func() {
					Expect(fakeActor.CreateRouteCallCount()).To(Equal(1))
					expectedOrgName, expectedSpaceName, expectedDomainName, expectedHostname := fakeActor.CreateRouteArgsForCall(0)
					Expect(expectedOrgName).To(Equal(orgName))
					Expect(expectedDomainName).To(Equal(domainName))
					Expect(expectedSpaceName).To(Equal(spaceName))
					Expect(expectedHostname).To(Equal(hostname))
				})
			})
		})

		When("the route already exists", func() {
			BeforeEach(func() {
				fakeActor.CreateRouteReturns(v7action.Warnings{"some-warning"}, actionerror.RouteAlreadyExistsError{Err: errors.New("api error for a route that already exists")})
			})

			It("displays all warnings, that the route already exists, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Creating route %s for org %s / space %s as the-user\.\.\.`, domainName, orgName, spaceName))
				Expect(testUI.Out).To(Say("api error for a route that already exists"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
