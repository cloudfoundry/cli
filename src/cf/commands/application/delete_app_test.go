package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/models"
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

func deleteApp(t mr.TestingT, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) {

	app := models.Application{}
	app.Name = "app-to-delete"
	app.Guid = "app-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{}
	appRepo = &testapi.FakeApplicationRepository{ReadApp: app}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	org := models.OrganizationFields{}
	org.Name = "my-org"
	space := models.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	ctxt := testcmd.NewContext("delete", args)
	cmd := NewDeleteApp(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteConfirmingWithY", func() {
			ui, _, appRepo := deleteApp(mr.T(), "y", []string{"app-to-delete"})

			assert.Equal(mr.T(), appRepo.ReadName, "app-to-delete")
			assert.Equal(mr.T(), appRepo.DeletedAppGuid, "app-to-delete-guid")
			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
				{"OK"},
			})
		})
		It("TestDeleteConfirmingWithYes", func() {

			ui, _, appRepo := deleteApp(mr.T(), "Yes", []string{"app-to-delete"})

			assert.Equal(mr.T(), appRepo.ReadName, "app-to-delete")
			assert.Equal(mr.T(), appRepo.DeletedAppGuid, "app-to-delete-guid")

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "app-to-delete"},
			})
			assert.Equal(mr.T(), len(ui.Outputs), 2)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
				{"OK"},
			})
		})
		It("TestDeleteWithForceOption", func() {

			app := models.Application{}
			app.Name = "app-to-delete"
			app.Guid = "app-to-delete-guid"

			reqFactory := &testreq.FakeReqFactory{}
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

			ui := &testterm.FakeUI{}
			ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

			cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), appRepo.ReadName, "app-to-delete")
			assert.Equal(mr.T(), appRepo.DeletedAppGuid, "app-to-delete-guid")
			assert.Equal(mr.T(), len(ui.Prompts), 0)
			assert.Equal(mr.T(), len(ui.Outputs), 2)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "app-to-delete"},
				{"OK"},
			})
		})
		It("TestDeleteAppThatDoesNotExist", func() {

			reqFactory := &testreq.FakeReqFactory{}
			appRepo := &testapi.FakeApplicationRepository{ReadNotFound: true}

			ui := &testterm.FakeUI{}
			ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

			cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), appRepo.ReadName, "app-to-delete")
			assert.Equal(mr.T(), appRepo.DeletedAppGuid, "")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "app-to-delete"},
				{"OK"},
				{"app-to-delete", "does not exist"},
			})
		})
		It("TestDeleteCommandFailsWithUsage", func() {

			ui, _, _ := deleteApp(mr.T(), "Yes", []string{})
			assert.True(mr.T(), ui.FailedWithUsage)

			ui, _, _ = deleteApp(mr.T(), "Yes", []string{"app-to-delete"})
			assert.False(mr.T(), ui.FailedWithUsage)
		})
	})
}
