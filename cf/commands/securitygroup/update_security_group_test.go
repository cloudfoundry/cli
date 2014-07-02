package securitygroup_test

import (
	"io/ioutil"
	"os"

	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("update-security-group command", func() {
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
		cmd := NewUpdateSecurityGroup(ui, configRepo, securityGroupRepo)
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
			securityGroup := models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name: "my-group-name",
					Guid: "my-group-guid",
				},
			}
			securityGroupRepo.ReadReturns(securityGroup, nil)
		})

		It("updates the security group", func() {
			runCommand("my-group")
			arg1, _ := securityGroupRepo.UpdateArgsForCall(0)
			Expect(arg1).To(Equal("my-group-guid"))
		})

		It("displays a message describing what its going to do", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating security group", "my-group", "my-user"},
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
					tempFile.Write([]byte(`[{"protocol":"udp","port":"8080-9090","destination":"198.41.191.47/1"}]`))
				})

				It("updates the security group with those rules, obviously", func() {
					jsonData := []map[string]interface{}{
						{"protocol": "udp", "port": "8080-9090", "destination": "198.41.191.47/1"},
					}

					_, jsonArg := securityGroupRepo.UpdateArgsForCall(0)

					Expect(jsonArg).To(Equal(jsonData))
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
					securityGroupRepo.UpdateReturns(errors.New("Wops I failed"))
					runCommand("my-group")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Updating security group", "my-group"},
						[]string{"FAILED"},
					))
				})
			})
		})
	})
})
