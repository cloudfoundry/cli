package application_test

import (
	"time"

	"github.com/cloudfoundry/cli/cf/api/appevents/appeventsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("events command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		eventsRepo          *appeventsfakes.FakeAppEventsRepository
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	const TIMESTAMP_FORMAT = "2006-01-02T15:04:05.00-0700"

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetAppEventsRepository(eventsRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("events").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		eventsRepo = new(appeventsfakes.FakeAppEventsRepository)
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("events", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails with usage when called without an app name", func() {
		passed := runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "argument"},
		))
		Expect(passed).To(BeFalse())
	})

	It("lists events given an app name", func() {
		earlierTimestamp, err := time.Parse(TIMESTAMP_FORMAT, "1999-12-31T23:59:11.00-0000")
		Expect(err).NotTo(HaveOccurred())

		timestamp, err := time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
		Expect(err).NotTo(HaveOccurred())

		app := models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
		requirementsFactory.Application = app

		eventsRepo.RecentEventsReturns([]models.EventFields{
			{
				GUID:        "event-guid-1",
				Name:        "app crashed",
				Timestamp:   earlierTimestamp,
				Description: "reason: app instance exited, exit_status: 78",
				ActorName:   "George Clooney",
			},
			{
				GUID:        "event-guid-2",
				Name:        "app crashed",
				Timestamp:   timestamp,
				Description: "reason: app instance was stopped, exit_status: 77",
				ActorName:   "Marcel Marceau",
			},
		}, nil)

		runCommand("my-app")

		Expect(eventsRepo.RecentEventsCallCount()).To(Equal(1))
		appGUID, limit := eventsRepo.RecentEventsArgsForCall(0)
		Expect(limit).To(Equal(int64(50)))
		Expect(appGUID).To(Equal("my-app-guid"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting events for app", "my-app", "my-org", "my-space", "my-user"},
			[]string{"time", "event", "actor", "description"},
			[]string{earlierTimestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "George Clooney", "app instance exited", "78"},
			[]string{timestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "Marcel Marceau", "app instance was stopped", "77"},
		))
	})

	It("tells the user when an error occurs", func() {
		eventsRepo.RecentEventsReturns(nil, errors.New("welp"))

		app := models.Application{}
		app.Name = "my-app"
		requirementsFactory.Application = app

		runCommand("my-app")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"events", "my-app"},
			[]string{"FAILED"},
			[]string{"welp"},
		))
	})

	It("tells the user when no events exist for that app", func() {
		app := models.Application{}
		app.Name = "my-app"
		requirementsFactory.Application = app

		runCommand("my-app")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"events", "my-app"},
			[]string{"No events", "my-app"},
		))
	})
})
