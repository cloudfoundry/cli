package application_test

import (
	"cf"
	. "cf/commands/application"
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
	mr "github.com/tjarratt/mr_t"
	"time"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestAppRequirements", func() {
		args := []string{"my-app", "/foo"}
		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: models.Application{}}
		callApp(args, reqFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: models.Application{}}
		callApp(args, reqFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		callApp(args, reqFactory, appSummaryRepo, appInstancesRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestAppFailsWithUsage", func() {

		appSummaryRepo := &testapi.FakeAppSummaryRepo{}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: models.Application{}}
		ui := callApp([]string{}, reqFactory, appSummaryRepo, appInstancesRepo)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestDisplayingAppSummary", func() {

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

		appSummary := models.AppSummary{}
		appSummary.State = "started"
		appSummary.InstanceCount = 2
		appSummary.RunningInstances = 2
		appSummary.Memory = 256
		appSummary.RouteSummaries = []models.RouteSummary{route1, route2}

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

		appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{GetInstancesResponses: [][]models.AppInstanceFields{instances}}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
		ui := callApp([]string{"my-app"}, reqFactory, appSummaryRepo, appInstancesRepo)

		Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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

		testDisplayingAppSummaryWithErrorCode(mr.T(), cf.APP_STOPPED)
	})
	It("TestDisplayingNotStagedAppSummary", func() {

		testDisplayingAppSummaryWithErrorCode(mr.T(), cf.APP_NOT_STAGED)
	})
})

func testDisplayingAppSummaryWithErrorCode(t mr.TestingT, errorCode string) {
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

	appSummary := models.AppSummary{}
	appSummary.ApplicationFields = app
	appSummary.RouteSummaries = routes

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary, GetSummaryErrorCode: errorCode}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp([]string{"my-app"}, reqFactory, appSummaryRepo, appInstancesRepo)

	Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))
	Expect(appInstancesRepo.GetInstancesAppGuid).To(Equal("my-app-guid"))

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Showing health and status", "my-app", "my-org", "my-space", "my-user"},
		{"state", "stopped"},
		{"instances", "0/2"},
		{"usage", "256M x 2 instances"},
		{"urls", "my-app.example.com, foo.example.com"},
		{"no running instances"},
	})
}

func callApp(args []string, reqFactory *testreq.FakeReqFactory, appSummaryRepo *testapi.FakeAppSummaryRepo, appInstancesRepo *testapi.FakeAppInstancesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("app", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewShowApp(ui, configRepo, appSummaryRepo, appInstancesRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
