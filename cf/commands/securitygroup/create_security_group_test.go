package securitygroup_test

import (
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
		securityGroupRepo   *fakeSecurityGroup.FakeSecurityGroup
		requirementsFactory *testreq.FakeReqFactory
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroup{}
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

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("the-security-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
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
			Expect(securityGroupRepo.CreateArgsForCall(0).Name).To(Equal("my-group"))
		})

		It("displays a message describing what its going to do", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating security group", "my-group", "my-user"},
				[]string{"OK"},
			))
		})

		It("allows the user to specify rules", func() {
			runCommand(
				"-rules",
				"[{\"protocol\":\"udp\",\"port\":\"8080-9090\",\"destination\":\"198.41.191.47/1\"}]",
				"security-groups-rule-everything-around-me",
			)

			Expect(securityGroupRepo.CreateArgsForCall(0).Rules).To(Equal([]map[string]string{
				{"protocol": "udp", "port": "8080-9090", "destination": "198.41.191.47/1"},
			}))
		})

		It("freaks out if the user specifies a rule incorrectly", func() {
			runCommand(
				"-rules",
				"Im so not right",
				"security group",
			)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
			))
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
