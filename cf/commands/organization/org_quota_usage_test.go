package organization_test

import (
	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("org command", func() {
	var (
		ui                  *testterm.FakeUI
		orgRepo             *test_org.FakeOrganizationRepository
		configRepo          core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		quotaUsage          models.QuotaUsage
	)

	runCommand := func(args ...string) bool {
		cmd := organization.NewQuotaUsage(ui, configRepo, orgRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		orgRepo = &test_org.FakeOrganizationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(runCommand()).To(BeFalse())
		})
		It("should fail with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("too", "much")).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

	})

	Context("when there are quota usage to be listed", func() {
		BeforeEach(func() {
			quotaUsage = models.QuotaUsage{}
			quotaUsage.Name = "Organization-Quota"
			quotaUsage.MemoryLimit = 10240
			quotaUsage.RoutesLimit = 1000
			quotaUsage.ServicesLimit = 100
			quotaUsage.OrgUsage.Routes = 6
			quotaUsage.OrgUsage.Services = 6
			quotaUsage.OrgUsage.Memory = 1280

			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			orgRepo.FindByNameReturns(org, nil)
		})

		It("lists quota usage", func() {
			orgRepo.GetOrganizationQuotaUsageReturns(quotaUsage, nil)

			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting quota usage info for org my-org as my-user"},
				[]string{"Routes", "6/1000"},
				[]string{"Services", "6/100"},
				[]string{"Memory", "1.2G/10G"},
				[]string{"OK"},
			))
		})

		It("lists quota usage when service limit is -1", func() {
			quotaUsage.ServicesLimit = int(-1)
			orgRepo.GetOrganizationQuotaUsageReturns(quotaUsage, nil)

			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting quota usage info for org my-org as my-user"},
				[]string{"Routes", "6/1000"},
				[]string{"Services", "6/unlimited"},
				[]string{"Memory", "1.2G/10G"},
				[]string{"OK"},
			))
		})
	})

	It("tells the user when org not found", func() {
		orgRepo.FindByNameReturns(models.Organization{}, errors.New("Organization dummy-org-name not found"))
		runCommand("dummy-Org-Name")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting quota usage info for org dummy-Org-Name as my-user"},
			[]string{"Organization dummy-org-name not found"},
		))
	})

	It("tell user error return from endpoint", func() {
		orgRepo.GetOrganizationQuotaUsageReturns(models.QuotaUsage{}, errors.New("User does not have access"))
		runCommand("my-org")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting quota usage info for org my-org as my-user"},
			[]string{"User does not have access"},
		))
	})
})
