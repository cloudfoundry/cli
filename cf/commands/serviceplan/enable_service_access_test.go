package serviceplan_test

import (
	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/serviceplan"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("enable-service-access command", func() {
	var (
		ui                  *testterm.FakeUI
		actor               *testactor.FakeServicePlanActor
		requirementsFactory *testreq.FakeReqFactory

		publicServicePlan  models.ServicePlanFields
		privateServicePlan models.ServicePlanFields
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		actor = &testactor.FakeServicePlanActor{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewEnableServiceAccess(ui, configuration.NewRepositoryWithDefaults(), actor)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			runCommand("foo")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when it does not recieve any arguments", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			publicServicePlan = models.ServicePlanFields{
				Guid:   "public-service-plan-guid",
				Name:   "public-service-plan",
				Public: true,
			}

			privateServicePlan = models.ServicePlanFields{
				Guid:   "private-service-plan-guid",
				Name:   "private-service-plan",
				Public: false,
			}
		})

		Context("when the named service exists", func() {
			It("tells the user the service is already public if all plans are public", func() {
			})

			It("tells the user private services have been set to public", func() {

			})

			Context("The user provides a plan", func() {
				It("tells the user if the plan is already public", func() {
					actor.GetSingleServicePlanForServiceReturns(publicServicePlan, nil)

					Expect(runCommand("public-service-plan", "p", "public-service-plan")).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Plan", "for service", "is already public"},
						[]string{"OK"},
					))
				})
			})
		})
	})
})
