package serviceplan_test

import (
	"errors"

	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/serviceplan"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-access command", func() {
	var (
		ui                  *testterm.FakeUI
		actor               *testactor.FakeServiceActor
		requirementsFactory *testreq.FakeReqFactory
		serviceBroker1      models.ServiceBroker
		serviceBroker2      models.ServiceBroker
		serviceBroker3      models.ServiceBroker
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		actor = &testactor.FakeServiceActor{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args []string) bool {
		cmd := NewServiceAccess(ui, testconfig.NewRepositoryWithDefaults(), actor)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand(nil)).ToNot(HavePassedRequirements())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			serviceBroker1 = models.ServiceBroker{
				Guid: "broker1",
				Name: "brokername1",
				Services: []models.ServiceOffering{
					{
						ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-1"},
						Plans: []models.ServicePlanFields{
							{Name: "beep", Public: true},
							{Name: "bot", Public: false},
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
				Guid: "broker2",
				Name: "brokername2",
				Services: []models.ServiceOffering{
					{ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-3"}},
				},
			}
			serviceBroker3 = models.ServiceBroker{
				Guid: "broker3",
				Name: "brokername3",
				Services: []models.ServiceOffering{
					{
						ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-4"},
						Plans: []models.ServicePlanFields{
							{Name: "weepweep", Public: true},
							{Name: "aoooga", Public: false, OrgNames: []string{"plink", "plonk"}},
						},
					},
				},
			}

			actor.GetBrokerWithSingleServiceReturns([]models.ServiceBroker{serviceBroker3}, nil)
			actor.GetBrokerWithDependenciesReturns([]models.ServiceBroker{serviceBroker1}, nil)

			actor.GetAllBrokersWithDependenciesReturns([]models.ServiceBroker{
				serviceBroker1,
				serviceBroker2,
			},
				nil,
			)
		})

		It("prints all of the brokers", func() {
			runCommand(nil)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"broker: brokername1"},
				[]string{"service", "plan", "access", "orgs"},
				[]string{"my-service-1", "beep", "all"},
				[]string{"my-service-1", "bot", "none"},
				[]string{"my-service-1", "boop", "limited", "fwip", "brzzt"},
				[]string{"my-service-2", "petaloideous-noncelebration"},
				[]string{"broker: brokername2"},
				[]string{"service", "plan", "access", "orgs"},
				[]string{"my-service-3"},
			))
		})

		Context("with a -b flag", func() {
			It("only prints out the specified broker's information", func() {
				args := []string{"-b", "broker1"}
				runCommand(args)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"broker: brokername1"},
					[]string{"service", "plan", "access", "orgs"},
					[]string{"my-service-1", "beep", "all"},
					[]string{"my-service-1", "boop", "limited", "fwip", "brzzt"},
					[]string{"my-service-2", "petaloideous-noncelebration"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"broker: brokername2"},
					[]string{"my-service-3"},
				))
			})
		})

		Context("with a -s flag", func() {
			It("only prints out the specified service's information", func() {
				args := []string{"-e", "my-service-4"}
				runCommand(args)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"broker: brokername3"},
					[]string{"service", "plan", "access", "orgs"},
					[]string{"my-service-4", "weepweep", "all"},
					[]string{"my-service-4", "aoooga", "limited", "plink", "plonk"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"my-service-1", "beep", "all"},
					[]string{"my-service-1", "boop", "limited", "fwip", "brzzt"},
					[]string{"my-service-2", "petaloideous-noncelebration"},
					[]string{"broker: brokername1"},
					[]string{"broker: brokername2"},
					[]string{"my-service-3"},
				))
			})
		})

		Context("when both -b and -s are in play", func() {
			It("prints the intersection set", func() {
				args := []string{"-b", "broker1", "-e", "my-service-4"}
				runCommand(args)
				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"broker: brokername3"},
				))
			})

			It("returns an error message when the broker does not exist", func() {
				actor.GetBrokerWithDependenciesReturns(nil, errors.New("Service broker BroKer1 not found."))
				args := []string{"-b", "BroKer1", "-e", "my-service-4"}
				runCommand(args)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Failed fetching service brokers."},
					[]string{"Service broker BroKer1 not found."},
				))
			})

			It("returns an error message when the service does not exist", func() {
				actor.GetBrokerWithSingleServiceReturns(nil, errors.New("Service My-Service-4 not found."))
				args := []string{"-b", "broker1", "-e", "My-Service-4"}
				runCommand(args)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Failed fetching service."},
					[]string{"Service My-Service-4 not found."},
				))
			})
		})
	})
})
