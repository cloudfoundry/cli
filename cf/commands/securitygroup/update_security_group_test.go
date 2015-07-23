package securitygroup_test

import (
	"io/ioutil"
	"os"

	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("update-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *fakeSecurityGroup.FakeSecurityGroupRepo
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(securityGroupRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("update-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("update-security-group", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			Expect(runCommand("the-security-group")).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails with usage when a file path is not provided", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("my-group-name")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when the user is logged in", func() {
		var tempFile *os.File
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			securityGroup := models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name: "my-group-name",
					Guid: "my-group-guid",
				},
			}
			securityGroupRepo.ReadReturns(securityGroup, nil)
			tempFile, _ = ioutil.TempFile("", "")
		})

		AfterEach(func() {
			tempFile.Close()
			os.Remove(tempFile.Name())
		})

		JustBeforeEach(func() {
			runCommand("my-group-name", tempFile.Name())
		})

		Context("when the file specified has valid json", func() {
			BeforeEach(func() {
				tempFile.Write([]byte(`[{"protocol":"udp","port":"8080-9090","destination":"198.41.191.47/1"}]`))
			})

			It("displays a message describing what its going to do", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating security group", "my-group-name", "my-user"},
					[]string{"OK"},
					[]string{"TIP: Changes will not apply to existing running applications until they are restarted."},
				))
			})

			It("updates the security group with those rules, obviously", func() {
				jsonData := []map[string]interface{}{
					{"protocol": "udp", "port": "8080-9090", "destination": "198.41.191.47/1"},
				}

				_, jsonArg := securityGroupRepo.UpdateArgsForCall(0)

				Expect(jsonArg).To(Equal(jsonData))
			})

			Context("when the API returns an error", func() {
				Context("some sort of awful terrible error that we were not prescient enough to anticipate", func() {
					BeforeEach(func() {
						securityGroupRepo.UpdateReturns(errors.New("Wops I failed"))
					})

					It("fails loudly", func() {
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Updating security group", "my-group-name"},
							[]string{"FAILED"},
						))
					})
				})
			})

			Context("when the file specified has invalid json", func() {
				BeforeEach(func() {
					tempFile.Write([]byte(`[{noquote: thiswontwork}]`))
				})

				It("freaks out", func() {
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
					))
				})
			})
		})
	})
})
