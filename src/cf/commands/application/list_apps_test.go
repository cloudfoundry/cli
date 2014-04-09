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

var _ = Describe("list-apps command", func() {
	It("TestApps", func() {
		domain := models.DomainFields{}
		domain.Name = "cfapps.io"
		domain2 := models.DomainFields{}
		domain2.Name = "example.com"

		route1 := models.RouteSummary{}
		route1.Host = "app1"
		route1.Domain = domain

		route2 := models.RouteSummary{}
		route2.Host = "app1"
		route2.Domain = domain2

		app1Routes := []models.RouteSummary{route1, route2}

		domain3 := models.DomainFields{}
		domain3.Name = "cfapps.io"

		route3 := models.RouteSummary{}
		route3.Host = "app2"
		route3.Domain = domain3

		app2Routes := []models.RouteSummary{route3}

		app := models.Application{}
		app.Name = "Application-1"
		app.State = "started"
		app.RunningInstances = 1
		app.InstanceCount = 1
		app.Memory = 512
		app.DiskQuota = 1024
		app.Routes = app1Routes

		app2 := models.Application{}
		app2.Name = "Application-2"
		app2.State = "started"
		app2.RunningInstances = 1
		app2.InstanceCount = 2
		app2.Memory = 256
		app2.DiskQuota = 1024
		app2.Routes = app2Routes

		apps := []models.Application{app, app2}

		appSummaryRepo := &testapi.FakeAppSummaryRepo{
			GetSummariesInCurrentSpaceApps: apps,
		}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		ui := callApps(appSummaryRepo, requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting apps in", "my-org", "my-space", "my-user"},
			{"OK"},
			{"Application-1", "started", "1/1", "512M", "1G", "app1.cfapps.io", "app1.example.com"},
			{"Application-2", "started", "1/2", "256M", "1G", "app2.cfapps.io"},
		})
	})
	It("TestAppsEmptyList", func() {

		appSummaryRepo := &testapi.FakeAppSummaryRepo{
			GetSummariesInCurrentSpaceApps: []models.Application{},
		}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		ui := callApps(appSummaryRepo, requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting apps in", "my-org", "my-space", "my-user"},
			{"OK"},
			{"No apps found"},
		})
	})
	It("TestAppsRequiresLogin", func() {

		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

		callApps(appSummaryRepo, requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestAppsRequiresASelectedSpaceAndOrg", func() {

		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

		callApps(appSummaryRepo, requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
})

func callApps(appSummaryRepo *testapi.FakeAppSummaryRepo, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	configRepo := testconfig.NewRepositoryWithDefaults()
	ctxt := testcmd.NewContext("apps", []string{})
	cmd := NewListApps(ui, configRepo, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)

	return
}
