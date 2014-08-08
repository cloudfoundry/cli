package serviceaccess_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors"
	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/serviceaccess"
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

			It("tells the user if all plans were already public", func() {
				actor.UpdateAllPlansForServiceReturns(true, nil)

				Expect(runCommand([]string{"service"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"All plans of the service", "are already accessible for all orgs"},
					[]string{"OK"},
				))
			})

			It("tells the user the plans are being updated if they weren't all already public", func() {
				actor.UpdateAllPlansForServiceReturns(false, nil)

				Expect(runCommand([]string{"service"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Enabling access to all plans of service service for all orgs as my-user..."},
					[]string{"OK"},
				))
			})

			It("prints an error if updating one of the plans fails", func() {
				actor.UpdateAllPlansForServiceReturns(true, errors.New("Kaboom!"))

				Expect(runCommand([]string{"service"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Kaboom!"},
				))
			})

			Context("The user provides a plan", func() {
				It("prints an error if the service does not exist", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.All, errors.New("could not find service"))

					Expect(runCommand([]string{"-p", "service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find service"},
					))
				})

				It("tells the user if the plan is already public", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.All, nil)

					Expect(runCommand([]string{"-p", "public-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"The plan is already accessible for all orgs"},
						[]string{"OK"},
					))
				})

				It("tells the user the plan is being updated if it is not public", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.None, nil)

					Expect(runCommand([]string{"-p", "private-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Enabling access of plan private-service-plan for service service"},
						[]string{"OK"},
					))
				})
			})

			Context("the user provides a plan and org", func() {
				It("fails if the org does not exist", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.All, errors.New("could not find org"))

					Expect(runCommand([]string{"-p", "service-plan", "-o", "not-findable-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find org"},
					))
				})

				It("tells the user if the plan is already public", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.All, nil)

					Expect(runCommand([]string{"-p", "public-service-plan", "-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"The plan is already accessible for this org"},
						[]string{"OK"},
					))
				})

				It("tells the user the plan is being updated if it is not public", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.None, nil)

					Expect(runCommand([]string{"-p", "private-service-plan", "-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Enabling access to plan private-service-plan of service service for org my-org as"},
						[]string{"OK"},
					))
				})
			})
			Context("the user provides an org", func() {
				It("fails if the org does not exist", func() {
					actor.UpdateOrgForServiceReturns(false, errors.New("could not find org"))

					Expect(runCommand([]string{"-o", "not-findable-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find org"},
					))
				})

				It("tells the user if the service's plans are already accessible", func() {
					actor.UpdateOrgForServiceReturns(true, nil)

					Expect(runCommand([]string{"-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"All plans of the service are already accessible for the org"},
						[]string{"OK"},
					))
				})

				It("tells the user the service's plans are being updated if it is not accessible", func() {
					actor.UpdateOrgForServiceReturns(false, nil)

					Expect(runCommand([]string{"-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Enabling access to all plans of service service for the org my-org as"},
						[]string{"OK"},
					))
				})
			})
		})
	})
})
