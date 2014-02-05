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
			configRepo := testconfig.NewRepository()
			serviceRepo = &testapi.FakeServiceRepo{}
			cmd = NewMigrateServiceInstances(ui, configRepo, serviceRepo)
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			args = []string{}
		})

		Describe("requirements", func() {
			It("should require you to be logged in", func() {
				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})

			It("should require you to be logged in and be provided five args to run", func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"one", "two", "three", "four", "five"}
				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			})

			It("should require five arguments to run", func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"one", "two", "three"}
				context = testcmd.NewContext("migrate-service-instances", args)
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})
		})

		Describe("migrating service instances", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = true
				args = []string{"v1-service-name", "v1-provider-name", "v1-plan-name", "v2-service-name", "v2-plan-name"}
				context = testcmd.NewContext("migrate-service-instances", args)
			})

			It("makes a request to migrate the v1 service instance", func() {
				serviceRepo.V1FoundGuid = "v1-guid"
				serviceRepo.V2FoundGuid = "v2-guid"
				testcmd.RunCommand(cmd, context, requirementsFactory)

				expectedV1 := api.V1ServicePlanDescription{
					ServicePlanName: "v1-plan-name",
					ServiceProvider: "v1-provider-name",
					ServiceName:     "v1-service-name",
				}

				expectedV2 := api.V2ServicePlanDescription{
					ServicePlanName: "v2-plan-name",
					ServiceName:     "v2-service-name",
				}
				Expect(serviceRepo.V1ServicePlanDescription).To(Equal(expectedV1))
				Expect(serviceRepo.V2ServicePlanDescription).To(Equal(expectedV2))
				Expect(serviceRepo.V1GuidToMigrate).To(Equal("v1-guid"))
				Expect(serviceRepo.V2GuidToMigrate).To(Equal("v2-guid"))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Migrating", "v1-plan-name", "v1-service-name", "v1-provider-name", "v2-plan-name", "v2-service-name"},
					{"OK"},
				})
			})

			It("warns the user when finding the plans fails", func() {
				serviceRepo.FindServicePlanToMigrateByDescriptionResponse = net.NewApiResponseWithMessage("uh oh")
				testcmd.RunCommand(cmd, context, requirementsFactory)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"FAILED"},
					{"uh oh"},
				})
			})

			It("warns the user when migrating the plans fails", func() {
				serviceRepo.MigrateServicePlanFromV1ToV2Response = net.NewApiResponseWithMessage("ruh roh")
				testcmd.RunCommand(cmd, context, requirementsFactory)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"FAILED"},
					{"ruh roh"},
				})
			})
		})
	})
}
