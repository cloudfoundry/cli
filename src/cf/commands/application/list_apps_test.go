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

func callApps(t mr.TestingT, appSummaryRepo *testapi.FakeAppSummaryRepo, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := models.SpaceFields{}
	space.Name = "development"
	org := models.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	ctxt := testcmd.NewContext("apps", []string{})
	cmd := NewListApps(ui, config, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
func init() {
	Describe("Testing with ginkgo", func() {
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

			app := models.AppSummary{}
			app.Name = "Application-1"
			app.State = "started"
			app.RunningInstances = 1
			app.InstanceCount = 1
			app.Memory = 512
			app.DiskQuota = 1024
			app.RouteSummaries = app1Routes

			app2 := models.AppSummary{}
			app2.Name = "Application-2"
			app2.State = "started"
			app2.RunningInstances = 1
			app2.InstanceCount = 2
			app2.Memory = 256
			app2.DiskQuota = 1024
			app2.RouteSummaries = app2Routes

			apps := []models.AppSummary{app, app2}

			appSummaryRepo := &testapi.FakeAppSummaryRepo{
				GetSummariesInCurrentSpaceApps: apps,
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

			ui := callApps(mr.T(), appSummaryRepo, reqFactory)

			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting apps in", "my-org", "development", "my-user"},
				{"OK"},
				{"Application-1", "started", "1/1", "512M", "1G", "app1.cfapps.io", "app1.example.com"},
				{"Application-2", "started", "1/2", "256M", "1G", "app2.cfapps.io"},
			})
		})
		It("TestAppsEmptyList", func() {

			appSummaryRepo := &testapi.FakeAppSummaryRepo{
				GetSummariesInCurrentSpaceApps: []models.AppSummary{},
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

			ui := callApps(mr.T(), appSummaryRepo, reqFactory)

			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting apps in", "my-org", "development", "my-user"},
				{"OK"},
				{"No apps found"},
			})
		})
		It("TestAppsRequiresLogin", func() {

			appSummaryRepo := &testapi.FakeAppSummaryRepo{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

			callApps(mr.T(), appSummaryRepo, reqFactory)

			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestAppsRequiresASelectedSpaceAndOrg", func() {

			appSummaryRepo := &testapi.FakeAppSummaryRepo{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

			callApps(mr.T(), appSummaryRepo, reqFactory)

			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
	})
}
