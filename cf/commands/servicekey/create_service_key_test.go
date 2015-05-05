package servicekey_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/servicekey"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-service-key command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		cmd                 CreateServiceKey
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		serviceKeyRepo      *testapi.FakeServiceKeyRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepo{}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "fake-instance-guid"
		serviceRepo.FindInstanceByNameMap = generic.NewMap()
		serviceRepo.FindInstanceByNameMap.Set("fake-service-instance", serviceInstance)
		serviceKeyRepo = testapi.NewFakeServiceKeyRepo()
		cmd = NewCreateServiceKey(ui, config, serviceRepo, serviceKeyRepo)
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstanceNotFound: false}
	})

	var callCreateService = func(args []string) bool {
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			Expect(callCreateService([]string{"fake-service-instance", "fake-service-key"})).To(BeFalse())
		})

		It("requires two arguments to run", func() {
			Expect(callCreateService([]string{})).To(BeFalse())
			Expect(callCreateService([]string{"fake-arg-one"})).To(BeFalse())
			Expect(callCreateService([]string{"fake-arg-one", "fake-arg-two", "fake-arg-three"})).To(BeFalse())
		})

		It("fails when service instance is not found", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: true}
			Expect(callCreateService([]string{"non-exist-service-instance", "fake-service-key"})).To(BeFalse())
		})

		It("fails when space is not targetted", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			Expect(callCreateService([]string{"non-exist-service-instance", "fake-service-key"})).To(BeFalse())
		})
	})

	Describe("requiremnts are satisfied", func() {
		It("create service key successfully", func() {
			callCreateService([]string{"fake-service-instance", "fake-service-key"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"OK"},
			))
			Expect(serviceKeyRepo.CreateServiceKeyMethod.InstanceGuid).To(Equal("fake-instance-guid"))
			Expect(serviceKeyRepo.CreateServiceKeyMethod.KeyName).To(Equal("fake-service-key"))
		})

		It("create service key failed when the service key already exists", func() {
			serviceKeyRepo.CreateServiceKeyMethod.Error = errors.NewModelAlreadyExistsError("Service key", "exist-service-key")

			callCreateService([]string{"fake-service-instance", "exist-service-key"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service key", "exist-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"OK"},
				[]string{"Service key exist-service-key already exists"}))
		})

		It("create service key failed when the service is unbindable", func() {
			serviceKeyRepo.CreateServiceKeyMethod.Error = errors.NewUnbindableServiceError()
			callCreateService([]string{"fake-service-instance", "exist-service-key"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service key", "exist-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
				[]string{"FAILED"},
				[]string{"This service doesn't support creation of keys."}))
		})
	})
})
