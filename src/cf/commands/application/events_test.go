package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

func TestEventsRequirements(t *testing.T) {
	ui := new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", []string{"my-app"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	eventsRepo := &testapi.FakeAppEventsRepo{}

	cmd := NewEvents(ui, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestEventsFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	eventsRepo := &testapi.FakeAppEventsRepo{}
	ui := new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", []string{})

	cmd := NewEvents(ui, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEventsSuccess(t *testing.T) {

	timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
	assert.NoError(t, err)

	app := cf.Application{
		Name: "my-app",
		Guid: "my-app-guid",
	}

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		Application:          app,
	}

	eventsRepo := &testapi.FakeAppEventsRepo{
		Events: []cf.Event{
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
		},
	}

	ui := new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", []string{"my-app"})

	cmd := NewEvents(ui, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "events")
	assert.Contains(t, ui.Outputs[0], "my-app")
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
	app := cf.Application{
		Name: "my-app",
		Guid: "my-app-guid",
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	eventsRepo := &testapi.FakeAppEventsRepo{}
	ui := new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", []string{"my-app"})

	cmd := NewEvents(ui, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "events")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "no events")
	assert.Contains(t, ui.Outputs[3], "my-app")
}
