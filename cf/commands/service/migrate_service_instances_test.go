package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("migrating service instances from v1 to v2", func() {
	var (
		ui                  *testterm.FakeUI
		serviceRepo         *testapi.FakeServiceRepository
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		args                []string
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("migrate-service-instances").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepository()
		serviceRepo = &testapi.FakeServiceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		args = []string{}
	})

	Describe("requirements", func() {
		It("requires you to be logged in", func() {
			Expect(testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})

		It("requires five arguments to run", func() {
			requirementsFactory.LoginSuccess = true
			args = []string{"one", "two", "three"}

			Expect(testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})

		It("passes requirements if user is logged in and provided five args to run", func() {
			requirementsFactory.LoginSuccess = true
			args = []string{"one", "two", "three", "four", "five"}
			ui.Inputs = append(ui.Inputs, "no")

			Expect(testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)).To(BeTrue())
		})
	})

	Describe("migrating service instances", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			args = []string{"v1-service-label", "v1-provider-name", "v1-plan-name", "v2-service-label", "v2-plan-name"}
			serviceRepo.GetServiceInstanceCountForServicePlanReturns(1, nil)
		})

		It("displays the warning and the prompt including info about the instances and plan to migrate", func() {
			ui.Inputs = []string{""}
			testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

			Expect(ui.Outputs).To(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
			Expect(ui.Prompts).To(ContainSubstrings(
				[]string{"Really migrate", "1 service instance",
					"from plan", "v1-service-label", "v1-provider-name", "v1-plan-name",
					"to", "v2-service-label", "v2-plan-name"},
			))
		})

		Context("when the user confirms", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"yes"}
			})

			Context("when the v1 and v2 service instances exists", func() {
				BeforeEach(func() {
					var findServicePlanByDescriptionCallCount int
					serviceRepo.FindServicePlanByDescriptionStub = func(planDescription resources.ServicePlanDescription) (string, error) {
						findServicePlanByDescriptionCallCount++
						if findServicePlanByDescriptionCallCount == 1 {
							return "v1-guid", nil
						}
						return "v2-guid", nil
					}

					serviceRepo.MigrateServicePlanFromV1ToV2Returns(1, nil)
				})

				It("makes a request to migrate the v1 service instance", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					v1PlanGUID, v2PlanGUID := serviceRepo.MigrateServicePlanFromV1ToV2ArgsForCall(0)
					Expect(v1PlanGUID).To(Equal("v1-guid"))
					Expect(v2PlanGUID).To(Equal("v2-guid"))
				})

				It("finds the v1 service plan by its name, provider and service label", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					expectedV1 := resources.ServicePlanDescription{
						ServicePlanName: "v1-plan-name",
						ServiceProvider: "v1-provider-name",
						ServiceLabel:    "v1-service-label",
					}
					Expect(serviceRepo.FindServicePlanByDescriptionArgsForCall(0)).To(Equal(expectedV1))
				})

				It("finds the v2 service plan by its name and service label", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					expectedV2 := resources.ServicePlanDescription{
						ServicePlanName: "v2-plan-name",
						ServiceLabel:    "v2-service-label",
					}
					Expect(serviceRepo.FindServicePlanByDescriptionArgsForCall(1)).To(Equal(expectedV2))
				})

				It("notifies the user that the migration was successful", func() {
					serviceRepo.GetServiceInstanceCountForServicePlanReturns(2, nil)
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Attempting to migrate", "2", "service instances"},
						[]string{"1", "service instance", "migrated"},
						[]string{"OK"},
					))
				})
			})

			Context("when finding the v1 plan fails", func() {
				Context("because the plan does not exist", func() {
					BeforeEach(func() {
						serviceRepo.FindServicePlanByDescriptionReturns("", errors.NewModelNotFoundError("Service Plan", ""))
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Plan", "v1-service-label", "v1-provider-name", "v1-plan-name", "cannot be found"},
						))
					})

					It("does not display the warning", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
					})
				})

				Context("because there was an http error", func() {
					BeforeEach(func() {
						serviceRepo.FindServicePlanByDescriptionReturns("", errors.New("uh oh"))
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"uh oh"},
						))
					})

					It("does not display the warning", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
					})
				})
			})

			Context("when finding the v2 plan fails", func() {
				Context("because the plan does not exist", func() {
					BeforeEach(func() {
						var findServicePlanByDescriptionCallCount int
						serviceRepo.FindServicePlanByDescriptionStub = func(planDescription resources.ServicePlanDescription) (string, error) {
							findServicePlanByDescriptionCallCount++
							if findServicePlanByDescriptionCallCount == 1 {
								return "", nil
							}
							return "", errors.NewModelNotFoundError("Service Plan", "")
						}
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Plan", "v2-service-label", "v2-plan-name", "cannot be found"},
						))
					})

					It("does not display the warning", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
					})
				})

				Context("because there was an http error", func() {
					BeforeEach(func() {
						serviceRepo.FindServicePlanByDescriptionReturns("", errors.New("uh oh"))
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"uh oh"},
						))
					})

					It("does not display the warning", func() {
						testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

						Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
					})
				})
			})

			Context("when migrating the plans fails", func() {
				BeforeEach(func() {
					serviceRepo.MigrateServicePlanFromV1ToV2Returns(0, errors.New("ruh roh"))
				})

				It("notifies the user of the failure", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"ruh roh"},
					))
				})
			})

			Context("when there are no instances to migrate", func() {
				BeforeEach(func() {
					var findServicePlanByDescriptionCallCount int
					serviceRepo.FindServicePlanByDescriptionStub = func(planDescription resources.ServicePlanDescription) (string, error) {
						findServicePlanByDescriptionCallCount++
						if findServicePlanByDescriptionCallCount == 1 {
							return "v1-guid", nil
						}
						return "v2-guid", nil
					}
					serviceRepo.GetServiceInstanceCountForServicePlanReturns(0, nil)
				})

				It("returns a meaningful error", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"no service instances to migrate"},
					))
				})

				It("does not show the user the warning", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING:", "this operation is to replace a service broker"}))
				})
			})

			Context("when it cannot fetch the number of instances", func() {
				BeforeEach(func() {
					serviceRepo.GetServiceInstanceCountForServicePlanReturns(0, errors.New("service instance fetch is very bad"))
				})

				It("notifies the user of the failure", func() {
					testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"service instance fetch is very bad"},
					))
				})
			})
		})

		Context("when the user does not confirm", func() {
			BeforeEach(func() {
				ui.Inputs = append(ui.Inputs, "no")
			})

			It("does not continue the migration", func() {
				testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Migrating"}))
				Expect(serviceRepo.MigrateServicePlanFromV1ToV2CallCount()).To(BeZero())
			})
		})

		Context("when the user ignores confirmation using the force flag", func() {
			It("does not prompt the user for confirmation", func() {
				args = []string{"-f", "v1-service-label", "v1-provider-name", "v1-plan-name", "v2-service-label", "v2-plan-name"}

				testcmd.RunCliCommand("migrate-service-instances", args, requirementsFactory, updateCommandDependency, false)

				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Really migrate"}))
				Expect(serviceRepo.MigrateServicePlanFromV1ToV2CallCount()).To(Equal(1))
			})
		})
	})
})
