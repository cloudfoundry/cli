package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Marketplace Services", func() {
	var ui *testterm.FakeUI
	var requirementsFactory *testreq.FakeReqFactory
	var config configuration.ReadWriter
	var serviceRepo *testapi.FakeServiceRepo
	var fakeServiceOfferings []models.ServiceOffering

	BeforeEach(func() {
		serviceRepo = &testapi.FakeServiceRepo{}
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{ApiEndpointSuccess: true}

		fakeServiceOfferings = []models.ServiceOffering{
			models.ServiceOffering{
				Plans: []models.ServicePlanFields{
					models.ServicePlanFields{Name: "service-plan-a"},
					models.ServicePlanFields{Name: "service-plan-b"},
				},
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label:       "zzz-my-service-offering",
					Description: "service offering 1 description",
				}},
			models.ServiceOffering{
				Plans: []models.ServicePlanFields{
					models.ServicePlanFields{Name: "service-plan-c"},
					models.ServicePlanFields{Name: "service-plan-d"}},
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label:       "aaa-my-service-offering",
					Description: "service offering 2 description",
				},
			}}
	})

	Context("when the an API endpoint is not targeted", func() {
		It("does not meet its requirements", func() {
			config := testconfig.NewRepository()
			cmd := NewMarketplaceServices(ui, config, serviceRepo)
			requirementsFactory.ApiEndpointSuccess = false

			testcmd.RunCommand(cmd, testcmd.NewContext("marketplace", []string{}), requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			config = testconfig.NewRepositoryWithDefaults()
		})

		Context("when the user has a space targeted", func() {
			BeforeEach(func() {
				config.SetSpaceFields(models.SpaceFields{
					Guid: "the-space-guid",
					Name: "the-space-name",
				})
			})

			It("lists all of the service offerings for the space", func() {
				serviceRepo := &testapi.FakeServiceRepo{}
				serviceRepo.GetServiceOfferingsForSpaceReturns.ServiceOfferings = fakeServiceOfferings
				cmd := NewMarketplaceServices(ui, config, serviceRepo)
				testcmd.RunCommand(cmd, testcmd.NewContext("marketplace", []string{}), requirementsFactory)

				Expect(serviceRepo.GetServiceOfferingsForSpaceArgs.SpaceGuid).To(Equal("the-space-guid"))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Getting services from marketplace in org", "my-org", "the-space-name", "my-user"},
					{"OK"},
					{"service", "plans", "description"},
					{"aaa-my-service-offering", "service offering 2 description", "service-plan-c", "service-plan-d"},
					{"zzz-my-service-offering", "service offering 1 description", "service-plan-a", "service-plan-b"},
				})
			})
		})

		Context("when the user doesn't have a space targeted", func() {
			BeforeEach(func() {
				config.SetSpaceFields(models.SpaceFields{})
			})

			It("tells the user to target a space", func() {
				cmd := NewMarketplaceServices(ui, config, serviceRepo)
				testcmd.RunCommand(cmd, testcmd.NewContext("marketplace", []string{}), requirementsFactory)
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"without", "space"},
				})
			})
		})
	})

	Context("when user is not logged in", func() {
		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("lists all public service offerings if any are available", func() {
			serviceRepo := &testapi.FakeServiceRepo{}
			serviceRepo.GetAllServiceOfferingsReturns.ServiceOfferings = fakeServiceOfferings

			cmd := NewMarketplaceServices(ui, config, serviceRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("marketplace", []string{}), requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Getting all services from marketplace"},
				{"OK"},
				{"service", "plans", "description"},
				{"aaa-my-service-offering", "service offering 2 description", "service-plan-c", "service-plan-d"},
				{"zzz-my-service-offering", "service offering 1 description", "service-plan-a", "service-plan-b"},
			})
		})

		It("does not display a table if no service offerings exist", func() {
			serviceRepo := &testapi.FakeServiceRepo{}
			serviceRepo.GetAllServiceOfferingsReturns.ServiceOfferings = []models.ServiceOffering{}

			cmd := NewMarketplaceServices(ui, config, serviceRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("marketplace", []string{}), requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"No service offerings found"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"service", "plans", "description"},
			})
		})
	})
})
