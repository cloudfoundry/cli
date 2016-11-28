package servicekey_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
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

var _ = Describe("service-key command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("service-key").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = new(apifakes.FakeServiceRepository)
		serviceInstance := models.ServiceInstance{}
		serviceInstance.GUID = "fake-service-instance-guid"
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

	var callGetServiceKey = func(args []string) bool {
		return testcmd.RunCLICommand("service-key", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(callGetServiceKey([]string{"fake-service-key-name"})).To(BeFalse())
		})

		It("requires two arguments to run", func() {
			Expect(callGetServiceKey([]string{})).To(BeFalse())
			Expect(callGetServiceKey([]string{"fake-arg-one"})).To(BeFalse())
			// This assertion is being skipped because the proper way to test this would be to refactor the existing integration style tests to real unit tests. We are going to rewrite the command and the tests in the refactor track anyways.
			// Expect(callGetServiceKey([]string{"fake-arg-one", "fake-arg-two"})).To(BeTrue())
			Expect(callGetServiceKey([]string{"fake-arg-one", "fake-arg-two", "fake-arg-three"})).To(BeFalse())
		})

		It("fails when service instance is not found", func() {
			serviceInstanceReq := new(requirementsfakes.FakeServiceInstanceRequirement)
			serviceInstanceReq.ExecuteReturns(errors.New("no service instance"))
			requirementsFactory.NewServiceInstanceRequirementReturns(serviceInstanceReq)
			Expect(callGetServiceKey([]string{"non-exist-service-instance"})).To(BeFalse())
		})

		It("fails when space is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "no targeted space"})
			Expect(callGetServiceKey([]string{"fake-service-instance", "fake-service-key-name"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		Context("gets service key successfully", func() {
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

			It("gets service credential", func() {
				callGetServiceKey([]string{"fake-service-instance", "fake-service-key"})
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"username", "fake-username"},
					[]string{"password", "fake-password"},
					[]string{"host", "fake-host"},
					[]string{"port", "3306"},
					[]string{"database", "fake-db-name"},
					[]string{"uri", "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"},
				))
				Expect(ui.Outputs()[1]).To(BeEmpty())
				Expect(serviceKeyRepo.GetServiceKeyMethod.InstanceGUID).To(Equal("fake-service-instance-guid"))
			})

			It("gets service guid when '--guid' flag is provided", func() {
				callGetServiceKey([]string{"--guid", "fake-service-instance", "fake-service-key"})

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"fake-service-key-guid"}))
				Expect(ui.Outputs()).ToNot(ContainSubstrings(
					[]string{"Getting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
				))
			})
		})

		Context("when service key does not exist", func() {
			It("shows no service key is found", func() {
				callGetServiceKey([]string{"fake-service-instance", "non-exist-service-key"})
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"No service key", "non-exist-service-key", "found for service instance", "fake-service-instance"},
				))
			})

			It("returns the empty string as guid when '--guid' flag is provided", func() {
				callGetServiceKey([]string{"--guid", "fake-service-instance", "non-exist-service-key"})

				Expect(len(ui.Outputs())).To(Equal(1))
				Expect(ui.Outputs()[0]).To(BeEmpty())
			})
		})

		Context("when api returned NotAuthorizedError", func() {
			It("shows no service key is found", func() {
				serviceKeyRepo.GetServiceKeyMethod.ServiceKey = models.ServiceKey{}
				serviceKeyRepo.GetServiceKeyMethod.Error = &errors.NotAuthorizedError{}

				callGetServiceKey([]string{"fake-service-instance", "fake-service-key"})
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"No service key", "fake-service-key", "found for service instance", "fake-service-instance"},
				))
			})
		})
	})
})
