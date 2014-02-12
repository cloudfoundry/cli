package application_test

import (
	. "cf/commands/application"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteConfirmingWithY", func() {
		ui, _, appRepo := deleteApp("y", []string{"app-to-delete"})

		Expect(appRepo.ReadName).To(Equal("app-to-delete"))
		Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))
		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteConfirmingWithYes", func() {

		ui, _, appRepo := deleteApp("Yes", []string{"app-to-delete"})

		Expect(appRepo.ReadName).To(Equal("app-to-delete"))
		Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "app-to-delete"},
		})
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
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

		Expect(appRepo.ReadName).To(Equal("app-to-delete"))
		Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))
		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
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

		Expect(appRepo.ReadName).To(Equal("app-to-delete"))
		Expect(appRepo.DeletedAppGuid).To(Equal(""))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting", "app-to-delete"},
			{"OK"},
			{"app-to-delete", "does not exist"},
		})
	})
	It("TestDeleteCommandFailsWithUsage", func() {

		ui, _, _ := deleteApp("Yes", []string{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui, _, _ = deleteApp("Yes", []string{"app-to-delete"})
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
})

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
