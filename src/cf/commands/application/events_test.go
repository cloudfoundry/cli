package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

func TestEventsRequirements(t *testing.T) {
	reqFactory, eventsRepo := getEventsDependencies()

	callEvents(t, []string{"my-app"}, reqFactory, eventsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestEventsFailsWithUsage(t *testing.T) {
	reqFactory, eventsRepo := getEventsDependencies()
	ui := callEvents(t, []string{}, reqFactory, eventsRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEventsSuccess(t *testing.T) {
	timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
	assert.NoError(t, err)

	reqFactory, eventsRepo := getEventsDependencies()
	app := cf.Application{}
	app.Name = "my-app"
	reqFactory.Application = app

	eventsRepo.Events = []cf.EventFields{
		{
			InstanceIndex:   98,
			Timestamp:       timestamp,
			ExitDescription: "app instance exited",
			ExitStatus:      78,
		},
		{
			InstanceIndex:   99,
			Timestamp:       timestamp,
			ExitDescription: "app instance was stopped",
			ExitStatus:      77,
		},
	}

	ui := callEvents(t, []string{"my-app"}, reqFactory, eventsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting events for app", "my-app", "my-org", "my-space", "my-user"},
		{"time", "instance", "description", "exit status"},
		{timestamp.Local().Format(TIMESTAMP_FORMAT), "98", "app instance exited", "78"},
		{timestamp.Local().Format(TIMESTAMP_FORMAT), "99", "app instance was stopped", "77"},
	})
}

func TestEventsWhenNoEventsAvailable(t *testing.T) {
	reqFactory, eventsRepo := getEventsDependencies()
	app := cf.Application{}
	app.Name = "my-app"
	reqFactory.Application = app

	ui := callEvents(t, []string{"my-app"}, reqFactory, eventsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"events", "my-app"},
		{"No events", "my-app"},
	})
}

func getEventsDependencies() (reqFactory *testreq.FakeReqFactory, eventsRepo *testapi.FakeAppEventsRepo) {
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	eventsRepo = &testapi.FakeAppEventsRepo{}
	return
}

func callEvents(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, eventsRepo *testapi.FakeAppEventsRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewEvents(ui, config, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
