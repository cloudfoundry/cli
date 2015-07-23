package securitygroup_test

import (
	"io/ioutil"
	"os"

	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-security-group command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-security-group", args, requirementsFactory, updateCommandDependency, false)
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

		It("fails with usage when a rules file is not provided", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("AWESOME_SECURITY_GROUP_NAME")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when the user is logged in", func() {
		var tempFile *os.File

		BeforeEach(func() {
			tempFile, _ = ioutil.TempFile("", "")
			requirementsFactory.LoginSuccess = true
		})

		AfterEach(func() {
			tempFile.Close()
			os.Remove(tempFile.Name())
		})

		JustBeforeEach(func() {
			runCommand("my-group", tempFile.Name())
		})

		Context("when the file specified has valid json", func() {
			BeforeEach(func() {
				tempFile.Write([]byte(`[{"protocol":"udp","ports":"8080-9090","destination":"198.41.191.47/1"}]`))
			})

			It("displays a message describing what its going to do", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating security group", "my-group", "my-user"},
					[]string{"OK"},
				))
			})

			It("creates the security group with those rules", func() {
				_, rules := securityGroupRepo.CreateArgsForCall(0)
				Expect(rules).To(Equal([]map[string]interface{}{
					{"protocol": "udp", "ports": "8080-9090", "destination": "198.41.191.47/1"},
				}))
			})

			Context("when the API returns an error", func() {
				Context("some sort of awful terrible error that we were not prescient enough to anticipate", func() {
					BeforeEach(func() {
						securityGroupRepo.CreateReturns(errors.New("Wops I failed"))
					})

					It("fails loudly", func() {
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Creating security group", "my-group"},
							[]string{"FAILED"},
						))
					})
				})

				Context("when the group already exists", func() {
					BeforeEach(func() {
						securityGroupRepo.CreateReturns(errors.NewHttpError(400, "300005", "The security group is taken: my-group"))
					})

					It("warns the user when group already exists", func() {
						Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
						Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"already exists"}))
					})
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
					[]string{"Incorrect json format: file:", tempFile.Name()},
					[]string{"Valid json file exampl"},
				))
			})
		})
	})
})
