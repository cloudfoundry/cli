package serviceaccess_test

import (
	"errors"

	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
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
		tokenRefresher      *testapi.FakeAuthenticationRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		actor = &testactor.FakeServiceActor{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		tokenRefresher = &testapi.FakeAuthenticationRepository{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewServiceAccess(ui, testconfig.NewRepositoryWithDefaults(), actor, tokenRefresher)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
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

			actor.FilterBrokersReturns([]models.ServiceBroker{
				serviceBroker1,
				serviceBroker2,
			},
				nil,
			)
		})

		It("refreshes the auth token", func() {
			runCommand("service")
			Expect(tokenRefresher.RefreshTokenCalled).To(BeTrue())
		})

		Context("when refreshing the auth token fails", func() {
			It("fails and returns the error", func() {
				tokenRefresher.RefreshTokenError = errors.New("Refreshing went wrong")
				runCommand()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Refreshing went wrong"},
					[]string{"FAILED"},
				))
			})
		})

		Context("When no flags are provided", func() {
			It("tells the user it is obtaining the service access", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access as", "my-user"},
				))
			})

			It("prints all of the brokers", func() {
				runCommand()
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

		Context("When the broker flag is provided", func() {
			It("tells the user it is obtaining the services access for a particular broker", func() {
				runCommand("-b", "brokername1")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1 as", "my-user"},
				))
			})
		})

		Context("when the service flag is provided", func() {
			It("tells the user it is obtaining the service access for a particular service", func() {
				runCommand("-e", "my-service-1")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for service my-service-1 as", "my-user"},
				))
			})
		})

		Context("when the org flag is provided", func() {
			It("tells the user it is obtaining the service access for a particular org", func() {
				runCommand("-o", "fwip")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for organization fwip as", "my-user"},
				))
			})
		})

		Context("when the broker and service flag are both provided", func() {
			It("tells the user it is obtaining the service access for a particular broker and service", func() {
				runCommand("-b", "brokername1", "-e", "my-service-1")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and service my-service-1", "as", "my-user"},
				))
			})
		})

		Context("when the broker and org name are both provided", func() {
			It("tells the user it is obtaining the service access for a particular broker and org", func() {
				runCommand("-b", "brokername1", "-o", "fwip")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and organization fwip", "as", "my-user"},
				))
			})
		})

		Context("when the service and org name are both provided", func() {
			It("tells the user it is obtaining the service access for a particular service and org", func() {
				runCommand("-e", "my-service-1", "-o", "fwip")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for service my-service-1", "and organization fwip", "as", "my-user"},
				))
			})
		})

		Context("when all flags are provided", func() {
			It("tells the user it is filtering on all options", func() {
				runCommand("-b", "brokername1", "-e", "my-service-1", "-o", "fwip")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting service access", "for broker brokername1", "and service my-service-1", "and organization fwip", "as", "my-user"},
				))
			})
		})
	})
})
