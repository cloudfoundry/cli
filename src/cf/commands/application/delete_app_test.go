package application_test

import (
	. "cf/commands/application"
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

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteConfirmingWithY", func() {
			ui, _, appRepo := deleteApp("y", []string{"app-to-delete"})

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

			ui, _, appRepo := deleteApp("Yes", []string{"app-to-delete"})

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

			cmd := NewDeleteApp(ui, testconfig.NewRepository(), appRepo)
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

			cmd := NewDeleteApp(ui, testconfig.NewRepository(), appRepo)
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

			ui, _, _ := deleteApp("Yes", []string{})
			assert.True(mr.T(), ui.FailedWithUsage)

			ui, _, _ = deleteApp("Yes", []string{"app-to-delete"})
			assert.False(mr.T(), ui.FailedWithUsage)
		})
	})
}

func deleteApp(confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) {

	app := models.Application{}
	app.Name = "app-to-delete"
	app.Guid = "app-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{}
	appRepo = &testapi.FakeApplicationRepository{ReadApp: app}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	configRepo := testconfig.NewRepositoryWithDefaults()

	ctxt := testcmd.NewContext("delete", args)
	cmd := NewDeleteApp(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
