package commands_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("target command", func() {
	var (
		orgRepo             *fake_org.FakeOrganizationRepository
		spaceRepo           *testapi.FakeSpaceRepository
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.ReadWriter
		ui                  *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		orgRepo = new(fake_org.FakeOrganizationRepository)
		spaceRepo = new(testapi.FakeSpaceRepository)
		requirementsFactory = new(testreq.FakeReqFactory)
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory.ApiEndpointSuccess = true
	})

	var callTarget = func(args []string) bool {
		cmd := NewTarget(ui, config, orgRepo, spaceRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails with usage when called with an argument but no flags", func() {
		callTarget([]string{"some-argument"})
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	Describe("when there is no api endpoint set", func() {
		BeforeEach(func() {
			requirementsFactory.ApiEndpointSuccess = false
		})

		It("fails requirements", func() {
			Expect(callTarget([]string{})).To(BeFalse())
		})
	})

	Describe("when the user is not logged in", func() {
		BeforeEach(func() {
			config.SetAccessToken("")
		})

		It("prints the target info when no org or space is specified", func() {
			Expect(callTarget([]string{})).To(BeTrue())
			Expect(ui.ShowConfigurationCalled).To(BeTrue())
		})

		It("panics silently so that it returns an exit code of 1", func() {
			callTarget([]string{})
			Expect(ui.PanickedQuietly).To(BeTrue())
		})

		It("fails requirements when targeting a space or org", func() {
			Expect(callTarget([]string{"-o", "some-crazy-org-im-not-in"})).To(BeFalse())

			Expect(callTarget([]string{"-s", "i-love-space"})).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		var expectOrgToBeCleared = func() {
			Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
		}

		var expectSpaceToBeCleared = func() {
			Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
		}

		Context("there are no errors", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "my-organization"
				org.Guid = "my-organization-guid"

				orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
				orgRepo.FindByNameReturns(org, nil)
			})

			It("it updates the organization in the config", func() {
				callTarget([]string{"-o", "my-organization"})

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())

				Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))
			})

			It("updates the space in the config", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"

				spaceRepo.Spaces = []models.Space{space}
				spaceRepo.FindByNameSpace = space

				callTarget([]string{"-s", "my-space"})

				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("updates both the organization and the space in the config", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"
				spaceRepo.Spaces = []models.Space{space}

				callTarget([]string{"-o", "my-organization", "-s", "my-space"})

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))

				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("only updates the organization in the config when the space can't be found", func() {
				config.SetSpaceFields(models.SpaceFields{})

				spaceRepo.FindByNameErr = true

				callTarget([]string{"-o", "my-organization", "-s", "my-space"})

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))

				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(config.SpaceFields().Guid).To(Equal(""))

				Expect(ui.ShowConfigurationCalled).To(BeFalse())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Unable to access space", "my-space"},
				))
			})
		})

		Context("there are errors", func() {
			It("fails when the user does not have access to the specified organization", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("Invalid access"))

				callTarget([]string{"-o", "my-organization"})
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
				expectOrgToBeCleared()
				expectSpaceToBeCleared()
			})

			It("fails when the organization is not found", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("my-organization not found"))

				callTarget([]string{"-o", "my-organization"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"my-organization", "not found"},
				))

				expectOrgToBeCleared()
				expectSpaceToBeCleared()
			})

			It("fails to target a space if no organization is targeted", func() {
				config.SetOrganizationFields(models.OrganizationFields{})

				callTarget([]string{"-s", "my-space"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An org must be targeted before targeting a space"},
				))

				expectSpaceToBeCleared()
			})

			It("fails when the user doesn't have access to the space", func() {
				spaceRepo.FindByNameErr = true

				callTarget([]string{"-s", "my-space"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Unable to access space", "my-space"},
				))

				Expect(config.SpaceFields().Guid).To(Equal(""))
				Expect(ui.ShowConfigurationCalled).To(BeFalse())

				Expect(config.OrganizationFields().Guid).NotTo(BeEmpty())
				expectSpaceToBeCleared()
			})

			It("fails when the space is not found", func() {
				spaceRepo.FindByNameNotFound = true

				callTarget([]string{"-s", "my-space"})

				expectSpaceToBeCleared()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"my-space", "not found"},
				))
			})
		})
	})
})
