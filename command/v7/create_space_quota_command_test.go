package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-space-quota Command", func() {
	var (
		cmd             v7.CreateSpaceQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateSpaceQuotaActor
		binaryName      string
		executeErr      error

		userName       string
		spaceQuotaName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateSpaceQuotaActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		spaceQuotaName = "some-space-quota"
		userName = "some-user-name"

		cmd = v7.CreateSpaceQuotaCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.SpaceQuota{SpaceQuota: spaceQuotaName},
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
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org-name",
				GUID: "some-org-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{
				Name:   userName,
				Origin: "some-user-origin",
			}, nil)
		})

		It("prints text indicating it is creating a space quota", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating space quota %s for org %s as %s\.\.\.`, spaceQuotaName, "some-org-name", userName))
		})

		When("creating the space quota errors", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceQuotaReturns(
					v7action.Warnings{"warnings-1", "warnings-2"},
					errors.New("err-create-space-quota"),
				)
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-space-quota"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("no flag limits are given", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceQuotaReturns(
					v7action.Warnings{"warnings-1", "warnings-2"},
					nil,
				)
			})

			It("creates the space quota in the targeted organization", func() {
				Expect(fakeActor.CreateSpaceQuotaCallCount()).To(Equal(1))
				expectedSpaceQuotaName, expectedOrgGUID, expectedLimits := fakeActor.CreateSpaceQuotaArgsForCall(0)
				Expect(expectedSpaceQuotaName).To(Equal(spaceQuotaName))
				Expect(expectedOrgGUID).To(Equal("some-org-guid"))
				Expect(expectedLimits).To(Equal(v7action.QuotaLimits{}))
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("Creating space quota some-space-quota for org some-org-name as some-user-name..."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("all flag limits are given", func() {
			BeforeEach(func() {
				cmd.TotalMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 47}
				cmd.PerProcessMemory = flag.MemoryWithUnlimited{IsSet: true, Value: 23}
				cmd.NumAppInstances = flag.IntegerLimit{IsSet: true, Value: 4}
				cmd.PaidServicePlans = true
				cmd.TotalServiceInstances = flag.IntegerLimit{IsSet: true, Value: 9}
				cmd.TotalRoutes = flag.IntegerLimit{IsSet: true, Value: 1}
				cmd.TotalReservedPorts = flag.IntegerLimit{IsSet: true, Value: 7}

				fakeActor.CreateSpaceQuotaReturns(
					v7action.Warnings{"warnings-1", "warnings-2"},
					nil,
				)
			})

			It("creates the space quota with the specified limits in the targeted organization", func() {
				Expect(fakeActor.CreateSpaceQuotaCallCount()).To(Equal(1))
				expectedSpaceQuotaName, expectedOrgGUID, expectedLimits := fakeActor.CreateSpaceQuotaArgsForCall(0)
				Expect(expectedSpaceQuotaName).To(Equal(spaceQuotaName))
				Expect(expectedOrgGUID).To(Equal("some-org-guid"))
				Expect(expectedLimits).To(Equal(v7action.QuotaLimits{
					TotalMemoryInMB:       types.NullInt{IsSet: true, Value: 47},
					PerProcessMemoryInMB:  types.NullInt{IsSet: true, Value: 23},
					TotalInstances:        types.NullInt{IsSet: true, Value: 4},
					PaidServicesAllowed:   true,
					TotalServiceInstances: types.NullInt{IsSet: true, Value: 9},
					TotalRoutes:           types.NullInt{IsSet: true, Value: 1},
					TotalReservedPorts:    types.NullInt{IsSet: true, Value: 7},
				}))
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("Creating space quota some-space-quota for org some-org-name as some-user-name..."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the space quota already exists", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceQuotaReturns(v7action.Warnings{"some-warning"}, ccerror.QuotaAlreadyExists{Message: "yikes"})
			})

			It("displays all warnings, that the space quota already exists, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Creating space quota %s for org %s as %s\.\.\.`, spaceQuotaName, "some-org-name", userName))
				Expect(testUI.Err).To(Say(`yikes`))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
