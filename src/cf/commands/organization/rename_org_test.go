package organization_test

import (
	"cf/commands/organization"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("rename-org command", func() {
	var reqFactory *testreq.FakeReqFactory
	var orgRepo *testapi.FakeOrgRepository

	BeforeEach(func() {
		reqFactory = &testreq.FakeReqFactory{}
		orgRepo = &testapi.FakeOrgRepository{}
	})

	It("fails with usage when given less than two args", func() {
		ui := callRenameOrg([]string{}, reqFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRenameOrg([]string{"foo"}, reqFactory, orgRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in and given an org to rename", func() {
		var ui *testterm.FakeUI

		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			reqFactory.Organization = org
			reqFactory.LoginSuccess = true
			ui = callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)
		})

		It("pass requirements", func() {
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("renames an organization", func() {
			Expect(reqFactory.OrganizationName).To(Equal("my-org"))
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Renaming org", "my-org", "my-new-org", "my-user"},
				{"OK"},
			})

			Expect(orgRepo.RenameOrganizationGuid).To(Equal("my-org-guid"))
			Expect(orgRepo.RenameNewName).To(Equal("my-new-org"))
		})
	})
})

func callRenameOrg(args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-org", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := organization.NewRenameOrg(ui, configRepo, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
