package servicebroker_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"

	"github.com/cloudfoundry/cli/cf/commands/servicebroker"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("service-brokers command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		repo                *testapi.FakeServiceBrokerRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceBrokerRepository(repo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("service-brokers").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		repo = &testapi.FakeServiceBrokerRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(testcmd.RunCliCommand("service-brokers", []string{}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd command_registry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &servicebroker.ListServiceBrokers{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(requirementsFactory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	It("lists service brokers", func() {
		repo.ListServiceBrokersStub = func(callback func(models.ServiceBroker) bool) error {
			sbs := []models.ServiceBroker{
				{
					Name: "service-broker-to-list-a",
					Guid: "service-broker-to-list-guid-a",
					Url:  "http://service-a-url.com",
				},
				{
					Name: "service-broker-to-list-b",
					Guid: "service-broker-to-list-guid-b",
					Url:  "http://service-b-url.com",
				},
				{
					Name: "service-broker-to-list-c",
					Guid: "service-broker-to-list-guid-c",
					Url:  "http://service-c-url.com",
				},
			}

			for _, sb := range sbs {
				callback(sb)
			}

			return nil
		}

		testcmd.RunCliCommand("service-brokers", []string{}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as", "my-user"},
			[]string{"name", "url"},
			[]string{"service-broker-to-list-a", "http://service-a-url.com"},
			[]string{"service-broker-to-list-b", "http://service-b-url.com"},
			[]string{"service-broker-to-list-c", "http://service-c-url.com"},
		))
	})

	It("lists service brokers by alphabetical order", func() {
		repo.ListServiceBrokersStub = func(callback func(models.ServiceBroker) bool) error {
			sbs := []models.ServiceBroker{
				{
					Name: "z-service-broker-to-list",
					Guid: "z-service-broker-to-list-guid-a",
					Url:  "http://service-a-url.com",
				},
				{
					Name: "a-service-broker-to-list",
					Guid: "a-service-broker-to-list-guid-c",
					Url:  "http://service-c-url.com",
				},
				{
					Name: "fun-service-broker-to-list",
					Guid: "fun-service-broker-to-list-guid-b",
					Url:  "http://service-b-url.com",
				},
				{
					Name: "123-service-broker-to-list",
					Guid: "123-service-broker-to-list-guid-c",
					Url:  "http://service-d-url.com",
				},
			}

			for _, sb := range sbs {
				callback(sb)
			}

			return nil
		}

		testcmd.RunCliCommand("service-brokers", []string{}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(BeInDisplayOrder(
			[]string{"Getting service brokers as", "my-user"},
			[]string{"name", "url"},
			[]string{"123-service-broker-to-list", "http://service-d-url.com"},
			[]string{"a-service-broker-to-list", "http://service-c-url.com"},
			[]string{"fun-service-broker-to-list", "http://service-b-url.com"},
			[]string{"z-service-broker-to-list", "http://service-a-url.com"},
		))
	})

	It("says when no service brokers were found", func() {
		testcmd.RunCliCommand("service-brokers", []string{}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as", "my-user"},
			[]string{"No service brokers found"},
		))
	})

	It("reports errors when listing service brokers", func() {
		repo.ListServiceBrokersReturns(errors.New("Error finding service brokers"))
		testcmd.RunCliCommand("service-brokers", []string{}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as ", "my-user"},
		))
		Expect(strings.Join(ui.Outputs, "\n")).To(MatchRegexp(`FAILED\nError finding service brokers`))
	})
})
