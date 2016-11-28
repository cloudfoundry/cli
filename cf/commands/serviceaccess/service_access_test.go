package serviceaccess_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/cf/actors/actorsfakes"
	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/serviceaccess"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-access command", func() {
	var (
		ui                  *testterm.FakeUI
		actor               *actorsfakes.FakeServiceActor
		requirementsFactory *requirementsfakes.FakeFactory
		serviceBroker1      models.ServiceBroker
		serviceBroker2      models.ServiceBroker
		authRepo            *authenticationfakes.FakeRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.ServiceHandler = actor
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("service-access").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		actor = new(actorsfakes.FakeServiceActor)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		authRepo = new(authenticationfakes.FakeRepository)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("service-access", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		Context("when arguments are provided", func() {
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &serviceaccess.ServiceAccess{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				err = testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			serviceBroker1 = models.ServiceBroker{
				GUID: "broker1",
				Name: "brokername1",
				Services: []models.ServiceOffering{
					{
						ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-1"},
						Plans: []models.ServicePlanFields{
							{Name: "beep", Public: true},
							{Name: "burp", Public: false},
							{Name: "boop", Public: false, OrgNames: []string{"fwip", "brzzt"}},
						},
					},
					{
						ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-2"},
						Plans: []models.ServicePlanFields{
							{Name: "petaloideous-noncelebration", Public: false},
						},
					},
				},
			}
			serviceBroker2 = models.ServiceBroker{
				GUID: "broker2",
				Name: "brokername2",
				Services: []models.ServiceOffering{
					{ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-3"}},
				},
			}

			actor.FilterBrokersReturns([]models.ServiceBroker{
				serviceBroker1,
				serviceBroker2,
			},
				nil,
			)
		})

		It("refreshes the auth token", func() {
			runCommand()
			Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
		})

		Context("when refreshing the auth token fails", func() {
			It("fails and returns the error", func() {
				authRepo.RefreshAuthTokenReturns("", errors.New("Refreshing went wrong"))
				runCommand()

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Refreshing went wrong"},
					[]string{"FAILED"},
				))
			})
		})

		Context("When no flags are provided", func() {
			It("tells the user it is obtaining the service access", func() {
				runCommand()
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access as", "my-user"},
				))
			})

			It("prints all of the brokers", func() {
				runCommand()
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"broker: brokername1"},
					[]string{"service", "plan", "access", "orgs"},
					[]string{"my-service-1", "beep", "all"},
					[]string{"my-service-1", "burp", "none"},
					[]string{"my-service-1", "boop", "limited", "fwip", "brzzt"},
					[]string{"my-service-2", "petaloideous-noncelebration"},
					[]string{"broker: brokername2"},
					[]string{"service", "plan", "access", "orgs"},
					[]string{"my-service-3"},
				))
			})
		})

		Context("When the broker flag is provided", func() {
			It("tells the user it is obtaining the services access for a particular broker", func() {
				runCommand("-b", "brokername1")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1 as", "my-user"},
				))
			})
		})

		Context("when the service flag is provided", func() {
			It("tells the user it is obtaining the service access for a particular service", func() {
				runCommand("-e", "my-service-1")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for service my-service-1 as", "my-user"},
				))
			})
		})

		Context("when the org flag is provided", func() {
			It("tells the user it is obtaining the service access for a particular org", func() {
				runCommand("-o", "fwip")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for organization fwip as", "my-user"},
				))
			})
		})

		Context("when the broker and service flag are both provided", func() {
			It("tells the user it is obtaining the service access for a particular broker and service", func() {
				runCommand("-b", "brokername1", "-e", "my-service-1")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and service my-service-1", "as", "my-user"},
				))
			})
		})

		Context("when the broker and org name are both provided", func() {
			It("tells the user it is obtaining the service access for a particular broker and org", func() {
				runCommand("-b", "brokername1", "-o", "fwip")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and organization fwip", "as", "my-user"},
				))
			})
		})

		Context("when the service and org name are both provided", func() {
			It("tells the user it is obtaining the service access for a particular service and org", func() {
				runCommand("-e", "my-service-1", "-o", "fwip")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for service my-service-1", "and organization fwip", "as", "my-user"},
				))
			})
		})

		Context("when all flags are provided", func() {
			It("tells the user it is filtering on all options", func() {
				runCommand("-b", "brokername1", "-e", "my-service-1", "-o", "fwip")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and service my-service-1", "and organization fwip", "as", "my-user"},
				))
			})
		})
		Context("when filter brokers returns an error", func() {
			It("gives only the access error", func() {
				err := errors.New("Error finding service brokers")
				actor.FilterBrokersReturns([]models.ServiceBroker{}, err)
				runCommand()

				Expect(strings.Join(ui.Outputs(), "\n")).To(MatchRegexp(`FAILED\nError finding service brokers`))
			})
		})
	})
})
