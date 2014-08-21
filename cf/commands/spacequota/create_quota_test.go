package spacequota_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/spacequota"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

var _ = Describe("create-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		orgRepo             *testapi.FakeOrgRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		orgRepo = &testapi.FakeOrgRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}

		org := models.Organization{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		orgRepo.Organizations = []models.Organization{org}
		orgRepo.FindByNameName = org.Name
	})

	runCommand := func(args ...string) {
		cmd := NewCreateSpaceQuota(ui, configuration.NewRepositoryWithDefaults(), quotaRepo, orgRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("my-quota", "-m", "50G")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("requires the user to target an org", func() {
			requirementsFactory.TargetedOrgSuccess = false
			runCommand("my-quota", "-m", "50G")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when requirements have been met", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
		})

		It("fails requirements when called without a quota name", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("creates a quota with a given name", func() {
			runCommand("my-quota")
			Expect(quotaRepo.CreateArgsForCall(0).Name).To(Equal("my-quota"))
			Expect(quotaRepo.CreateArgsForCall(0).OrgGuid).To(Equal("my-org-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating quota", "my-org", "my-quota", "my-user", "..."},
				[]string{"OK"},
			))
		})

		Context("when the -m flag is provided", func() {
			It("sets the memory limit", func() {
				runCommand("-m", "50G", "erryday makin fitty jeez")
				Expect(quotaRepo.CreateArgsForCall(0).MemoryLimit).To(Equal(int64(51200)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-m", "whoops", "wit mah hussle")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the -i flag is provided", func() {
			It("sets the memory limit", func() {
				runCommand("-i", "50G", "erryday makin fitty jeez")
				Expect(quotaRepo.CreateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(51200)))
			})

			It("accepts -1 without units as an appropriate value", func() {
				runCommand("-i", "-1", "wit mah hussle")
				Expect(quotaRepo.CreateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-i", "whoops", "yo", "12")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		It("sets the route limit", func() {
			runCommand("-r", "12", "ecstatic")

			Expect(quotaRepo.CreateArgsForCall(0).RoutesLimit).To(Equal(12))
		})

		It("sets the service instance limit", func() {
			runCommand("-s", "42", "black star")
			Expect(quotaRepo.CreateArgsForCall(0).ServicesLimit).To(Equal(42))
		})

		It("defaults to not allowing paid service plans", func() {
			runCommand("my-pro-bono-quota")
			Expect(quotaRepo.CreateArgsForCall(0).NonBasicServicesAllowed).To(BeFalse())
		})

		Context("when requesting to allow paid service plans", func() {
			It("creates the quota with paid service plans allowed", func() {
				runCommand("--allow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.CreateArgsForCall(0).NonBasicServicesAllowed).To(BeTrue())
			})
		})

		Context("when creating a quota returns an error", func() {
			It("alerts the user when creating the quota fails", func() {
				quotaRepo.CreateReturns(errors.New("WHOOP THERE IT IS"))
				runCommand("my-quota")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating quota", "my-quota", "my-org"},
					[]string{"FAILED"},
				))
			})

			It("warns the user when quota already exists", func() {
				quotaRepo.CreateReturns(errors.NewHttpError(400, "240002", "Quota Definition is taken: quota-sct"))
				runCommand("Banana")

				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"FAILED"},
				))
				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"already exists"}))
			})

		})
	})
})
