package servicekey_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
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

var _ = Describe("service-key command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		cmd                 *ServiceKey
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		serviceKeyRepo      *testapi.FakeServiceKeyRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepo{}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "fake-service-instance-guid"
		serviceInstance.Name = "fake-service-instance"
		serviceRepo.FindInstanceByNameMap = generic.NewMap()
		serviceRepo.FindInstanceByNameMap.Set("fake-service-instance", serviceInstance)
		serviceKeyRepo = testapi.NewFakeServiceKeyRepo()
		cmd = NewGetServiceKey(ui, config, serviceRepo, serviceKeyRepo)
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstanceNotFound: false}
		requirementsFactory.ServiceInstance = serviceInstance
	})

	var callGetServiceKey = func(args []string) bool {
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			Expect(callGetServiceKey([]string{"fake-service-key-name"})).To(BeFalse())
		})

		It("requires two arguments to run", func() {
			Expect(callGetServiceKey([]string{})).To(BeFalse())
			Expect(callGetServiceKey([]string{"fake-arg-one"})).To(BeFalse())
			Expect(callGetServiceKey([]string{"fake-arg-one", "fake-arg-two"})).To(BeTrue())
			Expect(callGetServiceKey([]string{"fake-arg-one", "fake-arg-two", "fake-arg-three"})).To(BeFalse())
		})

		It("fails when service instance is not found", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: true}
			Expect(callGetServiceKey([]string{"non-exist-service-instance"})).To(BeFalse())
		})

		It("fails when space is not targetted", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
			Expect(callGetServiceKey([]string{"fake-service-instance", "fake-service-key-name"})).To(BeFalse())
		})
	})

	Describe("requirements are satisfied", func() {
		Context("gets service key successfully", func() {
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

			It("gets service credential", func() {
				callGetServiceKey([]string{"fake-service-instance", "fake-service-key"})
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"username", "fake-username"},
					[]string{"password", "fake-password"},
					[]string{"host", "fake-host"},
					[]string{"port", "3306"},
					[]string{"database", "fake-db-name"},
					[]string{"uri", "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"},
				))
				Expect(ui.Outputs[1]).To(BeEmpty())
				Expect(serviceKeyRepo.GetServiceKeyMethod.InstanceGuid).To(Equal("fake-service-instance-guid"))
			})

			It("gets service guid when '--guid' flag is provided", func() {
				callGetServiceKey([]string{"--guid", "fake-service-instance", "fake-service-key"})

				Expect(ui.Outputs).To(ContainSubstrings([]string{"fake-service-key-guid"}))
				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"Getting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
				))
			})
		})

		Context("when service key does not exist", func() {
			It("shows no service key is found", func() {
				callGetServiceKey([]string{"fake-service-instance", "non-exist-service-key"})
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"No service key", "non-exist-service-key", "found for service instance", "fake-service-instance"},
				))
			})

			It("returns the empty string as guid when '--guid' flag is provided", func() {
				callGetServiceKey([]string{"--guid", "fake-service-instance", "non-exist-service-key"})

				Expect(len(ui.Outputs)).To(Equal(1))
				Expect(ui.Outputs[0]).To(BeEmpty())
			})
		})
	})
})
