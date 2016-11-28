package application_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/api/appevents/appeventsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TIMESTAMP_FORMAT = "2006-01-02T15:04:05.00-0700"

var _ = Describe("events command", func() {
	var (
		reqFactory  *requirementsfakes.FakeFactory
		eventsRepo  *appeventsfakes.FakeAppEventsRepository
		ui          *testterm.FakeUI
		config      *coreconfigfakes.FakeRepository
		deps        commandregistry.Dependency
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		applicationRequirement   *requirementsfakes.FakeApplicationRequirement

		cmd *application.Events
	)

	BeforeEach(func() {
		cmd = &application.Events{}

		ui = new(testterm.FakeUI)
		eventsRepo = new(appeventsfakes.FakeAppEventsRepository)
		config = new(coreconfigfakes.FakeRepository)

		config.OrganizationFieldsReturns(models.OrganizationFields{Name: "my-org"})
		config.SpaceFieldsReturns(models.SpaceFields{Name: "my-space"})
		config.UsernameReturns("my-user")

		deps = commandregistry.Dependency{
			UI:          ui,
			RepoLocator: api.RepositoryLocator{}.SetAppEventsRepository(eventsRepo),
			Config:      config,
		}

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		reqFactory = new(requirementsfakes.FakeFactory)
		loginRequirement = &passingRequirement{Name: "login-requirement"}
		reqFactory.NewLoginRequirementReturns(loginRequirement)
		targetedSpaceRequirement = &passingRequirement{Name: "targeted-space-requirement"}
		reqFactory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)
		applicationRequirement = new(requirementsfakes.FakeApplicationRequirement)
		applicationRequirement.ExecuteReturns(nil)
		reqFactory.NewApplicationRequirementReturns(applicationRequirement)
	})

	Describe("Requirements", func() {
		BeforeEach(func() {
			cmd.SetDependency(deps, false)
		})

		Context("when not provided exactly 1 argument", func() {
			It("fails", func() {
				err := flagContext.Parse("too", "many")
				Expect(err).NotTo(HaveOccurred())
				_, err = cmd.Requirements(reqFactory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage", "Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			var actualRequirements []requirements.Requirement

			BeforeEach(func() {
				err := flagContext.Parse("service-name")
				Expect(err).NotTo(HaveOccurred())
				actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a LoginRequirement", func() {
				Expect(reqFactory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				Expect(reqFactory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns a ApplicationRequirement", func() {
				Expect(reqFactory.NewApplicationRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(applicationRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var executeCmdErr error

		BeforeEach(func() {
			applicationRequirement.GetApplicationReturns(models.Application{
				ApplicationFields: models.ApplicationFields{
					Name: "my-app",
					GUID: "my-app-guid",
				},
			})

			err := flagContext.Parse("my-app")
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			executeCmdErr = cmd.Execute(flagContext)
		})

		Context("when no events exist", func() {
			BeforeEach(func() {
				eventsRepo.RecentEventsReturns([]models.EventFields{}, nil)

				cmd.SetDependency(deps, false)
				cmd.Requirements(reqFactory, flagContext)
			})

			It("tells the user", func() {
				Expect(executeCmdErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"events", "my-app"},
					[]string{"No events", "my-app"},
				))
			})
		})

		Context("when events exist", func() {
			var (
				earlierTimestamp time.Time
				timestamp        time.Time
			)

			BeforeEach(func() {
				var err error

				earlierTimestamp, err = time.Parse(TIMESTAMP_FORMAT, "1999-12-31T23:59:11.00-0000")
				Expect(err).NotTo(HaveOccurred())

				timestamp, err = time.Parse(TIMESTAMP_FORMAT, "2000-01-01T00:01:11.00-0000")
				Expect(err).NotTo(HaveOccurred())

				eventsRepo.RecentEventsReturns([]models.EventFields{
					{
						GUID:        "event-guid-1",
						Name:        "app crashed",
						Timestamp:   earlierTimestamp,
						Description: "reason: app instance exited, exit_status: 78",
						Actor:       "george-clooney",
						ActorName:   "George Clooney",
					},
					{
						GUID:        "event-guid-2",
						Name:        "app crashed",
						Timestamp:   timestamp,
						Description: "reason: app instance was stopped, exit_status: 77",
						Actor:       "marcel-marceau",
					},
				}, nil)

				cmd.SetDependency(deps, false)
				cmd.Requirements(reqFactory, flagContext)
			})

			It("lists events given an app name", func() {
				Expect(executeCmdErr).NotTo(HaveOccurred())
				Expect(eventsRepo.RecentEventsCallCount()).To(Equal(1))
				appGUID, limit := eventsRepo.RecentEventsArgsForCall(0)
				Expect(limit).To(Equal(int64(50)))
				Expect(appGUID).To(Equal("my-app-guid"))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting events for app", "my-app", "my-org", "my-space", "my-user"},
					[]string{"time", "event", "actor", "description"},
					[]string{earlierTimestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "George Clooney", "app instance exited", "78"},
					[]string{timestamp.Local().Format(TIMESTAMP_FORMAT), "app crashed", "marcel-marceau", "app instance was stopped", "77"},
				))
			})
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				eventsRepo.RecentEventsReturns([]models.EventFields{}, errors.New("welp"))

				cmd.SetDependency(deps, false)
				cmd.Requirements(reqFactory, flagContext)
			})

			It("tells the user when an error occurs", func() {
				Expect(executeCmdErr).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"events", "my-app"},
				))
				errStr := executeCmdErr.Error()
				Expect(errStr).To(ContainSubstring("welp"))
			})
		})
	})
})
