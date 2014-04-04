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

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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
