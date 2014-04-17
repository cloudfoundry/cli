package quota_test

import (
	. "cf/commands/quota"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "testhelpers/matchers"

	"cf/errors"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("quotas command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *testapi.FakeQuotaRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &testapi.FakeQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func() {
		cmd := NewListQuotas(ui, testconfig.NewRepositoryWithDefaults(), quotaRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("create-quota", []string{}), requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false

			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when quotas exist", func() {
		BeforeEach(func() {
			quotaRepo.FindAllReturns.Quotas = []models.QuotaFields{
				models.QuotaFields{
					Name:                    "quota-name",
					MemoryLimit:             1024,
					RoutesLimit:             111,
					ServicesLimit:           222,
					NonBasicServicesAllowed: true,
				},
				models.QuotaFields{
					Name:                    "quota-non-basic-not-allowed",
					MemoryLimit:             434,
					RoutesLimit:             1,
					ServicesLimit:           2,
					NonBasicServicesAllowed: false,
				},
			}
		})

		It("lists quotas", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting quotas as", "my-user"},
				[]string{"OK"},
				[]string{"name", "memory limit", "routes", "service instances", "paid service plans"},
				[]string{"quota-name", "1G", "111", "222", "allowed"},
				[]string{"quota-non-basic-not-allowed", "434M", "1", "2", "disallowed"},
			))
		})
	})

	Context("when an error occurs fetching quotas", func() {
		BeforeEach(func() {
			quotaRepo.FindAllReturns.Error = errors.New("I haz a borken!")
		})

		It("prints an error", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting quotas as", "my-user"},
				[]string{"FAILED"},
			))
		})
	})

})
