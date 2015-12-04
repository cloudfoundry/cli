package servicekey_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-keys command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepository
		serviceKeyRepo      *testapi.FakeServiceKeyRepo
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.RepoLocator = deps.RepoLocator.SetServiceKeyRepository(serviceKeyRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("service-keys").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepository{}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "fake-instance-guid"
		serviceInstance.Name = "fake-service-instance"
		serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
		serviceKeyRepo = testapi.NewFakeServiceKeyRepo()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstanceNotFound: false}
		requirementsFactory.ServiceInstance = serviceInstance
	})

	var callListServiceKeys = func(args []string) bool {
		return testcmd.RunCliCommand("service-keys", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			Expect(callListServiceKeys([]string{"fake-service-instance", "fake-service-key"})).To(BeFalse())
		})

		It("requires one argument to run", func() {
			Expect(callListServiceKeys([]string{})).To(BeFalse())
			Expect(callListServiceKeys([]string{"fake-arg-one"})).To(BeTrue())
			Expect(callListServiceKeys([]string{"fake-arg-one", "fake-arg-two"})).To(BeFalse())
		})

		It("fails when service instance is not found", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: true}
			Expect(callListServiceKeys([]string{"non-exist-service-instance"})).To(BeFalse())
		})

		It("fails when space is not targetted", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			Expect(callListServiceKeys([]string{"non-exist-service-instance"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		It("list service keys successfully", func() {
			serviceKeyRepo.ListServiceKeysMethod.ServiceKeys = []models.ServiceKey{
				models.ServiceKey{
					Fields: models.ServiceKeyFields{
						Name: "fake-service-key-1",
					},
				},
				models.ServiceKey{
					Fields: models.ServiceKeyFields{
						Name: "fake-service-key-2",
					},
				},
			}
			callListServiceKeys([]string{"fake-service-instance"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting keys for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"name"},
				[]string{"fake-service-key-1"},
				[]string{"fake-service-key-2"},
			))
			Expect(ui.Outputs[1]).To(BeEmpty())
			Expect(serviceKeyRepo.ListServiceKeysMethod.InstanceGuid).To(Equal("fake-instance-guid"))
		})

		It("does not list service keys when none are returned", func() {
			callListServiceKeys([]string{"fake-service-instance"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting keys for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"No service key for service instance", "fake-service-instance"},
			))
		})
	})
})
