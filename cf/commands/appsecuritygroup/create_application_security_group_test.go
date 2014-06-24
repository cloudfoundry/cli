package appsecuritygroup_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/appsecuritygroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-app-security-group", func() {
	var (
		ui                   *testterm.FakeUI
		appSecurityGroupRepo *testapi.FakeAppSecurityGroup
		requirementsFactory  *testreq.FakeReqFactory
		spaceRepo            *testapi.FakeSpaceRepository
		configRepo           configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		appSecurityGroupRepo = &testapi.FakeAppSecurityGroup{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		space := models.Space{}
		space.Guid = "space-guid-1"
		space.Name = "space-1"
		space2 := models.Space{}
		space2.Guid = "space-guid-2"
		space2.Name = "space-2"

		spaceRepo = &testapi.FakeSpaceRepository{
			Spaces: []models.Space{space, space2},
		}
	})

	runCommand := func(args ...string) {
		cmd := NewCreateSecurityGroup(ui, configRepo, appSecurityGroupRepo, spaceRepo)
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

		It("creates the application security group", func() {
			runCommand("my-group")
			Expect(appSecurityGroupRepo.CreateArgsForCall(0).Name).To(Equal("my-group"))
		})

		It("displays a message describing what its going to do", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(matchers.ContainSubstrings(
				[]string{"Creating application security group", "my-group", "my-user"},
				[]string{"OK"},
			))
		})

		It("allows the user to specify rules", func() {
			runCommand(
				"-rules",
				"{\"protocol\":\"udp\",\"port\":\"8080-9090\",\"destination\":\"198.41.191.47/1\"}",
				"app-security-groups-rule-everything-around-me",
			)

			Expect(appSecurityGroupRepo.CreateArgsForCall(0).Rules).To(Equal([]map[string]string{
				{"protocol": "udp", "port": "8080-9090", "destination": "198.41.191.47/1"},
			}))
		})

		It("freaks out if the user specifies a rule incorrectly", func() {
			runCommand(
				"-rules",
				"Im so not right",
				"security group",
			)

			Expect(ui.Outputs).To(matchers.ContainSubstrings(
				[]string{"FAILED"},
			))
		})

		It("allows the user to specify multiple spaces", func() {
			runCommand(
				"-space",
				"space-1",
				"-space",
				"space-2",
				"Multi space security group",
			)

			Expect(appSecurityGroupRepo.CreateArgsForCall(0).SpaceGuids).To(Equal(
				[]string{"space-guid-1", "space-guid-2"},
			))
		})

		Context("when creating a security group returns an error", func() {
			It("alerts the user when creating the security group fails", func() {
				appSecurityGroupRepo.CreateReturns(errors.New("Wops I failed"))
				runCommand("my-group")

				Expect(ui.Outputs).To(matchers.ContainSubstrings(
					[]string{"reating application security group", "my-group"},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the user specifies a space that does not exist", func() {
			It("Fails", func() {
				runCommand(
					"-space",
					"space-does-not-exist",
					"why-would-this-ever-work",
				)

				Expect(ui.Outputs).To(matchers.ContainSubstrings(
					[]string{"FAILED"},
					[]string{"space-does-not-exist"},
				))
			})
		})
	})
})
