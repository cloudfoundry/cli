package organization_test

import (
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/configuration"
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

var _ = Describe("rename-org command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *testapi.FakeOrgRepository
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{}
		orgRepo = &testapi.FakeOrgRepository{}
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	var callRenameOrg = func(args []string) {
		cmd := organization.NewRenameOrg(ui, configRepo, orgRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails with usage when given less than two args", func() {
		callRenameOrg([]string{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		callRenameOrg([]string{"foo"})
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		callRenameOrg([]string{"my-org", "my-new-org"})
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in and given an org to rename", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "the-old-org-name"
			org.Guid = "the-old-org-guid"
			requirementsFactory.Organization = org
			requirementsFactory.LoginSuccess = true
		})

		It("passes requirements", func() {
			callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("renames an organization", func() {
			targetedOrgName := configRepo.OrganizationFields().Name
			callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming org", "the-old-org-name", "the-new-org-name", "my-user"},
				[]string{"OK"},
			))

			Expect(requirementsFactory.OrganizationName).To(Equal("the-old-org-name"))
			Expect(orgRepo.RenameOrganizationGuid).To(Equal("the-old-org-guid"))
			Expect(orgRepo.RenameNewName).To(Equal("the-new-org-name"))
			Expect(configRepo.OrganizationFields().Name).To(Equal(targetedOrgName))
		})

		Describe("when the organization is currently targeted", func() {
			It("updates the name of the org in the config", func() {
				configRepo.SetOrganizationFields(models.OrganizationFields{
					Guid: "the-old-org-guid",
					Name: "the-old-org-name",
				})
				callRenameOrg([]string{"the-old-org-name", "the-new-org-name"})
				Expect(configRepo.OrganizationFields().Name).To(Equal("the-new-org-name"))
			})
		})
	})
})
