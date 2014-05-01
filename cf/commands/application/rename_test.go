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
	It("TestRenameAppFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		appRepo := &testapi.FakeApplicationRepository{}

		ui := callRename([]string{}, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRename([]string{"foo"}, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})
	It("TestRenameRequirements", func() {

		appRepo := &testapi.FakeApplicationRepository{}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callRename([]string{"my-app", "my-new-app"}, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestRenameRun", func() {

		appRepo := &testapi.FakeApplicationRepository{}
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
		ui := callRename([]string{"my-app", "my-new-app"}, requirementsFactory, appRepo)

		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.Name).To(Equal("my-new-app"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
	})
})

func callRename(args []string, requirementsFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewRenameApp(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}
