package organization_test

import (
	. "cf/commands/organization"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	var config configuration.ReadWriter
	var ui *testterm.FakeUI
	var reqFactory *testreq.FakeReqFactory
	var orgRepo *testapi.FakeOrgRepository

	BeforeEach(func() {
		reqFactory = &testreq.FakeReqFactory{}

		ui = &testterm.FakeUI{}

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

	It("TestDeleteOrgConfirmingWithY", func() {
		ui.Inputs = []string{"y"}
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), reqFactory)

		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete"},
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting", "org-to-delete"},
			{"OK"},
		})
		Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
		Expect(orgRepo.DeletedOrganizationGuid).To(Equal("org-to-delete-guid"))
	})

	It("TestDeleteOrgConfirmingWithYes", func() {
		ui.Inputs = []string{"Yes"}

		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), reqFactory)

		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete", "org-to-delete"},
		})
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), reqFactory)

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
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), reqFactory)

		Expect(config.OrganizationFields().Name).To(Equal("some-other-org"))
		Expect(config.SpaceFields().Name).To(Equal("some-other-space"))
	})

	It("TestDeleteOrgWithForceOption", func() {
		ui.Inputs = []string{"Yes"}
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"-f", "org-to-delete"}), reqFactory)

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting", "org-to-delete"},
			{"OK"},
		})
		Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
		Expect(orgRepo.DeletedOrganizationGuid).To(Equal("org-to-delete-guid"))
	})

	It("FailsWithUsage when 1st argument is omitted", func() {
		ui.Inputs = []string{"Yes"}
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{}), reqFactory)
		assert.True(mr.T(), ui.FailedWithUsage)
	})

	It("TestDeleteOrgWhenOrgDoesNotExist", func() {
		orgRepo.FindByNameNotFound = true

		ui.Inputs = []string{"y"}
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-org", []string{"org-to-delete"}), reqFactory)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting", "org-to-delete"},
			{"OK"},
			{"org-to-delete", "does not exist."},
		})

		Expect(orgRepo.FindByNameName).To(Equal("org-to-delete"))
	})
})
