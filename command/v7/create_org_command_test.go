package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-org Command", func() {
	var (
		cmd             v7.CreateOrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateOrgActor
		binaryName      string
		executeErr      error

		orgName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateOrgActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		orgName = "some-org"
	})

	JustBeforeEach(func() {
		cmd = v7.CreateOrgCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.Organization{Organization: orgName},
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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("prints text indicating it is creating a org", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating org %s as the-user\.\.\.`, orgName))
		})

		When("creating the org errors", func() {
			BeforeEach(func() {
				fakeActor.CreateOrganizationReturns(v7action.Organization{}, v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-org"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-org"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the org is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateOrganizationReturns(v7action.Organization{Name: orgName}, v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			})

			//TODO: modify or remove these tests upon set-org-role implementation. they are included in the tests commented out below
			It("creates the org", func() {
				Expect(fakeActor.CreateOrganizationCallCount()).To(Equal(1))
				expectedOrgName := fakeActor.CreateOrganizationArgsForCall(0)
				Expect(expectedOrgName).To(Equal(orgName))
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgName))
			})

			//TODO: add tests for setting org roles once V7/V3 set-org-role is implemented and included in create-org
		})

		When("the org already exists", func() {
			BeforeEach(func() {
				fakeActor.CreateOrganizationReturns(
					v7action.Organization{},
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
				Expect(testUI.Out).To(Say(`Creating org %s as the-user\.\.\.`, orgName))
				Expect(testUI.Out).To(Say(`Organization '%s' already exists\.`, orgName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
