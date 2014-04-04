/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package organization_test

import (
	. "cf/commands/organization"
	"cf/configuration"
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

var _ = Describe("delete-org command", func() {
	var (
		config              configuration.ReadWriter
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *testapi.FakeOrgRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		spaceFields := models.SpaceFields{}
		spaceFields.Name = "my-space"

		orgFields := models.OrganizationFields{}
		orgFields.Name = "my-org"

		token := configuration.TokenInfo{Username: "my-user"}
		config = testconfig.NewRepositoryWithAccessToken(token)
		config.SetSpaceFields(spaceFields)
		config.SetOrganizationFields(orgFields)

		org := models.Organization{}
		org.Name = "org-to-delete"
		org.Guid = "org-to-delete-guid"
		orgRepo = &testapi.FakeOrgRepository{Organizations: []models.Organization{org}}
	})

	It("fails requirements when not logged in", func() {
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"some-org-name"}), requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("TestDeleteOrgConfirmingWithY", func() {
			ui.Inputs = []string{"y"}
			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), requirementsFactory)

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
			})
			Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
			Expect(orgRepo.DeletedOrganizationGuid).To(Equal("org-to-delete-guid"))
		})

		It("TestDeleteOrgConfirmingWithYes", func() {
			ui.Inputs = []string{"Yes"}

			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), requirementsFactory)

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete", "org-to-delete"},
			})
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting org", "org-to-delete", "my-user"},
				{"OK"},
			})

			Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
			Expect(orgRepo.DeletedOrganizationGuid).To(Equal("org-to-delete-guid"))
		})

		It("TestDeleteTargetedOrganizationClearsConfig", func() {
			config.SetOrganizationFields(orgRepo.Organizations[0].OrganizationFields)

			spaceFields := models.SpaceFields{}
			spaceFields.Name = "space-to-delete"
			config.SetSpaceFields(spaceFields)

			ui.Inputs = []string{"Yes"}

			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), requirementsFactory)

			Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
			Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
		})

		It("TestDeleteUntargetedOrganizationDoesNotClearConfig", func() {
			otherOrgFields := models.OrganizationFields{}
			otherOrgFields.Guid = "some-other-org-guid"
			otherOrgFields.Name = "some-other-org"
			config.SetOrganizationFields(otherOrgFields)

			spaceFields := models.SpaceFields{}
			spaceFields.Name = "some-other-space"
			config.SetSpaceFields(spaceFields)

			ui.Inputs = []string{"Yes"}

			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), requirementsFactory)

			Expect(config.OrganizationFields().Name).To(Equal("some-other-org"))
			Expect(config.SpaceFields().Name).To(Equal("some-other-space"))
		})

		It("TestDeleteOrgWithForceOption", func() {
			ui.Inputs = []string{"Yes"}
			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"-f", "org-to-delete"}), requirementsFactory)

			Expect(len(ui.Prompts)).To(Equal(0))
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
			})
			Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
			Expect(orgRepo.DeletedOrganizationGuid).To(Equal("org-to-delete-guid"))
		})

		It("FailsWithUsage when 1st argument is omitted", func() {
			ui.Inputs = []string{"Yes"}
			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{}), requirementsFactory)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("TestDeleteOrgWhenOrgDoesNotExist", func() {
			orgRepo.FindByNameNotFound = true

			ui.Inputs = []string{"y"}
			cmd := NewDeleteOrg(ui, config, orgRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), requirementsFactory)

			Expect(len(ui.Outputs)).To(Equal(3))
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
				{"org-to-delete", "does not exist."},
			})

			Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
		})
	})
})
