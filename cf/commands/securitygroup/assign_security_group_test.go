package securitygroup_test

import (
	"github.com/cloudfoundry/cli/cf/api/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	zoidberg "github.com/cloudfoundry/cli/cf/api/security_groups/spaces/fakes"
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

var _ = Describe("assign-security-group command", func() {
	var (
		ui                    *testterm.FakeUI
		cmd                   AssignSecurityGroup
		configRepo            configuration.ReadWriter
		fakeSecurityGroupRepo *testapi.FakeSecurityGroup
		requirementsFactory   *testreq.FakeReqFactory
		fakeSpaceRepo         *fakes.FakeSpaceRepository
		fakeOrgRepo           *fakes.FakeOrgRepository
		fakeSpaceBinder       *zoidberg.FakeSecurityGroupSpaceBinder
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		fakeOrgRepo = &fakes.FakeOrgRepository{}
		fakeSpaceRepo = &fakes.FakeSpaceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &testapi.FakeSecurityGroup{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		fakeSpaceBinder = &zoidberg.FakeSecurityGroupSpaceBinder{}
		cmd = NewAssignSecurityGroup(ui, configRepo, fakeSecurityGroupRepo, fakeSpaceRepo, fakeOrgRepo, fakeSpaceBinder)
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("my-craaaaaazy-security-group", "my-org", "my-space")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("my-craaaaaazy-security-group", "my-org", "my-space")

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails with usage when not provided the name of a security group, org, and space", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("one fish", "two fish", "three fish", "purple fish")

			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in and provides the name of a security group", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when a security group with that name does not exist", func() {
			BeforeEach(func() {
				fakeSecurityGroupRepo.ReadReturns.Error = errors.NewModelNotFoundError("security group", "my-nonexistent-security-group")
			})

			It("fails and tells the user", func() {
				runCommand("my-nonexistent-security-group", "my-org", "my-space")

				Expect(fakeSecurityGroupRepo.ReadCalledWith.Name).To(Equal("my-nonexistent-security-group"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"security group", "my-nonexistent-security-group", "not found"},
				))
			})
		})

		Context("when the org does not exist", func() {
			BeforeEach(func() {
				fakeOrgRepo.FindByNameNotFound = true
			})

			It("fails and tells the user", func() {
				runCommand("sec group", "org", "space")

				Expect(fakeOrgRepo.FindByNameName).To(Equal("org"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Org", "org", "not found"},
				))
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "org-name"
				org.Guid = "org-guid"
				fakeOrgRepo.Organizations = append(fakeOrgRepo.Organizations, org) // TODO: replace this with countfeiter
				fakeSpaceRepo.FindByNameInOrgError = errors.NewModelNotFoundError("Space", "space-name")
			})

			It("fails and tells the user", func() {
				runCommand("sec group", "org-name", "space-name")

				Expect(fakeSpaceRepo.FindByNameInOrgName).To(Equal("space-name"))
				Expect(fakeSpaceRepo.FindByNameInOrgOrgGuid).To(Equal("org-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Space", "space-name", "not found"},
				))
			})
		})

		Context("everything is hunky dory", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "org-name"
				org.Guid = "org-guid"
				fakeOrgRepo.Organizations = append(fakeOrgRepo.Organizations, org) // TODO: replace this with countfeiter

				space := models.Space{}
				space.Name = "space-name"
				space.Guid = "space-guid"
				fakeSpaceRepo.FindByNameInOrgSpace = space

				securityGroup := models.SecurityGroup{}
				securityGroup.Name = "security-group"
				securityGroup.Guid = "security-group-guid"
				fakeSecurityGroupRepo.ReadReturns.SecurityGroup = securityGroup
			})

			JustBeforeEach(func() {
				runCommand("security-group", "org-name", "space-name")
			})

			It("assigns the security group to the space", func() {
				secGroupGuid, spaceGuid := fakeSpaceBinder.BindSpaceArgsForCall(0)
				Expect(secGroupGuid).To(Equal("security-group-guid"))
				Expect(spaceGuid).To(Equal("space-guid"))
			})

			It("describes what it is doing for the user's benefit", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Assigning", "security-group", "space-name", "org-name", "my-user"},
					[]string{"OK"},
				))
			})
		})
	})
})
