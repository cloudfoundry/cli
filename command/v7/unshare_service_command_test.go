package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unshare-service Command", func() {
	var (
		cmd             UnshareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UnshareServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	Context("user not targeting space", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("space not targeted"))
		})

		It("fails the command", func() {
			Expect(executeErr).To(Not(BeNil()))
			Expect(executeErr).To(MatchError("space not targeted"))
		})
	})

	Context("user is targeting a space and org", func() {
		var (
			expectedServiceInstanceName = "fake-service-instance-name"
			expectedSpaceName           = "fake-space-name"
			expectedTargetedSpaceGuid   = "fake-space-guid"
			expectedTargetedOrgName     = "fake-org-name"
			expectedTargetedOrgGuid     = "fake-org-guid"
			expectedUser                = "fake-username"
		)

		BeforeEach(func() {
			cmd.RequiredArgs.ServiceInstance = expectedServiceInstanceName
			setFlag(&cmd, "-s", expectedSpaceName)

			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: expectedTargetedSpaceGuid})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: expectedTargetedOrgGuid, Name: expectedTargetedOrgName})
			fakeConfig.CurrentUserReturns(configv3.User{Name: expectedUser}, nil)
		})

		When("the unshare completes successfully", func() {
			BeforeEach(func() {
				fakeActor.UnshareServiceInstanceFromSpaceAndOrgReturns(v7action.Warnings{"warning one", "warning two"}, nil)
			})

			It("returns an OK message", func() {
				Expect(executeErr).To(BeNil())

				Expect(testUI.Out).To(
					Say(`Unsharing service instance %s from org %s / space %s as %s`,
						expectedServiceInstanceName,
						expectedTargetedOrgName,
						expectedSpaceName,
						expectedUser))
				Expect(testUI.Out).To(Say(`OK`))
				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})

		It("calls the actor to share in specified space and targeted org", func() {
			Expect(fakeActor.UnshareServiceInstanceFromSpaceAndOrgCallCount()).To(Equal(1))

			actualServiceInstance, actualTargetedSpace, actualTargetedOrg, actualSharingParams := fakeActor.UnshareServiceInstanceFromSpaceAndOrgArgsForCall(0)
			Expect(actualServiceInstance).To(Equal(expectedServiceInstanceName))
			Expect(actualTargetedSpace).To(Equal(expectedTargetedSpaceGuid))
			Expect(actualTargetedOrg).To(Equal(expectedTargetedOrgGuid))
			Expect(actualSharingParams).To(Equal(v7action.ServiceInstanceSharingParams{
				SpaceName: expectedSpaceName,
				OrgName:   types.OptionalString{},
			}))
		})

		When("organization flag is specified", func() {
			var expectedSpecifiedOrgName = "fake-org-name"

			BeforeEach(func() {
				setFlag(&cmd, "-o", types.NewOptionalString(expectedSpecifiedOrgName))
			})

			It("calls the actor to share in specified space and org", func() {
				Expect(fakeActor.UnshareServiceInstanceFromSpaceAndOrgCallCount()).To(Equal(1))

				_, _, _, actualSharingParams := fakeActor.UnshareServiceInstanceFromSpaceAndOrgArgsForCall(0)
				Expect(actualSharingParams).To(Equal(v7action.ServiceInstanceSharingParams{
					SpaceName: expectedSpaceName,
					OrgName:   types.NewOptionalString(expectedSpecifiedOrgName),
				}))
			})
		})

		When("the actor errors", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				fakeActor.UnshareServiceInstanceFromSpaceAndOrgReturns(v7action.Warnings{"actor warning"}, errors.New("test error"))
			})

			It("prints all warnings and fails with an error", func() {
				Expect(executeErr).To(Not(BeNil()))
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError("test error"))

			})
		})
	})

})
