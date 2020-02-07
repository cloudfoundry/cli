package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("space-quotas command", func() {
	var (
		cmd             SpaceQuotasCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSpaceQuotasActor
		executeErr      error
		args            []string
		binaryName      string
		trueValue       = true
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSpaceQuotasActor)
		args = nil

		cmd = SpaceQuotasCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "targeted-org-guid"})
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	When("running the command successfully", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "apple"}, nil)
			spaceQuotas := []resources.SpaceQuota{
				{
					Quota: resources.Quota{
						Name: "space-quota-1",
						Apps: resources.AppLimit{
							TotalMemory:       &types.NullInt{Value: 1048576, IsSet: true},
							InstanceMemory:    &types.NullInt{Value: 32, IsSet: true},
							TotalAppInstances: &types.NullInt{Value: 3, IsSet: true},
						},
						Services: resources.ServiceLimit{
							TotalServiceInstances: &types.NullInt{Value: 3, IsSet: true},
							PaidServicePlans:      &trueValue,
						},
						Routes: resources.RouteLimit{
							TotalRoutes:        &types.NullInt{Value: 5, IsSet: true},
							TotalReservedPorts: &types.NullInt{Value: 2, IsSet: true},
						},
					},
				},
			}
			fakeActor.GetSpaceQuotasByOrgGUIDReturns(spaceQuotas, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
		})

		It("should print text indicating the command status", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Getting space quotas as apple\.\.\.`))
			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})

		It("retrieves and displays all quotas", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeActor.GetSpaceQuotasByOrgGUIDCallCount()).To(Equal(1))
			orgGUID := fakeActor.GetSpaceQuotasByOrgGUIDArgsForCall(0)
			Expect(orgGUID).To(Equal(fakeConfig.TargetedOrganization().GUID))

			Expect(testUI.Out).To(Say(`name\s+total memory\s+instance memory\s+routes\s+service instances\s+paid service plans\s+app instances\s+route ports`))
			Expect(testUI.Out).To(Say(`space-quota-1\s+1T\s+32M\s+5\s+3\s+allowed\s+3\s+2`))
		})

		When("there are limits that have not been configured", func() {
			BeforeEach(func() {
				spaceQuotas := []resources.SpaceQuota{
					{
						Quota: resources.Quota{
							Name: "default",
							Apps: resources.AppLimit{
								TotalMemory:       &types.NullInt{Value: 0, IsSet: false},
								InstanceMemory:    &types.NullInt{Value: 0, IsSet: false},
								TotalAppInstances: &types.NullInt{Value: 0, IsSet: false},
							},
							Services: resources.ServiceLimit{
								TotalServiceInstances: &types.NullInt{Value: 0, IsSet: false},
								PaidServicePlans:      &trueValue,
							},
							Routes: resources.RouteLimit{
								TotalRoutes:        &types.NullInt{Value: 0, IsSet: false},
								TotalReservedPorts: &types.NullInt{Value: 0, IsSet: false},
							},
						},
					},
				}
				fakeActor.GetSpaceQuotasByOrgGUIDReturns(spaceQuotas, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)

			})

			It("should convert default values from the API into readable outputs", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say(`name\s+total memory\s+instance memory\s+routes\s+service instances\s+paid service plans\s+app instances\s+route ports`))
				Expect(testUI.Out).To(Say(`default\s+unlimited\s+unlimited\s+unlimited\s+unlimited\s+allowed\s+unlimited\s+unlimited`))

			})
		})
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

	When("getting quotas fails", func() {
		BeforeEach(func() {
			fakeActor.GetSpaceQuotasByOrgGUIDReturns(nil, v7action.Warnings{"some-warning-1", "some-warning-2"}, errors.New("some-error"))
		})

		It("prints warnings and returns error", func() {
			Expect(executeErr).To(MatchError("some-error"))

			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})
	})

	When("the quota list is empty", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "apple"}, nil)
			fakeActor.GetSpaceQuotasByOrgGUIDReturns([]resources.SpaceQuota{}, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
		})

		It("prints warnings and returns error", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
			Expect(testUI.Out).To(Say(`Getting space quotas as apple\.\.\.`))
			Expect(testUI.Out).To(Say("No space quotas found."))
		})
	})
})
