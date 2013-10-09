package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
	"time"
)

func TestEventsRequirements(t *testing.T) {
	ui := new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("events", []string{"my-app"})
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	eventsRepo := &testhelpers.FakeAppEventsRepo{}

	cmd := NewEvents(ui, eventsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestEventsFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	eventsRepo := &testhelpers.FakeAppEventsRepo{}
	ui := new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("events", []string{})

	cmd := NewEvents(ui, eventsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestEventsSuccess(t *testing.T) {

	timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
	assert.NoError(t, err)

	app := cf.Application{
		Name: "my-app",
		Guid: "my-app-guid",
	}

	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		Application:          app,
	}

	eventsRepo := &testhelpers.FakeAppEventsRepo{
		Events: []cf.Event{
			{
				InstanceIndex:   9,
				Timestamp:       timestamp,
				ExitDescription: "app instance exited",
				ExitStatus:      7,
			},
		},
	}

	ui := new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("events", []string{"my-app"})

	cmd := NewEvents(ui, eventsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "events")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "time")
	assert.Contains(t, ui.Outputs[2], "instance")
	assert.Contains(t, ui.Outputs[2], "description")
	assert.Contains(t, ui.Outputs[2], "exit status")
	assert.Contains(t, ui.Outputs[3], timestamp.Local().Format(TIMESTAMP_FORMAT))
	assert.Contains(t, ui.Outputs[3], "9")
	assert.Contains(t, ui.Outputs[3], "app instance exited")
	assert.Contains(t, ui.Outputs[3], "7")
}

func TestEventsWhenNoEventsAvailable(t *testing.T) {

	app := cf.Application{
		Name: "my-app",
		Guid: "my-app-guid",
	}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	eventsRepo := &testhelpers.FakeAppEventsRepo{}
	ui := new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("events", []string{"my-app"})

	cmd := NewEvents(ui, eventsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "events")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "no events")
	assert.Contains(t, ui.Outputs[2], "my-app")
}
