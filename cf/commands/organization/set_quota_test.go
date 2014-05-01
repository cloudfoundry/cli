package organization_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("set-quota command", func() {
	var (
		cmd                 *SetQuota
		ui                  *testterm.FakeUI
		quotaRepo           *testapi.FakeQuotaRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, testcmd.NewContext("set-quota", args), requirementsFactory)
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		quotaRepo = &testapi.FakeQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		cmd = NewSetQuota(ui, testconfig.NewRepositoryWithDefaults(), quotaRepo)
	})

	It("fails with usage when provided too many or two few args", func() {
		runCommand("org")
		Expect(ui.FailedWithUsage).To(BeTrue())

		runCommand("org", "quota", "extra-stuff")
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		runCommand("my-org", "my-quota")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("passes requirements when provided two args", func() {
			runCommand("my-org", "my-quota")
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			Expect(requirementsFactory.OrganizationName).To(Equal("my-org"))
		})

		It("assigns a quota to an org", func() {
			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			quota := models.QuotaFields{Name: "my-quota", Guid: "my-quota-guid"}

			quotaRepo.FindByNameReturns.Quota = quota
			requirementsFactory.Organization = org

			runCommand("my-org", "my-quota")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Setting quota", "my-quota", "my-org", "my-user"},
				[]string{"OK"},
			))

			Expect(quotaRepo.FindByNameCalledWith.Name).To(Equal("my-quota"))
			Expect(quotaRepo.AssignQuotaToOrgCalledWith.OrgGuid).To(Equal("my-org-guid"))
			Expect(quotaRepo.AssignQuotaToOrgCalledWith.QuotaGuid).To(Equal("my-quota-guid"))
		})
	})
})
