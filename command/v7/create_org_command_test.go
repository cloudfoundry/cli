package v7_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-org Command", func() {
	var (
		cmd             v7.CreateOrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error

		orgName         string
		quotaName       string
		currentUsername string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		orgName = "some-org"
		currentUsername = "bob"
		fakeActor.GetCurrentUserReturns(configv3.User{Name: currentUsername}, nil)
		quotaName = "quota-name"

		cmd = v7.CreateOrgCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Organization{Organization: orgName},
			Quota:        quotaName,
		}
	})

	JustBeforeEach(func() {
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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	It("prints text indicating it is creating a org", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(testUI.Out).To(Say(`Creating org %s as %s\.\.\.`, orgName, currentUsername))
	})

	When("creating the org errors", func() {
		BeforeEach(func() {
			fakeActor.CreateOrganizationReturns(resources.Organization{}, v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-org"))
		})

		It("returns an error and displays warnings", func() {
			Expect(executeErr).To(MatchError("err-create-org"))
			Expect(testUI.Err).To(Say("warnings-1"))
			Expect(testUI.Err).To(Say("warnings-2"))
		})
	})

	When("the org already exists", func() {
		BeforeEach(func() {
			fakeActor.CreateOrganizationReturns(
				resources.Organization{},
				v7action.Warnings{"some-warning"},
				ccerror.OrganizationNameTakenError{
					UnprocessableEntityError: ccerror.UnprocessableEntityError{
						Message: "Organization 'some-org' already exists.",
					},
				},
			)
		})

		It("displays all warnings, that the org already exists, and does not error", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Creating org %s as %s\.\.\.`, orgName, currentUsername))
			Expect(testUI.Out).To(Say(`Organization '%s' already exists\.`, orgName))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("applying the quota errors", func() {
		BeforeEach(func() {
			cmd.Quota = "quota-name"
			fakeActor.CreateOrganizationReturns(resources.Organization{Name: orgName, GUID: "some-org-guid"}, v7action.Warnings{}, nil)
			fakeActor.ApplyOrganizationQuotaByNameReturns(v7action.Warnings{"quota-warnings-1", "quota-warnings-2"}, errors.New("quota-error"))

		})

		It("returns an error and displays warnings", func() {
			Expect(executeErr).To(MatchError("quota-error"))
			Expect(testUI.Err).To(Say("quota-warnings-1"))
			Expect(testUI.Err).To(Say("quota-warnings-2"))
		})
	})

	When("the quota does not exist", func() {
		BeforeEach(func() {
			fakeActor.ApplyOrganizationQuotaByNameReturns(v7action.Warnings{"quota-warnings-1", "quota-warnings-2"}, actionerror.OrganizationQuotaNotFoundForNameError{Name: quotaName})
		})

		It("returns an error and displays warnings", func() {
			Expect(testUI.Err).To(Say("quota-warnings-1"))
			Expect(testUI.Err).To(Say("quota-warnings-2"))
			Expect(executeErr).To(MatchError(fmt.Sprintf("Organization quota with name '%s' not found.", quotaName)))
		})
	})

	When("creating the org is successful", func() {
		var orgGUID = "some-org-guid"
		BeforeEach(func() {
			fakeActor.CreateOrganizationReturns(resources.Organization{Name: orgName, GUID: orgGUID}, v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			fakeActor.ApplyOrganizationQuotaByNameReturns(v7action.Warnings{"quota-warnings-1", "quota-warnings-2"}, nil)
		})

		It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Creating org %s as %s\.\.\.`, orgName, currentUsername))
			Expect(testUI.Err).To(Say("warnings-1"))
			Expect(testUI.Err).To(Say("warnings-2"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Out).To(Say(`Setting org quota %s to org %s as %s\.\.\.`, quotaName, orgName, currentUsername))
			Expect(testUI.Err).To(Say("quota-warnings-1"))
			Expect(testUI.Err).To(Say("quota-warnings-2"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgName))
		})

		It("creates the org", func() {
			Expect(fakeActor.CreateOrganizationCallCount()).To(Equal(1))
			expectedOrgName := fakeActor.CreateOrganizationArgsForCall(0)
			Expect(expectedOrgName).To(Equal(orgName))
		})

		It("applies he quota to the org", func() {
			passedQuotaName, passedGuid := fakeActor.ApplyOrganizationQuotaByNameArgsForCall(0)
			Expect(passedQuotaName).To(Equal(quotaName))
			Expect(passedGuid).To(Equal(orgGUID))
		})
	})
})
