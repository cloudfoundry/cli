package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("org-quotas command", func() {
	var (
		cmd             OrgQuotasCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeOrgQuotasActor
		executeErr      error
		args            []string
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeOrgQuotasActor)
		args = nil

		cmd = OrgQuotasCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	When("running the command successfully", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "apple"}, nil)
			orgQuotas := []v7action.OrganizationQuota{
				{
					Name: "org-quota-1",
					Apps: ccv3.AppLimit{
						TotalMemory:       types.NullInt{Value: 1048576, IsSet: true},
						InstanceMemory:    types.NullInt{Value: 1024, IsSet: true},
						TotalAppInstances: types.NullInt{Value: 3, IsSet: true},
					},
					Services: ccv3.ServiceLimit{
						TotalServiceInstances: types.NullInt{Value: 3, IsSet: true},
						PaidServicePlans:      true,
					},
					Routes: ccv3.RouteLimit{
						TotalRoutes:     types.NullInt{Value: 5, IsSet: true},
						TotalRoutePorts: types.NullInt{Value: 2, IsSet: true},
					},
				},
			}
			fakeActor.GetOrganizationQuotasReturns(orgQuotas, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
		})

		It("should print text indicating its running", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Getting org quotas as apple\.\.\.`))
		})

		It("prints a table of buildpacks", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
			Expect(testUI.Out).To(Say(`name\s+total memory\s+instance memory\s+routes\s+service instances\s+paid service plans\s+app instances\s+route ports`))
			Expect(testUI.Out).To(Say(`org-quota-1\s+1048576\s+1024\s+5\s+3\s+allowed\s+3\s+2`))
		})

		When("there are limits that have not been configured", func() {
			BeforeEach(func() {
				orgQuotas := []v7action.OrganizationQuota{
					{
						Name: "default",
						Apps: ccv3.AppLimit{
							TotalMemory:       types.NullInt{Value: 0, IsSet: false},
							InstanceMemory:    types.NullInt{Value: 0, IsSet: false},
							TotalAppInstances: types.NullInt{Value: 0, IsSet: false},
						},
						Services: ccv3.ServiceLimit{
							TotalServiceInstances: types.NullInt{Value: 0, IsSet: false},
							PaidServicePlans:      true,
						},
						Routes: ccv3.RouteLimit{
							TotalRoutes:     types.NullInt{Value: 0, IsSet: false},
							TotalRoutePorts: types.NullInt{Value: 0, IsSet: false},
						},
					},
				}
				fakeActor.GetOrganizationQuotasReturns(orgQuotas, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)

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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("getting buildpacks fails", func() {
		BeforeEach(func() {
			fakeActor.GetOrganizationQuotasReturns(nil, v7action.Warnings{"some-warning-1", "some-warning-2"}, errors.New("some-error"))
		})

		It("prints warnings and returns error", func() {
			Expect(executeErr).To(MatchError("some-error"))

			Expect(testUI.Err).To(Say("some-warning-1"))
			Expect(testUI.Err).To(Say("some-warning-2"))
		})
	})
})
