package application_test

import (
	. "cf/commands/application"
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
	"time"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestEventsRequirements", func() {
			reqFactory, eventsRepo := getEventsDependencies()

			callEvents([]string{"my-app"}, reqFactory, eventsRepo)

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestEventsFailsWithUsage", func() {

			reqFactory, eventsRepo := getEventsDependencies()
			ui := callEvents([]string{}, reqFactory, eventsRepo)

			assert.True(mr.T(), ui.FailedWithUsage)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestEventsSuccess", func() {

			timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
			assert.NoError(mr.T(), err)

			reqFactory, eventsRepo := getEventsDependencies()
			app := models.Application{}
			app.Name = "my-app"
			reqFactory.Application = app

			event1 := models.EventFields{}
			event1.Guid = "event-guid-1"
			event1.Name = "app crashed"
			event1.Timestamp = timestamp
			event1.Description = "reason: app instance exited, exit_status: 78"

			event2 := models.EventFields{}
			event2.Guid = "event-guid-2"
			event2.Name = "app crashed"
			event2.Timestamp = timestamp
			event2.Description = "reason: app instance was stopped, exit_status: 77"

			eventsRepo.Events = []models.EventFields{
				event1,
				event2,
			}

			ui := callEvents([]string{"my-app"}, reqFactory, eventsRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting events for app", "my-app", "my-org", "my-space", "my-user"},
				{"time", "event", "description"},
				{timestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "app instance exited", "78"},
				{timestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "app instance was stopped", "77"},
			})
		})
		It("TestEventsWhenNoEventsAvailable", func() {

			reqFactory, eventsRepo := getEventsDependencies()
			app := models.Application{}
			app.Name = "my-app"
			reqFactory.Application = app

			ui := callEvents([]string{"my-app"}, reqFactory, eventsRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"events", "my-app"},
				{"No events", "my-app"},
			})
		})
	})
}

func getEventsDependencies() (reqFactory *testreq.FakeReqFactory, eventsRepo *testapi.FakeAppEventsRepo) {
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	eventsRepo = &testapi.FakeAppEventsRepo{}
	return
}

func callEvents(args []string, reqFactory *testreq.FakeReqFactory, eventsRepo *testapi.FakeAppEventsRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("events", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewEvents(ui, configRepo, eventsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
