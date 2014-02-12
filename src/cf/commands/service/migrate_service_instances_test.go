package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"cf/net"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("migrating service instances from v1 to v2", func() {
		var (
			ui                  *testterm.FakeUI
			serviceRepo         *testapi.FakeServiceRepo
			cmd                 *MigrateServiceInstances
			requirementsFactory *testreq.FakeReqFactory
			context             *cli.Context
			args                []string
		)

		BeforeEach(func() {
			ui = &testterm.FakeUI{}
			config := testconfig.NewRepository()
			serviceRepo = &testapi.FakeServiceRepo{}
			cmd = NewMigrateServiceInstances(ui, config, serviceRepo)
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			args = []string{}
		})

		Describe("requirements", func() {
			It("requires you to be logged in", func() {
				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})

			It("requires five arguments to run", func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"one", "two", "three"}
				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})

			It("passes requirements if user is logged in and provided five args to run", func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"one", "two", "three", "four", "five"}
				ui.Inputs = append(ui.Inputs, "no")

				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			})
		})

		Describe("migrating service instances", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"v1-service-name", "v1-provider-name", "v1-plan-name", "v2-service-name", "v2-plan-name"}
				context = testcmd.NewContext("migrate-service-instances", args)
				serviceRepo.ServiceInstanceCountForServicePlan = 1
			})

			It("displays the warning and the prompt including info about the instances and plan to migrate", func() {
				ui.Inputs = []string{""}
				testcmd.RunCommand(cmd, context, requirementsFactory)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"WARNING:", "this operation is to replace a service broker"},
				})
				testassert.SliceContains(ui.Prompts, testassert.Lines{
					{"Really migrate", "1 service instance",
						"from plan", "v1-service-name", "v1-provider-name", "v1-plan-name",
						"to", "v2-service-name", "v2-plan-name"},
				})
			})

			Context("when the user confirms", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"yes"}
				})

				Context("when the v1 and v2 service instances exists", func() {
					BeforeEach(func() {
						serviceRepo.V1FoundGuid = "v1-guid"
						serviceRepo.V2FoundGuid = "v2-guid"
					})

					It("makes a request to migrate the v1 service instance", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						Expect(serviceRepo.V1GuidToMigrate).To(Equal("v1-guid"))
						Expect(serviceRepo.V2GuidToMigrate).To(Equal("v2-guid"))
					})

					It("finds the v1 service plan by its name, provider and service label", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						expectedV1 := api.V1ServicePlanDescription{
							ServicePlanName: "v1-plan-name",
							ServiceProvider: "v1-provider-name",
							ServiceName:     "v1-service-name",
						}
						Expect(serviceRepo.V1ServicePlanDescription).To(Equal(expectedV1))
					})

					It("finds the v2 service plan by its name and service label", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						expectedV2 := api.V2ServicePlanDescription{
							ServicePlanName: "v2-plan-name",
							ServiceName:     "v2-service-name",
						}
						Expect(serviceRepo.V2ServicePlanDescription).To(Equal(expectedV2))
					})

					It("notifies the user that the migration was successful", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceContains(ui.Outputs, testassert.Lines{
							{"Migrating", "1", "service instance"},
							{"OK"},
						})
					})
				})

				Context("when finding the plan fails", func() {
					BeforeEach(func() {
						serviceRepo.FindServicePlanToMigrateByDescriptionResponse = net.NewApiResponseWithMessage("uh oh")
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceContains(ui.Outputs, testassert.Lines{
							{"FAILED"},
							{"uh oh"},
						})
					})

					It("does not display the warning", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
							{"WARNING:", "this operation is to replace a service broker"},
						})
					})
				})

				Context("when migrating the plans fails", func() {
					BeforeEach(func() {
						serviceRepo.MigrateServicePlanFromV1ToV2Response = net.NewApiResponseWithMessage("ruh roh")
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceContains(ui.Outputs, testassert.Lines{
							{"FAILED"},
							{"ruh roh"},
						})
					})
				})

				Context("when there are no instances to migrate", func() {
					BeforeEach(func() {
						serviceRepo.V1FoundGuid = "v1-guid"
						serviceRepo.V2FoundGuid = "v2-guid"
						serviceRepo.ServiceInstanceCountForServicePlan = 0
					})

					It("returns a meaningful error", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceContains(ui.Outputs, testassert.Lines{
							{"FAILED"},
							{"no service instances to migrate"},
						})
					})

					It("does not show the user the warning", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
							{"WARNING:", "this operation is to replace a service broker"},
						})
					})
				})

				Context("when it cannot fetch the number of instances", func() {
					BeforeEach(func() {
						serviceRepo.ServiceInstanceCountApiResponse = net.NewApiResponseWithMessage("service instance fetch is very bad")
					})

					It("notifies the user of the failure", func() {
						testcmd.RunCommand(cmd, context, requirementsFactory)

						testassert.SliceContains(ui.Outputs, testassert.Lines{
							{"FAILED"},
							{"service instance fetch is very bad"},
						})
					})
				})
			})

			Context("when the user does not confirm", func() {
				BeforeEach(func() {
					ui.Inputs = append(ui.Inputs, "no")
				})

				It("does not continue the migration", func() {
					testcmd.RunCommand(cmd, context, requirementsFactory)

					testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{{"Migrating"}})
					Expect(serviceRepo.MigrateServicePlanFromV1ToV2Called).To(BeFalse())
				})
			})
		})
	})
}
