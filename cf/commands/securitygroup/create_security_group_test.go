package securitygroup_test

import (
	"io/ioutil"
	"os"

	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *fakeSecurityGroup.FakeSecurityGroupRepo
		requirementsFactory *testreq.FakeReqFactory
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		cmd := NewCreateSecurityGroup(ui, configRepo, securityGroupRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("the-security-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates the security group", func() {
			runCommand("my-group")

			name, _ := securityGroupRepo.CreateArgsForCall(0)
			Expect(name).To(Equal("my-group"))
		})

		It("displays a message describing what its going to do", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating security group", "my-group", "my-user"},
				[]string{"OK"},
			))
		})

		Context("when the user specifies rules", func() {
			var tempFile *os.File

			BeforeEach(func() {
				tempFile, _ = ioutil.TempFile("", "")
			})

			AfterEach(func() {
				tempFile.Close()
				os.Remove(tempFile.Name())
			})

			JustBeforeEach(func() {
				runCommand("--json", tempFile.Name(), "security-groups-rule-everything-around-me")
			})

			Context("when the file specified has valid json", func() {
				BeforeEach(func() {
					tempFile.Write([]byte(`[{"protocol":"udp","ports":"8080-9090","destination":"198.41.191.47/1"}]`))
				})

				It("creates the security group with those rules, obviously", func() {
					_, rules := securityGroupRepo.CreateArgsForCall(0)
					Expect(rules).To(Equal([]map[string]interface{}{
						{"protocol": "udp", "ports": "8080-9090", "destination": "198.41.191.47/1"},
					}))
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

		Context("when the API returns an error", func() {
			Context("some sort of awful terrible error that we were not prescient enough to anticipate", func() {
				It("fails loudly", func() {
					securityGroupRepo.CreateReturns(errors.New("Wops I failed"))
					runCommand("my-group")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating security group", "my-group"},
						[]string{"FAILED"},
					))
				})
			})

			Context("when the group already exists", func() {
				It("warns the user when group already exists", func() {
					securityGroupRepo.CreateReturns(errors.NewHttpError(400, "300005", "The security group is taken: my-group"))
					runCommand("my-group")

					Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
					Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"already exists"}))
				})
			})
		})
	})
})
