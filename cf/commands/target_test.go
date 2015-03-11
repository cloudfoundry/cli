package commands_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
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
		endpointRepo        *testapi.FakeEndpointRepo
		config              core_config.ReadWriter
		ui                  *testterm.FakeUI
		Flags               []string
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		orgRepo = new(fake_org.FakeOrganizationRepository)
		spaceRepo = new(testapi.FakeSpaceRepository)
		requirementsFactory = new(testreq.FakeReqFactory)
		endpointRepo = &testapi.FakeEndpointRepo{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory.ApiEndpointSuccess = true
	})

	var callTarget = func(args []string) bool {
		cmd := NewTarget(ui, config, endpointRepo, orgRepo, spaceRepo)
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

		It("able to set api endpoint", func() {
			callTarget([]string{"-a", "https://api.example.com"})
			Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
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

		It("able to set api endpoint", func() {
			callTarget([]string{"-a", "https://api.example.com"})
			Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
		})

		It("fails requirements when targeting a space or org with api endpoint set", func() {
			Expect(callTarget([]string{"-a", "https://api.example.com", "-o", "some-crazy-org-im-not-in"})).To(BeFalse())

			Expect(callTarget([]string{"-a", "https://api.example.com", "-s", "i-love-space"})).To(BeFalse())
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

			It("it updates the organization in the config for new api endpoint", func() {
				callTarget([]string{"-a", "https://api.example.com", "-o", "my-organization"})

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())

				Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
			})

			It("updates the space in the config for new api endpoint", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"

				spaceRepo.Spaces = []models.Space{space}
				spaceRepo.FindByNameSpace = space

				callTarget([]string{"-a", "https://api.example.com", "-s", "my-space"})

				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
			})

			It("updates both the organization and the space in the config for new api endpoint", func() {
				space := models.Space{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"
				spaceRepo.Spaces = []models.Space{space}

				callTarget([]string{"-a", "https://api.example.com", "-o", "my-organization", "-s", "my-space"})

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-organization"))
				Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))

				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
			})

			It("only updates the api endpoint in the config when the org can't be found", func() {
				config.SetOrganizationFields(models.OrganizationFields{})

				orgRepo.FindByNameReturns(models.Organization{}, errors.New("my-organization not found"))

				callTarget([]string{"-a", "https://api.example.com", "-o", "my-organization"})

				Expect(config.OrganizationFields().Guid).To(Equal(""))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))

				Expect(ui.ShowConfigurationCalled).To(BeFalse())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Could not target org."},
					[]string{"my-organization not found"},
				))
			})

			Context("when the user is setting an API", func() {
				BeforeEach(func() {
					Flags = []string{"-a", "https://api.example.com"}
				})

				Context("when the --skip-ssl-validation flag is provided", func() {
					BeforeEach(func() {
						Flags = append(Flags, "--skip-ssl-validation")
					})

					It("stores the API endpoint and the skip-ssl flag", func() {
						callTarget(Flags)

						Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
						Expect(config.IsSSLDisabled()).To(BeTrue())
					})
				})

				Context("when the --skip-ssl-validation flag is not provided", func() {
					Describe("setting api endpoint is successful", func() {
						BeforeEach(func() {
							config.SetSSLDisabled(true)
						})

						It("updates the API endpoint and enables SSL validation", func() {
							callTarget(Flags)

							Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.example.com"))
							Expect(config.IsSSLDisabled()).To(BeFalse())
						})
					})

				})

				Context("when there is an invalid SSL cert", func() {
					BeforeEach(func() {
						endpointRepo.UpdateEndpointError = cferrors.NewInvalidSSLCert("https://bobs-burgers.com", "SELF SIGNED SADNESS")
						ui.Inputs = []string{"bobs-burgers.com"}
					})

					It("fails and suggests the user skip SSL validation", func() {
						callTarget(Flags)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"SSL Cert", "https://bobs-burgers.com"},
							[]string{"TIP", "target", "--skip-ssl-validation"},
						))
					})
				})
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

			Context("when setting api endpoint failed", func() {
				BeforeEach(func() {
					config.SetSSLDisabled(true)
					endpointRepo.UpdateEndpointError = errors.New("API endpoint not found")
				})

				It("clears api endpoint from config", func() {
					callTarget([]string{"-a", "https://api.example.com"})

					Expect(config.ApiEndpoint()).To(BeEmpty())
					Expect(config.IsSSLDisabled()).To(BeFalse())
				})

			})
		})
	})
})
