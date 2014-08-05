package serviceaccess_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors"
	testactor "github.com/cloudfoundry/cli/cf/actors/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/serviceaccess"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disable-service-access command", func() {
	var (
		ui                  *testterm.FakeUI
		actor               *testactor.FakeServicePlanActor
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}
		actor = &testactor.FakeServicePlanActor{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args []string) bool {
		cmd := NewDisableServiceAccess(ui, configuration.NewRepositoryWithDefaults(), actor)
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

			Context("The user provides a plan", func() {
				It("prints an error if the service does not exist", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.All, errors.New("could not find service"))

					Expect(runCommand([]string{"-p", "service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find service"},
					))
				})

				It("tells the user if the plan is already private", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.None, nil)

					Expect(runCommand([]string{"-p", "private-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Plan", "for service", "is already private"},
						[]string{"OK"},
					))
				})

				It("tells the user the plan is being updated if it is not private", func() {
					actor.UpdateSinglePlanForServiceReturns(actors.All, nil)

					Expect(runCommand([]string{"-p", "public-service-plan", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Disabling access of plan public-service-plan for service service"},
						[]string{"OK"},
					))
				})
			})

			Context("the user provides a plan and org", func() {
				Context("when the command is confirmed", func() {
					It("returns OK when ran successfully", func() {
						actor.UpdatePlanAndOrgForServiceReturns(actors.Limited, nil)
						Expect(runCommand([]string{"-p", "limited-service-plan", "-o", "my-org", "service"})).To(BeTrue())
						Expect(ui.Prompts).To(ContainSubstrings([]string{
							"Really disable access",
							"plan limited-service-plan",
							"service service",
							"org my-org",
						}))
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Disabling access to",
								"plan limited-service-plan",
								"service service",
								"org my-org",
							},
							[]string{"OK"},
						))
					})
					It("skips confirmation when the -f flag is given", func() {
						actor.UpdatePlanAndOrgForServiceReturns(actors.Limited, nil)
						runCommand([]string{"-f", "-p", "limited-service-plan", "-o", "my-org", "service"})

						Expect(ui.Prompts).To(BeEmpty())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Disabling access to plan limited-service-plan of service service for org my-org as"},
							[]string{"OK"},
						))
					})
				})
				It("fails if the org does not exist", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.All, errors.New("could not find org"))

					Expect(runCommand([]string{"-p", "service-plan", "-o", "not-findable-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"could not find org"},
					))
				})

				It("tells the user if the plan is already private", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.None, nil)

					Expect(runCommand([]string{"-p", "private-service-plan", "-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Plan", "of service", "is already inaccessible for all orgs"},
						[]string{"OK"},
					))
				})

				It("tells the user the use if the plan is being updated if the plan is limited", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.Limited, nil)

					Expect(runCommand([]string{"-p", "limited-service-plan", "-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Disabling access to plan limited-service-plan of service service for org my-org as"},
						[]string{"OK"},
					))
				})

				It("tells the user the plan is accessible to all orgs if the plan is public", func() {
					actor.UpdatePlanAndOrgForServiceReturns(actors.All, nil)

					Expect(runCommand([]string{"-p", "public-service-plan", "-o", "my-org", "service"})).To(BeTrue())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"The plan public-service-plan is accessible to all orgs."},
						[]string{"OK"},
					))
				})
			})
		})
	})
})
