package organization_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"

	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	test_quota "github.com/cloudfoundry/cli/cf/api/quotas/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/organization"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-org command", func() {
	var (
		config              core_config.ReadWriter
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *test_org.FakeOrganizationRepository
		quotaRepo           *test_quota.FakeQuotaRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &test_org.FakeOrganizationRepository{}
		quotaRepo = &test_quota.FakeQuotaRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewCreateOrg(ui, config, orgRepo, quotaRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			Expect(runCommand("my-org")).To(BeFalse())
		})
	})

	Context("when logged in and provided the name of an org to create", func() {
		BeforeEach(func() {
			orgRepo.CreateReturns(nil)
			requirementsFactory.LoginSuccess = true
		})

		It("creates an org", func() {
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org", "my-user"},
				[]string{"OK"},
			))
			Expect(orgRepo.CreateArgsForCall(0).Name).To(Equal("my-org"))
		})

		It("fails and warns the user when the org already exists", func() {
			err := errors.NewHttpError(400, errors.ORG_EXISTS, "org already exists")
			orgRepo.CreateReturns(err)
			runCommand("my-org")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating org", "my-org"},
				[]string{"OK"},
				[]string{"my-org", "already exists"},
			))
		})

		Context("when allowing a non-defualt quota", func() {
			var (
				quota models.QuotaFields
			)

			BeforeEach(func() {
				quota = models.QuotaFields{
					Name: "non-default-quota",
					Guid: "non-default-quota-guid",
				}
			})

			It("creates an org with a non-default quota", func() {
				quotaRepo.FindByNameReturns(quota, nil)
				runCommand("-q", "non-default-quota", "my-org")

				Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("non-default-quota"))
				Expect(orgRepo.CreateArgsForCall(0).QuotaDefinition.Guid).To(Equal("non-default-quota-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"OK"},
				))
			})

			It("fails and warns the user when the quota cannot be found", func() {
				quotaRepo.FindByNameReturns(models.QuotaFields{}, errors.New("Could not find quota"))
				runCommand("-q", "non-default-quota", "my-org")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating org", "my-org"},
					[]string{"Could not find quota"},
				))
			})
		})
	})
})
