package application_test

import (
	. "cf/commands/application"
	"cf/errors"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("events command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		eventsRepo          *testapi.FakeAppEventsRepo
		ui                  *testterm.FakeUI
	)

	BeforeEach(func() {
		eventsRepo = &testapi.FakeAppEventsRepo{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		ui = new(testterm.FakeUI)
	})

	runCommand := func(args ...string) {
		configRepo := testconfig.NewRepositoryWithDefaults()
		cmd := NewEvents(ui, configRepo, eventsRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("events", args), requirementsFactory)
	}

	It("fails with usage when called without an app name", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("lists events given an app name", func() {
		earlierTimestamp, err := time.Parse(TIMESTAMP_FORMAT, "1999-12-31T23:59:11.00-0000")
		Expect(err).NotTo(HaveOccurred())

		timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
		Expect(err).NotTo(HaveOccurred())

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory.Application = app

		eventsRepo.RecentEventsReturns.Events = []models.EventFields{
			{
				Guid:        "event-guid-1",
				Name:        "app crashed",
				Timestamp:   earlierTimestamp,
				Description: "reason: app instance exited, exit_status: 78",
			},
			{
				Guid:        "event-guid-2",
				Name:        "app crashed",
				Timestamp:   timestamp,
				Description: "reason: app instance was stopped, exit_status: 77",
			},
		}

		runCommand("my-app")

		Expect(eventsRepo.RecentEventsArgs.Limit).To(Equal(uint64(50)))
		Expect(eventsRepo.RecentEventsArgs.AppGuid).To(Equal("my-app-guid"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting events for app", "my-app", "my-org", "my-space", "my-user"},
			{"time", "event", "description"},
			{earlierTimestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "app instance exited", "78"},
			{timestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "app instance was stopped", "77"},
		})
	})

	It("tells the user when an error occurs", func() {
		eventsRepo.RecentEventsReturns.Error = errors.New("welp")

		app := models.Application{}
		app.Name = "my-app"
		requirementsFactory.Application = app

		runCommand("my-app")

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"events", "my-app"},
			{"FAILED"},
			{"welp"},
		})
	})

	It("tells the user when no events exist for that app", func() {
		app := models.Application{}
		app.Name = "my-app"
		requirementsFactory.Application = app

		runCommand("my-app")

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"events", "my-app"},
			{"No events", "my-app"},
		})
	})
})
