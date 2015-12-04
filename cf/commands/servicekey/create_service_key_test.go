package servicekey_test

import (
	"io/ioutil"
	"os"

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

var _ = Describe("create-service-key command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-service-key").SetDependency(deps, pluginCall))
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

	var callCreateService = func(args []string) bool {
		return testcmd.RunCliCommand("create-service-key", args, requirementsFactory, updateCommandDependency, false)
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

	Context("when passing arbitrary params", func() {
		Context("as a json string", func() {
			It("successfully creates a service key and passes the params as a json string", func() {
				callCreateService([]string{"fake-service-instance", "fake-service-key", "-c", `{"foo": "bar"}`})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating service key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceKeyRepo.CreateServiceKeyMethod.InstanceGuid).To(Equal("fake-instance-guid"))
				Expect(serviceKeyRepo.CreateServiceKeyMethod.KeyName).To(Equal("fake-service-key"))
				Expect(serviceKeyRepo.CreateServiceKeyMethod.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})
		})

		Context("that are not valid json", func() {
			It("returns an error to the UI", func() {
				callCreateService([]string{"fake-service-instance", "fake-service-key", "-c", `bad-json`})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
				))
			})
		})
		Context("as a file that contains json", func() {
			var jsonFile *os.File
			var params string

			BeforeEach(func() {
				params = "{\"foo\": \"bar\"}"
			})

			AfterEach(func() {
				if jsonFile != nil {
					jsonFile.Close()
					os.Remove(jsonFile.Name())
				}
			})

			JustBeforeEach(func() {
				var err error
				jsonFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				err = ioutil.WriteFile(jsonFile.Name(), []byte(params), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("successfully creates a service key and passes the params as a json", func() {
				callCreateService([]string{"fake-service-instance", "fake-service-key", "-c", jsonFile.Name()})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating service key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceKeyRepo.CreateServiceKeyMethod.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				BeforeEach(func() {
					params = "bad-json"
				})

				It("returns an error to the UI", func() {
					callCreateService([]string{"fake-service-instance", "fake-service-key", "-c", jsonFile.Name()})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
					))
				})
			})
		})
	})
})
