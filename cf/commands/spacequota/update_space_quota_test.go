package spacequota_test

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/spacequota"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("update-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		requirementsFactory *testreq.FakeReqFactory

		quota            models.SpaceQuota
		quotaPaidService models.SpaceQuota
	)

	runCommand := func(args ...string) bool {
		cmd := NewUpdateSpaceQuota(ui, configuration.NewRepositoryWithDefaults(), quotaRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand("my-quota", "-m", "50G")).NotTo(HavePassedRequirements())
		})

		It("fails when the user does not have an org targeted", func() {
			requirementsFactory.TargetedOrgSuccess = false
			requirementsFactory.LoginSuccess = true
			Expect(runCommand()).NotTo(HavePassedRequirements())
			Expect(runCommand("my-quota", "-m", "50G")).NotTo(HavePassedRequirements())
		})

		It("fails with usage if space quota name is not provided", func() {
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.LoginSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			quota = models.SpaceQuota{
				Guid:                    "my-quota-guid",
				Name:                    "my-quota",
				MemoryLimit:             1024,
				InstanceMemoryLimit:     512,
				RoutesLimit:             111,
				ServicesLimit:           222,
				NonBasicServicesAllowed: false,
				OrgGuid:                 "my-org-guid",
			}

			quotaPaidService = models.SpaceQuota{NonBasicServicesAllowed: true}

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
		})

		JustBeforeEach(func() {
			quotaRepo.FindByNameReturns(quota, nil)
		})

		Context("when the -m flag is provided", func() {
			It("updates the memory limit", func() {
				runCommand("-m", "15G", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("my-quota"))
				Expect(quotaRepo.UpdateArgsForCall(0).MemoryLimit).To(Equal(int64(15360)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-m", "whoops", "wit mah hussle", "my-org")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the -i flag is provided", func() {
			It("sets the memory limit", func() {
				runCommand("-i", "50G", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(51200)))
			})

			It("sets the memory limit to -1", func() {
				runCommand("-i", "-1", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-i", "whoops", "my-quota")
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the -r flag is provided", func() {
			It("sets the route limit", func() {
				runCommand("-r", "12", "ecstatic")
				Expect(quotaRepo.UpdateArgsForCall(0).RoutesLimit).To(Equal(12))
			})
		})

		Context("when the -s flag is provided", func() {
			It("sets the service instance limit", func() {
				runCommand("-s", "42", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).ServicesLimit).To(Equal(42))
			})
		})

		Context("when the -n flag is provided", func() {
			It("sets the service instance name", func() {
				runCommand("-n", "foo", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("foo"))
			})
		})

		Context("when --allow-non-basic-services is provided", func() {
			It("updates the quota to allow paid service plans", func() {
				runCommand("--allow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeTrue())
			})
		})

		Context("when --disallow-non-basic-services is provided", func() {
			It("updates the quota to disallow paid service plans", func() {
				quotaRepo.FindByNameReturns(quotaPaidService, nil)

				runCommand("--disallow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeFalse())
			})
		})

		Context("when updating a quota returns an error", func() {
			It("alerts the user when creating the quota fails", func() {
				quotaRepo.UpdateReturns(errors.New("WHOOP THERE IT IS"))
				runCommand("my-quota")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating space quota", "my-quota", "my-user"},
					[]string{"FAILED"},
				))
			})

			It("fails if the allow and disallow flag are both passed", func() {
				runCommand("--disallow-paid-service-plans", "--allow-paid-service-plans", "my-for-profit-quota")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
