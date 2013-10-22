package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
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
	reqFactory.Application = cf.Application{Name: "my-app", Guid: "my-app-guid"}

	eventsRepo.Events = []cf.Event{
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

	assert.Contains(t, ui.Outputs[0], "Getting events for app")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Showing all 2 events")
	assert.Contains(t, ui.Outputs[4], "time")
	assert.Contains(t, ui.Outputs[4], "instance")
	assert.Contains(t, ui.Outputs[4], "description")
	assert.Contains(t, ui.Outputs[4], "exit status")
	assert.Contains(t, ui.Outputs[5], timestamp.Local().Format(TIMESTAMP_FORMAT))
	assert.Contains(t, ui.Outputs[5], "99")
	assert.Contains(t, ui.Outputs[5], "app instance was stopped")
	assert.Contains(t, ui.Outputs[5], "77")
	assert.Contains(t, ui.Outputs[6], timestamp.Local().Format(TIMESTAMP_FORMAT))
	assert.Contains(t, ui.Outputs[6], "98")
	assert.Contains(t, ui.Outputs[6], "app instance exited")
	assert.Contains(t, ui.Outputs[6], "78")
}

func TestEventsWhenNoEventsAvailable(t *testing.T) {
	reqFactory, eventsRepo := getEventsDependencies()
	reqFactory.Application = cf.Application{Name: "my-app", Guid: "my-app-guid"}

	ui := callEvents(t, []string{"my-app"}, reqFactory, eventsRepo)

	assert.Contains(t, ui.Outputs[0], "events")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "no events")
	assert.Contains(t, ui.Outputs[3], "my-app")
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

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewEvents(ui, config, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
