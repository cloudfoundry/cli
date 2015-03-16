package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testOrgApi "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("user-info command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
		orgRepo             *testOrgApi.FakeOrganizationRepository
		spaceRepo           *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		orgRepo = &testOrgApi.FakeOrganizationRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(ShowUserInfo(ui, configRepo, orgRepo, spaceRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			Expect(runCommand()).To(BeFalse())
		})

		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when no org or space is targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = false

			Expect(runCommand()).To(BeFalse())
		})

	})

	Context("when logged in and targeted with proper orgs", func() {
		BeforeEach(func() {
			org := models.Organization{}
			org.Name = "Org1"
			org.Guid = "org1-guid"

			user := models.UserFields{}
			user.Username = "user1"

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.UserFields = user
			requirementsFactory.Organization = org

			configRepo = testconfig.NewRepositoryWithAccessToken(core_config.TokenInfo{Username: "user1"})

			configRepo.SetOrganizationFields(models.OrganizationFields{
				Name: "Org1",
				Guid: "org1-guid",
			})
		})

		It("shows the roles for orgs when only Org is targeted", func() {

			orgRepo.ListOfRolesOfAnOrg = map[string]string{
				"manager_guid":         "Org1",
				"billing_manager_guid": "Org1",
				"auditor_guid":         "Org1",
			}
			configRepo.SetSpaceFields(models.SpaceFields{})
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting user information..."},
				[]string{"User", "Org", "Space", "Role"},
				[]string{"user", "Org1", "ORG MANAGER", "BILLING MANAGER", "ORG AUDITOR"},
			))
		})

		It("shows the roles when proper org and space is targeted", func() {

			orgRepo.ListOfRolesOfAnOrg = map[string]string{
				"manager_guid":         "Org1",
				"billing_manager_guid": "Org1",
				"auditor_guid":         "Org1",
			}

			spaceRepo.ListOfRolesOfSpace = map[string]string{
				"managed_spaces": "Space1",
				"developer_guid": "Space1",
				"audited_spaces": "Space1",
			}

			space := models.Space{}
			space.Name = "Space1"
			space.Guid = "space1-guid"

			spaceRepo.FindByNameInOrgSpace = space
			requirementsFactory.Space = space

			configRepo.SetSpaceFields(models.SpaceFields{
				Name: "Space1",
				Guid: "space1-guid",
			})

			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting user information..."},
				[]string{"User", "Org", "Space", "Role"},
				[]string{"user", "Org1", "Space1", "SPACE MANAGER", "SPACE DEVELOPER", "SPACE AUDITOR", "ORG MANAGER", "BILLING MANAGER", "ORG AUDITOR"},
			))
		})

		It("should fail when QueryParam is wrong", func() {

			orgRepo.ListOfRolesOfAnOrg = map[string]string{
				"apiErrorCheck1":  "Org1",
				"apiErrorCheck12": "Org1",
				"apiErrorCheck13": "Org1",
			}
			configRepo.SetSpaceFields(models.SpaceFields{})
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Failed"},
				[]string{"query param not found"},
			))

		})

	})

})
