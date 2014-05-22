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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/application"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestFilesRequirements", func() {
		args := []string{"my-app", "/foo"}
		appFilesRepo := &testapi.FakeAppFilesRepo{}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: models.Application{}}
		callFiles(args, requirementsFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: models.Application{}}
		callFiles(args, requirementsFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		callFiles(args, requirementsFactory, appFilesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestFilesFailsWithUsage", func() {

		appFilesRepo := &testapi.FakeAppFilesRepo{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		ui := callFiles([]string{}, requirementsFactory, appFilesRepo)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestListingDirectoryEntries", func() {

		app := models.Application{}
		app.Name = "my-found-app"
		app.Guid = "my-app-guid"

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
		appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

		ui := callFiles([]string{"my-app", "/foo"}, requirementsFactory, appFilesRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"file 1"},
			[]string{"file 2"},
		))

		Expect(appFilesRepo.AppGuid).To(Equal("my-app-guid"))
		Expect(appFilesRepo.Path).To(Equal("/foo"))
	})

	It("TestListingFilesWithTemplateTokens", func() {

		app := models.Application{}
		app.Name = "my-found-app"
		app.Guid = "my-app-guid"

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
		appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "%s %d %i"}

		ui := callFiles([]string{"my-app", "/foo"}, requirementsFactory, appFilesRepo)

		Expect(ui.Outputs).To(ContainSubstrings([]string{"%s %d %i"}))
	})
})

func callFiles(args []string, requirementsFactory *testreq.FakeReqFactory, appFilesRepo *testapi.FakeAppFilesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewFiles(ui, configRepo, appFilesRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)

	return
}
