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

var _ = Describe("delete-service-key command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		cmd                 DeleteServiceKey
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
		serviceRepo.FindInstanceByNameMap = generic.NewMap()
		serviceRepo.FindInstanceByNameMap.Set("fake-service-instance", serviceInstance)
		serviceKeyRepo = testapi.NewFakeServiceKeyRepo()
		cmd = NewDeleteServiceKey(ui, config, serviceRepo, serviceKeyRepo)
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstanceNotFound: false}
	})

	var callDeleteServiceKey = func(args []string) bool {
		return testcmd.RunCommand(cmd, args, requirementsFactory)
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

		It("fails when service instance is not found", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: true}
			Expect(callDeleteServiceKey([]string{"non-exist-service-instance"})).To(BeFalse())
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
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: false, TargetedSpaceSuccess: true}

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key", "-f"})).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("deletes service key successfully when '-f' option is not provided and confirmed 'yes'", func() {
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: false, TargetedSpaceSuccess: true}
				ui.Inputs = append(ui.Inputs, "yes")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "fake-service-instance", "as", "my-user"},
					[]string{"OK"}))
			})

			It("skips to delete service key when '-f' option is not provided and confirmed 'no'", func() {
				requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, ServiceInstanceNotFound: false, TargetedSpaceSuccess: true}
				ui.Inputs = append(ui.Inputs, "no")

				Expect(callDeleteServiceKey([]string{"fake-service-instance", "fake-service-key"})).To(BeTrue())
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service key", "fake-service-key"}))
				Expect(ui.Outputs).To(BeEmpty())
			})

		})

		Context("deletes service key unsuccessful", func() {
			It("fails to delete service key when service instance does not exist", func() {
				serviceRepo.FindInstanceByNameNotFound = true
				callDeleteServiceKey([]string{"non-exist-service-instance", "fake-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "fake-service-key", "for service instance", "non-exist-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service instance", "non-exist-service-instance", "does not exist."},
				))
			})

			It("fails to delete service key when service key does not exist", func() {
				serviceKeyRepo.GetServiceKeyMethod.Error = errors.New("")
				callDeleteServiceKey([]string{"fake-service-instance", "non-exist-service-key", "-f"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting key", "non-exist-service-key", "for service instance", "fake-service-instance", "as", "my-user..."},
					[]string{"OK"},
					[]string{"Service key", "non-exist-service-key", "does not exist for service instance", "fake-service-instance"},
				))
			})
		})
	})
})
