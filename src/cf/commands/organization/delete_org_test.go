package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func deleteOrg(t mr.TestingT, confirmation string, args []string, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	reqFactory := &testreq.FakeReqFactory{}
	configRepo := &testconfig.FakeConfigRepository{}
	configRepo.EnsureInitialized()

	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := NewDeleteOrg(ui, config, orgRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteOrgConfirmingWithY", func() {
			org := cf.Organization{}
			org.Name = "org-to-delete"
			org.Guid = "org-to-delete-guid"
			orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}

			ui := deleteOrg(mr.T(), "y", []string{org.Name}, orgRepo)

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete"},
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
			})
			assert.Equal(mr.T(), orgRepo.FindByNameName, "org-to-delete")
			assert.Equal(mr.T(), orgRepo.DeletedOrganizationGuid, "org-to-delete-guid")
		})
		It("TestDeleteOrgConfirmingWithYes", func() {

			org := cf.Organization{}
			org.Name = "org-to-delete"
			org.Guid = "org-to-delete-guid"
			orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}

			ui := deleteOrg(mr.T(), "Yes", []string{"org-to-delete"}, orgRepo)

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "org-to-delete"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting org", "org-to-delete", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), orgRepo.FindByNameName, "org-to-delete")
			assert.Equal(mr.T(), orgRepo.DeletedOrganizationGuid, "org-to-delete-guid")
		})
		It("TestDeleteTargetedOrganizationClearsConfig", func() {

			configRepo := &testconfig.FakeConfigRepository{}
			config, _ := configRepo.Get()

			organizationFields := cf.OrganizationFields{}
			organizationFields.Name = "org-to-delete"
			organizationFields.Guid = "org-to-delete-guid"
			config.OrganizationFields = organizationFields

			spaceFields := cf.SpaceFields{}
			spaceFields.Name = "space-to-delete"
			config.SpaceFields = spaceFields
			configRepo.Save()

			org := cf.Organization{}
			org.OrganizationFields = organizationFields
			orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}
			deleteOrg(mr.T(), "Yes", []string{"org-to-delete"}, orgRepo)

			updatedConfig, err := configRepo.Get()
			assert.NoError(mr.T(), err)

			assert.Equal(mr.T(), updatedConfig.OrganizationFields, cf.OrganizationFields{})
			assert.Equal(mr.T(), updatedConfig.SpaceFields, cf.SpaceFields{})
		})
		It("TestDeleteUntargetedOrganizationDoesNotClearConfig", func() {

			org := cf.Organization{}
			org.Name = "org-to-delete"
			org.Guid = "org-to-delete-guid"
			orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}

			configRepo := &testconfig.FakeConfigRepository{}
			config, _ := configRepo.Get()
			otherOrgFields := cf.OrganizationFields{}
			otherOrgFields.Guid = "some-other-org-guid"
			otherOrgFields.Name = "some-other-org"
			config.OrganizationFields = otherOrgFields

			spaceFields := cf.SpaceFields{}
			spaceFields.Name = "some-other-space"
			config.SpaceFields = spaceFields
			configRepo.Save()

			deleteOrg(mr.T(), "Yes", []string{"org-to-delete"}, orgRepo)

			updatedConfig, err := configRepo.Get()
			assert.NoError(mr.T(), err)

			assert.Equal(mr.T(), updatedConfig.OrganizationFields.Name, "some-other-org")
			assert.Equal(mr.T(), updatedConfig.SpaceFields.Name, "some-other-space")
		})
		It("TestDeleteOrgWithForceOption", func() {

			org := cf.Organization{}
			org.Name = "org-to-delete"
			org.Guid = "org-to-delete-guid"
			orgRepo := &testapi.FakeOrgRepository{Organizations: []cf.Organization{org}}

			ui := deleteOrg(mr.T(), "Yes", []string{"-f", "org-to-delete"}, orgRepo)

			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
			})
			assert.Equal(mr.T(), orgRepo.FindByNameName, "org-to-delete")
			assert.Equal(mr.T(), orgRepo.DeletedOrganizationGuid, "org-to-delete-guid")
		})
		It("TestDeleteOrgCommandFailsWithUsage", func() {

			orgRepo := &testapi.FakeOrgRepository{}
			ui := deleteOrg(mr.T(), "Yes", []string{}, orgRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = deleteOrg(mr.T(), "Yes", []string{"org-to-delete"}, orgRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestDeleteOrgWhenOrgDoesNotExist", func() {

			orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
			ui := deleteOrg(mr.T(), "y", []string{"org-to-delete"}, orgRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 3)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "org-to-delete"},
				{"OK"},
				{"org-to-delete", "does not exist."},
			})

			assert.Equal(mr.T(), orgRepo.FindByNameName, "org-to-delete")
		})
	})
}
