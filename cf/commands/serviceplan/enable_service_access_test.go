package serviceplan_test

import (
	"errors"

	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
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
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		actor = &testactor.FakeServicePlanActor{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args []string) bool {
		cmd := NewEnableServiceAccess(ui, configuration.NewRepositoryWithDefaults(), actor)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			runCommand([]string{"foo"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when it does not recieve any arguments", func() {
			requirementsFactory.LoginSuccess = true
			runCommand(nil)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the named service exists", func() {
			It("returns OK when ran successfully", func() {
				Expect(runCommand([]string{"service"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
				))
			})

			Context("The user provides a plan", func() {
				It("prints an error if the service does not exist", func() {
					actor.UpdateSinglePlanForServiceReturns(true, errors.New("could not find service"))

					Expect(runCommand([]string{"-p", "service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find service"},
					))
				})

				It("tells the user if the plan is already public", func() {
					actor.UpdateSinglePlanForServiceReturns(true, nil)

					Expect(runCommand([]string{"-p", "public-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Plan", "for service", "is already public"},
						[]string{"OK"},
					))
				})

				It("tells the user the plan is being updated if it is not public", func() {
					actor.UpdateSinglePlanForServiceReturns(false, nil)

					Expect(runCommand([]string{"-p", "private-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Enabling access of plan private-service-plan for service service"},
						[]string{"OK"},
					))
				})
			})
		})
	})
})
