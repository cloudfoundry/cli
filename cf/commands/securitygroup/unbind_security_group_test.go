package securitygroup_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"

	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	fakeBinder "github.com/cloudfoundry/cli/cf/api/security_groups/spaces/fakes"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unbind-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *fakeSecurityGroup.FakeSecurityGroupRepo
		orgRepo             *fake_org.FakeOrganizationRepository
		spaceRepo           *fakes.FakeSpaceRepository
		secBinder           *fakeBinder.FakeSecurityGroupSpaceBinder
		requirementsFactory *testreq.FakeReqFactory
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		orgRepo = &fake_org.FakeOrganizationRepository{}
		spaceRepo = &fakes.FakeSpaceRepository{}
		secBinder = &fakeBinder.FakeSecurityGroupSpaceBinder{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		cmd := NewUnbindSecurityGroup(ui, configRepo, securityGroupRepo, orgRepo, spaceRepo, secBinder)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("should fail if not logged in", func() {
			runCommand("my-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("should fail with usage when not provided with any arguments", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("should fail with usage when provided with a number of arguments that is either 2 or 4 or a number larger than 4", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("I", "like")
			Expect(ui.FailedWithUsage).To(BeTrue())
			runCommand("Turn", "down", "for", "what")
			Expect(ui.FailedWithUsage).To(BeTrue())
			runCommand("My", "Very", "Excellent", "Mother", "Just", "Sat", "Under", "Nine", "ThingsThatArentPlanets")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when everything exists", func() {
			BeforeEach(func() {
				securityGroup := models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name:  "my-group",
						Guid:  "my-group-guid",
						Rules: []map[string]interface{}{},
					},
				}

				securityGroupRepo.ReadReturns(securityGroup, nil)

				orgRepo.ListOrgsReturns([]models.Organization{{
					OrganizationFields: models.OrganizationFields{
						Name: "my-org",
						Guid: "my-org-guid",
					}},
				}, nil)

				spaceRepo.FindByNameInOrgSpace = models.Space{SpaceFields: models.SpaceFields{Name: "my-space", Guid: "my-space-guid"}}
			})

			It("removes the security group when we only pass the security group name (using the targeted org and space)", func() {
				runCommand("my-group")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding security group", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				securityGroupGuid, spaceGuid := secBinder.UnbindSpaceArgsForCall(0)
				Expect(securityGroupGuid).To(Equal("my-group-guid"))
				Expect(spaceGuid).To(Equal("my-space-guid"))
			})

			It("removes the security group when we pass the org and space", func() {
				runCommand("my-group", "my-org", "my-space")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding security group", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				securityGroupGuid, spaceGuid := secBinder.UnbindSpaceArgsForCall(0)
				Expect(securityGroupGuid).To(Equal("my-group-guid"))
				Expect(spaceGuid).To(Equal("my-space-guid"))
			})
		})

		Context("when one of the things does not exist", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.New("I accidentally the"))
			})

			It("fails with an error", func() {
				runCommand("my-group", "my-org", "my-space")
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})
	})
})
