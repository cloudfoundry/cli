package quota_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/quota"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/errors"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

var _ = Describe("create-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *testapi.FakeQuotaRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &testapi.FakeQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		cmd := NewCreateQuota(ui, configuration.NewRepositoryWithDefaults(), quotaRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("create-quota", args), requirementsFactory)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("fails requirements", func() {
			runCommand("my-quota", "-m", "50G")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails requirements when called without a quota name", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("creates a quota with a given name", func() {
			runCommand("my-quota")
			Expect(quotaRepo.CreateCalledWith.Name).To(Equal("my-quota"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating quota", "my-quota", "my-user", "..."},
				[]string{"OK"},
			))
		})

		Context("when the -m flag is provided", func() {
			It("sets the memory limit", func() {
				runCommand("-m", "50G", "erryday makin fitty jeez")
				Expect(quotaRepo.CreateCalledWith.MemoryLimit).To(Equal(uint64(51200)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("whoops", "12")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		It("sets the route limit", func() {
			runCommand("-r", "12", "ecstatic")

			Expect(quotaRepo.CreateCalledWith.RoutesLimit).To(Equal(12))
		})

		It("sets the service instance limit", func() {
			runCommand("-s", "42", "black star")
			Expect(quotaRepo.CreateCalledWith.ServicesLimit).To(Equal(42))
		})

		It("defaults to not allowing paid service plans", func() {
			runCommand("my-pro-bono-quota")
			Expect(quotaRepo.CreateCalledWith.NonBasicServicesAllowed).To(BeFalse())
		})

		Context("when requesting to allow paid service plans", func() {
			It("creates the quota with paid service plans allowed", func() {
				runCommand("--allow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.CreateCalledWith.NonBasicServicesAllowed).To(BeTrue())
			})
		})

		Context("when creating a quota returns an error", func() {
			It("alerts the user when creating the quota fails", func() {
				quotaRepo.CreateReturns.Error = errors.New("WHOOP THERE IT IS")
				runCommand("my-quota")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating quota", "my-quota"},
					[]string{"FAILED"},
				))
			})

			It("warns the user when quota already exists", func() {
				quotaRepo.CreateReturns.Error = errors.NewHttpError(400, "240002", "Quota Definition is taken: quota-sct")
				runCommand("Banana")

				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"FAILED"},
				))
				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"already exists"}))
			})

		})
	})
})
