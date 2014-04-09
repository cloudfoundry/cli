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
	"cf/errors"
	"cf/formatters"
	"cf/models"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	"time"
)

var _ = Describe("Show App Command", func() {
	It("requires the user to be logged in and have a targeted space", func() {
		args := []string{"my-app", "/foo"}
		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: models.Application{}}
		callApp(args, requirementsFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: models.Application{}}
		callApp(args, requirementsFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		callApp(args, requirementsFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
	})

	It("requires an app name", func() {
		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		ui := callApp([]string{}, requirementsFactory, appSummaryRepo, appInstancesRepo)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("displays a summary of the app", func() {
		reqApp := models.Application{}
		reqApp.Name = "my-app"
		reqApp.Guid = "my-app-guid"

		route1 := models.RouteSummary{}
		route1.Host = "my-app"

		domain := models.DomainFields{}
		domain.Name = "example.com"
		route1.Domain = domain

		route2 := models.RouteSummary{}
		route2.Host = "foo"
		domain2 := models.DomainFields{}
		domain2.Name = "example.com"
		route2.Domain = domain2

		application := models.Application{}
		application.State = "started"
		application.InstanceCount = 2
		application.RunningInstances = 2
		application.Memory = 256
		application.Routes = []models.RouteSummary{route1, route2}

		time1, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012")
		Expect(err).NotTo(HaveOccurred())

		time2, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012")
		Expect(err).NotTo(HaveOccurred())

		appInstance := models.AppInstanceFields{}
		appInstance.State = models.InstanceRunning
		appInstance.Since = time1
		appInstance.CpuUsage = 1.0
		appInstance.DiskQuota = 1 * formatters.GIGABYTE
		appInstance.DiskUsage = 32 * formatters.MEGABYTE
		appInstance.MemQuota = 64 * formatters.MEGABYTE
		appInstance.MemUsage = 13 * formatters.BYTE

		appInstance2 := models.AppInstanceFields{}
		appInstance2.State = models.InstanceDown
		appInstance2.Since = time2

		instances := []models.AppInstanceFields{appInstance, appInstance2}

		appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: application}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{GetInstancesResponses: [][]models.AppInstanceFields{instances}}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
		ui := callApp([]string{"my-app"}, requirementsFactory, appSummaryRepo, appInstancesRepo)

		Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Showing health and status", "my-app"},
			{"state", "started"},
			{"instances", "2/2"},
			{"usage", "256M x 2 instances"},
			{"urls", "my-app.example.com", "foo.example.com"},
			{"#0", "running", "2012-01-02 03:04:05 PM", "100.0%", "13 of 64M", "32M of 1G"},
			{"#1", "down", "2012-04-01 03:04:05 PM", "0%", "0 of 0", "0 of 0"},
		})
	})

	It("TestDisplayingStoppedAppSummary", func() {
		testDisplayingAppSummaryWithErrorCode(errors.APP_STOPPED)
	})

	It("TestDisplayingNotStagedAppSummary", func() {
		testDisplayingAppSummaryWithErrorCode(errors.APP_NOT_STAGED)
	})
})

func testDisplayingAppSummaryWithErrorCode(errorCode string) {
	reqApp := models.Application{}
	reqApp.Name = "my-app"
	reqApp.Guid = "my-app-guid"

	domain3 := models.DomainFields{}
	domain3.Name = "example.com"
	domain4 := models.DomainFields{}
	domain4.Name = "example.com"

	route1 := models.RouteSummary{}
	route1.Host = "my-app"
	route1.Domain = domain3

	route2 := models.RouteSummary{}
	route2.Host = "foo"
	route2.Domain = domain4

	routes := []models.RouteSummary{
		route1,
		route2,
	}

	app := models.ApplicationFields{}
	app.State = "stopped"
	app.InstanceCount = 2
	app.RunningInstances = 0
	app.Memory = 256

	application := models.Application{}
	application.ApplicationFields = app
	application.Routes = routes

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: application, GetSummaryErrorCode: errorCode}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp([]string{"my-app"}, requirementsFactory, appSummaryRepo, appInstancesRepo)

	Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))
	Expect(appInstancesRepo.GetInstancesAppGuid).To(Equal("my-app-guid"))

	testassert.SliceContains(ui.Outputs, testassert.Lines{
		{"Showing health and status", "my-app", "my-org", "my-space", "my-user"},
		{"state", "stopped"},
		{"instances", "0/2"},
		{"usage", "256M x 2 instances"},
		{"urls", "my-app.example.com, foo.example.com"},
		{"no running instances"},
	})
}

func callApp(args []string, requirementsFactory *testreq.FakeReqFactory, appSummaryRepo *testapi.FakeAppSummaryRepo, appInstancesRepo *testapi.FakeAppInstancesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("app", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewShowApp(ui, configRepo, appSummaryRepo, appInstancesRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)

	return
}
