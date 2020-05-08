package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "code.cloudfoundry.org/cli/command/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-space-quota Command", func() {
	var (
		cmd             SetSpaceQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		input           *Buffer

		org     configv3.Organization
		orgGUID string

		spaceName      string
		spaceQuotaName string

		getSpaceWarning   string
		applyQuotaWarning string
		currentUser       string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		getSpaceWarning = RandomString("get-space-warning")
		applyQuotaWarning = RandomString("apply-quota-warning")

		orgGUID = RandomString("org-guid")
		spaceName = RandomString("space-name")
		spaceQuotaName = RandomString("space-quota-name")

		cmd = SetSpaceQuotaCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.SetSpaceQuotaArgs{
				Space:      spaceName,
				SpaceQuota: spaceQuotaName,
			},
		}

		org = configv3.Organization{
			GUID: orgGUID,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		currentUser = "current-user"
		fakeConfig.CurrentUserNameReturns(currentUser, nil)

		fakeConfig.TargetedOrganizationReturns(org)

		fakeActor.GetSpaceByNameAndOrganizationReturns(
			v7action.Space{GUID: "some-space-guid"},
			v7action.Warnings{getSpaceWarning},
			nil,
		)

		fakeActor.ApplySpaceQuotaByNameReturns(
			v7action.Warnings{applyQuotaWarning},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("check-target-failure"))
		})

		It("requires a targeted org (but not space)", func() {
			targetedOrgRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetedOrgRequired).To(BeTrue())
			Expect(targetedSpaceRequired).To(BeFalse())
		})

		It("should return the error", func() {
			Expect(executeErr).To(MatchError("check-target-failure"))
		})
	})

	It("signals what it's going to do", func() {
		Expect(testUI.Out).To(Say("Setting space quota %s to space %s as %s...",
			spaceQuotaName, spaceName, currentUser))
	})

	When("it can't get the space information", func() {
		BeforeEach(func() {
			fakeActor.GetSpaceByNameAndOrganizationReturns(
				v7action.Space{},
				v7action.Warnings{getSpaceWarning},
				actionerror.SpaceNotFoundError{Name: spaceName},
			)
		})

		It("returns an error", func() {
			Expect(testUI.Err).To(Say(getSpaceWarning))
			Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: spaceName}))
		})
	})

	It("passes the expected arguments when getting the space", func() {
		spaceNameArg, orgGUIDArg := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
		Expect(spaceNameArg).To(Equal(spaceName))
		Expect(orgGUIDArg).To(Equal(orgGUID))
	})

	When("it can't apply the space quota", func() {
		BeforeEach(func() {
			fakeActor.ApplySpaceQuotaByNameReturns(
				v7action.Warnings{getSpaceWarning},
				errors.New("some-apply-error"),
			)
		})

		It("returns an error", func() {
			Expect(testUI.Err).To(Say(getSpaceWarning))
			Expect(executeErr).To(MatchError("some-apply-error"))
		})
	})

	It("prints the warnings", func() {
		Expect(fakeActor.ApplySpaceQuotaByNameCallCount()).To(Equal(1))
		Expect(testUI.Err).To(Say(applyQuotaWarning))
	})

	It("prints 'OK'", func() {
		Expect(testUI.Out).To(Say("OK"))
	})

	It("does not error", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
	})
})
