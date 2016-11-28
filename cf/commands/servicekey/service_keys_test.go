package servicekey_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-keys command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		serviceRepo         *apifakes.FakeServiceRepository
		serviceKeyRepo      *apifakes.OldFakeServiceKeyRepo
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.RepoLocator = deps.RepoLocator.SetServiceKeyRepository(serviceKeyRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("service-keys").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = new(apifakes.FakeServiceRepository)
		serviceInstance := models.ServiceInstance{}
		serviceInstance.GUID = "fake-instance-guid"
		serviceInstance.Name = "fake-service-instance"
		serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
		serviceKeyRepo = apifakes.NewFakeServiceKeyRepo()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		serviceInstanceReq := new(requirementsfakes.FakeServiceInstanceRequirement)
		requirementsFactory.NewServiceInstanceRequirementReturns(serviceInstanceReq)
		serviceInstanceReq.GetServiceInstanceReturns(serviceInstance)
	})

	var callListServiceKeys = func(args []string) bool {
		return testcmd.RunCLICommand("service-keys", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(callListServiceKeys([]string{"fake-service-instance", "fake-service-key"})).To(BeFalse())
		})

		It("requires one argument to run", func() {
			Expect(callListServiceKeys([]string{})).To(BeFalse())
			Expect(callListServiceKeys([]string{"fake-arg-one"})).To(BeTrue())
			Expect(callListServiceKeys([]string{"fake-arg-one", "fake-arg-two"})).To(BeFalse())
		})

		It("fails when service instance is not found", func() {
			serviceInstanceReq := new(requirementsfakes.FakeServiceInstanceRequirement)
			serviceInstanceReq.ExecuteReturns(errors.New("no service instance"))
			requirementsFactory.NewServiceInstanceRequirementReturns(serviceInstanceReq)
			Expect(callListServiceKeys([]string{"non-exist-service-instance"})).To(BeFalse())
		})

		It("fails when space is not targetted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "no targeted space"})
			Expect(callListServiceKeys([]string{"non-exist-service-instance"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		It("list service keys successfully", func() {
			serviceKeyRepo.ListServiceKeysMethod.ServiceKeys = []models.ServiceKey{
				{
					Fields: models.ServiceKeyFields{
						Name: "fake-service-key-1",
					},
				},
				{
					Fields: models.ServiceKeyFields{
						Name: "fake-service-key-2",
					},
				},
			}
			callListServiceKeys([]string{"fake-service-instance"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting keys for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"name"},
				[]string{"fake-service-key-1"},
				[]string{"fake-service-key-2"},
			))
			Expect(ui.Outputs()[1]).To(BeEmpty())
			Expect(serviceKeyRepo.ListServiceKeysMethod.InstanceGUID).To(Equal("fake-instance-guid"))
		})

		It("does not list service keys when none are returned", func() {
			callListServiceKeys([]string{"fake-service-instance"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting keys for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"No service key for service instance", "fake-service-instance"},
			))
		})
	})
})
