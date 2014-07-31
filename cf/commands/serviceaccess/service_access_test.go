package serviceaccess_test

import (
	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/serviceaccess"
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
				Guid: "broker2",
				Name: "brokername2",
				Services: []models.ServiceOffering{
					{ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service-3"}},
				},
			}

			actor.GetBrokerWithDependenciesReturns([]models.ServiceBroker{serviceBroker1}, nil)

			actor.FilterBrokersReturns([]models.ServiceBroker{
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
				[]string{"my-service-1", "burp", "none"},
				[]string{"my-service-1", "boop", "limited", "fwip", "brzzt"},
				[]string{"my-service-2", "petaloideous-noncelebration"},
				[]string{"broker: brokername2"},
				[]string{"service", "plan", "access", "orgs"},
				[]string{"my-service-3"},
			))
		})
	})
})
