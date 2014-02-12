package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestFilesRequirements", func() {
		args := []string{"my-app", "/foo"}
		appFilesRepo := &testapi.FakeAppFilesRepo{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: models.Application{}}
		callFiles(args, reqFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: models.Application{}}
		callFiles(args, reqFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		callFiles(args, reqFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestFilesFailsWithUsage", func() {

		appFilesRepo := &testapi.FakeAppFilesRepo{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		ui := callFiles([]string{}, reqFactory, appFilesRepo)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestListingDirectoryEntries", func() {

		app := models.Application{}
		app.Name = "my-found-app"
		app.Guid = "my-app-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
		appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

		ui := callFiles([]string{"my-app", "/foo"}, reqFactory, appFilesRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
			{"OK"},
			{"file 1"},
			{"file 2"},
		})

		Expect(appFilesRepo.AppGuid).To(Equal("my-app-guid"))
		Expect(appFilesRepo.Path).To(Equal("/foo"))
	})
	It("TestListingFilesWithTemplateTokens", func() {

		app := models.Application{}
		app.Name = "my-found-app"
		app.Guid = "my-app-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
		appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "%s %d %i"}

		ui := callFiles([]string{"my-app", "/foo"}, reqFactory, appFilesRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"%s %d %i"},
		})
	})
})

func callFiles(args []string, reqFactory *testreq.FakeReqFactory, appFilesRepo *testapi.FakeAppFilesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("files", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewFiles(ui, configRepo, appFilesRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
