package quota_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/quota"
	"github.com/cloudfoundry/cli/cf/errors"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app Command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		quotaRepo           *testapi.FakeQuotaRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		quotaRepo = &testapi.FakeQuotaRepository{}
	})

	runCommand := func(args ...string) {
		cmd := NewUpdateQuota(ui, testconfig.NewRepositoryWithDefaults(), quotaRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("cf-plays-dwarf-fortress")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("updating quota fields", func() {
		BeforeEach(func() {
			quotaRepo.FindByNameReturns.Quota = models.QuotaFields{
				Guid:          "quota-guid",
				Name:          "quota-name",
				MemoryLimit:   1024,
				RoutesLimit:   111,
				ServicesLimit: 222,
			}
		})

		Context("when the -m flag is provided", func() {
			It("updates the memory limit", func() {
				runCommand("-m", "15G", "quota-name")
				Expect(quotaRepo.UpdateCalledWith.Name).To(Equal("quota-name"))
				Expect(quotaRepo.UpdateCalledWith.MemoryLimit).To(Equal(uint64(15360)))
			})

			It("fails with usage when the value cannot be parsed", func() {
				runCommand("-m", "blasé", "le-tired")
				Expect(ui.FailedWithUsage).To(BeTrue())
			})
		})

		Context("when the -n flag is provided", func() {
			It("updates the quota name", func() {
				runCommand("-n", "quota-new-name", "quota-name")

				Expect(quotaRepo.UpdateCalledWith.Name).To(Equal("quota-new-name"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating quota", "quota-name", "as", "my-user"},
					[]string{"OK"},
				))
			})
		})

		It("updates the total allowed services", func() {
			runCommand("-s", "9000", "quota-name")
			Expect(quotaRepo.UpdateCalledWith.ServicesLimit).To(Equal(9000))
		})

		It("updates the total allowed routes", func() {
			runCommand("-r", "9001", "quota-name")
			Expect(quotaRepo.UpdateCalledWith.RoutesLimit).To(Equal(9001))
		})

		Context("update paid service plans", func() {
			It("changes to paid service plan when --allow flag is provided", func() {
				runCommand("--allow-paid-service-plans", "quota-name")
				Expect(quotaRepo.UpdateCalledWith.AllowPaidServicePlans).To(BeTrue())
			})

			It("changes to non-paid service plan when --disallow flag is provided", func() {
				quotaRepo.FindByNameReturns.Quota.NonBasicServicesAllowed = true // updating an existing quota

				runCommand("--disallow-paid-service-plans", "quota-name")
				Expect(quotaRepo.UpdateCalledWith.AllowPaidServicePlans).To(BeFalse())
			})

			It("shows an error when both --allow and --disallow flags are provided", func() {
				runCommand("--allow-paid-service-plans", "--disallow-paid-service-plans", "quota-name")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Both flags are not permitted"},
				))
			})
		})
	})

	It("shows an error when updating fails", func() {
		quotaRepo.UpdateReturns.Error = errors.New("I accidentally a quota")
		runCommand("-m", "1M", "dead-serious")
		Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
	})

	It("shows the user an error when finding the quota fails", func() {
		quotaRepo.FindByNameReturns.Error = errors.New("i can't believe it's not quotas!")

		runCommand("-m", "50Somethings", "what-could-possibly-go-wrong?")
		Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
	})

	It("shows a message explaining the update", func() {
		quotaRepo.FindByNameReturns.Quota.Name = "i-love-ui"

		runCommand("-m", "50G", "i-love-ui")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating quota", "i-love-ui", "as", "my-user"},
			[]string{"OK"},
		))
	})
})
