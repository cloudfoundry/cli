package servicekey_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-service-key command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-service-key").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepository{}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "fake-service-instance-guid"
		serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
		serviceKeyRepo = testapi.NewFakeServiceKeyRepo()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	})

	var callDeleteServiceKey = func(args []string) bool {
		return testcmd.RunCliCommand("delete-service-key", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements are not satisfied", func() {
		It("fails when not logged in", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			Expect(callDeleteServiceKey([]string{"fake-service-key-name"})).To(BeFalse())
		})

		It("requires two arguments and one option to run", func() {
			Expect(callDeleteServiceKey([]string{})).To(BeFalse())
			Expect(callDeleteServiceKey([]string{"fake-arg-one"})).To(BeFalse())
			Expect(callDeleteServiceKey([]string{"fake-arg-one", "fake-arg-two", "fake-arg-three"})).To(BeFalse())
		})

		It("fails when space is not targetted", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		Context("deletes service key successfully", func() {
			BeforeEach(func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{
					Fields: models.ServiceKeyFields{
						Name:                "fake-service-key",
						Guid:                "fake-service-key-guid",
						Url:                 "fake-service-key-url",
						ServiceInstanceGuid: "fake-service-instance-guid",
						ServiceInstanceUrl:  "fake-service-instance-url",
					},
					Credentials: map[string]interface{}{
						"username": "fake-username",
						"password": "fake-password",
						"host":     "fake-host",
						"port":     "3306",
						"database": "fake-db-name",
						"uri":      "mysql://fake-user:fake-password@fake-host:3306/fake-db-name",
					},
				}
			})

			It("deletes service key successfully when '-f' option is provided", func() {
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key", "-f"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("deletes service key successfully when '-f' option is not provided and confirmed 'yes'", func() {
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
				ui.Inputs = append(ui.Inputs, "yes")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("skips to delete service key when '-f' option is not provided and confirmed 'no'", func() {
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
				ui.Inputs = append(ui.Inputs, "no")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs).To(BeEmpty())
			})

		})

		Context("deletes service key unsuccessful", func() {
			It("fails to delete service key when service instance does not exist", func() {
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.NewModelNotFoundError("Service instance", "non-exist-service-instance"))

				callDeleteServiceKey([]string{"non-exist-service-instance", "fake-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "non-exist-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service instance", "non-exist-service-instance", "does not exist."},
				))
			})

			It("fails to delete service key when the service key repository returns an error", func() {
				serviceKeyRepo.GetServiceKeyMethod.Error = errors.New("")
				callDeleteServiceKey([]string{"fake-service-instance", "non-exist-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service key", "non-exist-service-key", "does not exist for service instance", "fake-service-instance"},
				))
			})

			It("fails to delete service key when service key does not exist", func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{}
				callDeleteServiceKey([]string{"fake-service-instance", "non-exist-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service key", "non-exist-service-key", "does not exist for service instance", "fake-service-instance"},
				))
			})

			It("shows no service key is found", func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{}
				serviceKeyRepo.GetServiceKeyMethod.Error = &errors.NotAuthorizedError{}
				callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"No service key", "fake-service-key", "found for service instance", "fake-service-instance"},
				))
			})
		})
	})
})
