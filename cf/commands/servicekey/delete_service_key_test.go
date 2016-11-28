package servicekey_test

import (
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-service-key command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-service-key").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &apifakes.FakeServiceRepository{}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.GUID = "fake-service-instance-guid"
		serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
		serviceKeyRepo = apifakes.NewFakeServiceKeyRepo()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
	})

	var callDeleteServiceKey = func(args []string) bool {
		return testcmd.RunCLICommand("delete-service-key", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements are not satisfied", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(callDeleteServiceKey([]string{"fake-service-key-name"})).To(BeFalse())
		})

		It("requires two arguments and one option to run", func() {
			Expect(callDeleteServiceKey([]string{})).To(BeFalse())
			Expect(callDeleteServiceKey([]string{"fake-arg-one"})).To(BeFalse())
			Expect(callDeleteServiceKey([]string{"fake-arg-one", "fake-arg-two", "fake-arg-three"})).To(BeFalse())
		})

		It("fails when space is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "no targeted space"})
			Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		Context("deletes service key successfully", func() {
			BeforeEach(func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{
					Fields: models.ServiceKeyFields{
						Name:                "fake-service-key",
						GUID:                "fake-service-key-guid",
						URL:                 "fake-service-key-url",
						ServiceInstanceGUID: "fake-service-instance-guid",
						ServiceInstanceURL:  "fake-service-instance-url",
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
				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key", "-f"})).To(BeTrue())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("deletes service key successfully when '-f' option is not provided and confirmed 'yes'", func() {
				ui.Inputs = append(ui.Inputs, "yes")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("skips to delete service key when '-f' option is not provided and confirmed 'no'", func() {
				ui.Inputs = append(ui.Inputs, "no")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs()).To(BeEmpty())
			})

		})

		Context("deletes service key unsuccessful", func() {
			It("fails to delete service key when service instance does not exist", func() {
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.NewModelNotFoundError("Service instance", "non-exist-service-instance"))

				callDeleteServiceKey([]string{"non-exist-service-instance", "fake-service-key", "-f"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "non-exist-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service instance", "non-exist-service-instance", "does not exist."},
				))
			})

			It("fails to delete service key when the service key repository returns an error", func() {
				serviceKeyRepo.GetServiceKeyMethod.Error = errors.New("")
				callDeleteServiceKey([]string{"fake-service-instance", "non-exist-service-key", "-f"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service key", "non-exist-service-key", "does not exist for service instance", "fake-service-instance"},
				))
			})

			It("fails to delete service key when service key does not exist", func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{}
				callDeleteServiceKey([]string{"fake-service-instance", "non-exist-service-key", "-f"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service key", "non-exist-service-key", "does not exist for service instance", "fake-service-instance"},
				))
			})

			It("shows no service key is found", func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{}
				serviceKeyRepo.GetServiceKeyMethod.Error = &errors.NotAuthorizedError{}
				callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key", "-f"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"No service key", "fake-service-key", "found for service instance", "fake-service-instance"},
				))
			})
		})
	})
})
