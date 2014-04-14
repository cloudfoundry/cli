package quota_test

import (
	. "cf/commands/quota"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "testhelpers/matchers"

	"cf/errors"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	"testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
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
				[]string{"creating quota", "my-quota", "my-user"},
				[]string{"Ok"},
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

		It("alerts the user when creating the quota fails", func() {
			quotaRepo.CreateReturns.Error = errors.New("WHOOP THERE IT IS")
			runCommand("my-quota")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"creating quota", "my-quota"},
				[]string{"FAILED"},
			))
		})
	})
})
