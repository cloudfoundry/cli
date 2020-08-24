package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/flag"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/cli/actor/v7action"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("share-service Command", func() {
	var (
		cmd             ShareServiceCommand
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

		cmd = ShareServiceCommand{
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
			Expect(executeErr.Error()).To(ContainSubstring("space not targeted"))
		})
	})

	Context("user is logged in", func() {
		var (
			expectedServiceInstanceName = "fake-service-instance-name"
			expectedSpaceName           = "fake-space-name"
			expectedTargetedSpaceGuid   = "fake-space-guid"
			expectedTargetedOrgGuid     = "fake-org-guid"
		)

		BeforeEach(func() {
			cmd.RequiredArgs.ServiceInstance = expectedServiceInstanceName
			cmd.RequiredArgs.SpaceName = expectedSpaceName

			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: expectedTargetedSpaceGuid})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: expectedTargetedOrgGuid})
		})

		It("calls the actor to share in specified space and targeted org", func() {
			Expect(fakeActor.ShareServiceInstanceToSpaceAndOrgCallCount()).To(Equal(1))

			actualServiceInstance, actualTargetedSpace, actualTargetedOrg, actualSharingParams := fakeActor.ShareServiceInstanceToSpaceAndOrgArgsForCall(0)
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
				setFlag(&cmd, "-o", flag.OptionalString{IsSet: true, Value: expectedSpecifiedOrgName})
			})

			It("calls the actor to share in specified space and org", func() {
				Expect(fakeActor.ShareServiceInstanceToSpaceAndOrgCallCount()).To(Equal(1))

				actualServiceInstance, actualTargetedSpace, actualTargetedOrg, actualSharingParams := fakeActor.ShareServiceInstanceToSpaceAndOrgArgsForCall(0)
				Expect(actualServiceInstance).To(Equal(expectedServiceInstanceName))
				Expect(actualTargetedSpace).To(Equal(expectedTargetedSpaceGuid))
				Expect(actualTargetedOrg).To(Equal(expectedTargetedOrgGuid))
				Expect(actualSharingParams).To(Equal(v7action.ServiceInstanceSharingParams{
					SpaceName: expectedSpaceName,
					OrgName:   types.NewOptionalString(expectedSpecifiedOrgName),
				}))
			})
		})

		When("the actor errors", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				fakeActor.ShareServiceInstanceToSpaceAndOrgReturns(v7action.Warnings{}, errors.New("test error"))
			})

			It("fails with an error", func() {
				Expect(executeErr).To(Not(BeNil()))
				Expect(executeErr.Error()).To(ContainSubstring("test error"))
			})
		})
	})
})
