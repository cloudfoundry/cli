package commands_test

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_organizations "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Login Command", func() {
	var (
		Flags        []string
		Config       core_config.Repository
		ui           *testterm.FakeUI
		authRepo     *testapi.FakeAuthenticationRepository
		endpointRepo *testapi.FakeEndpointRepo
		orgRepo      *fake_organizations.FakeOrganizationRepository
		spaceRepo    *testapi.FakeSpaceRepository

		org  models.Organization
		deps command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = Config
		deps.RepoLocator = deps.RepoLocator.SetEndpointRepository(endpointRepo)
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("login").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		Flags = []string{}
		Config = testconfig.NewRepository()
		ui = &testterm.FakeUI{}
		authRepo = &testapi.FakeAuthenticationRepository{
			AccessToken:  "my_access_token",
			RefreshToken: "my_refresh_token",
			Config:       Config,
		}
		endpointRepo = &testapi.FakeEndpointRepo{}

		org = models.Organization{}
		org.Name = "my-new-org"
		org.Guid = "my-new-org-guid"

		orgRepo = &fake_organizations.FakeOrganizationRepository{}
		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)

		space := models.Space{}
		space.Guid = "my-space-guid"
		space.Name = "my-space"

		spaceRepo = &testapi.FakeSpaceRepository{
			Spaces: []models.Space{space},
		}

		authRepo.GetLoginPromptsReturns.Prompts = map[string]core_config.AuthPrompt{
			"username": core_config.AuthPrompt{
				DisplayName: "Username",
				Type:        core_config.AuthPromptTypeText,
			},
			"password": core_config.AuthPrompt{
				DisplayName: "Password",
				Type:        core_config.AuthPromptTypePassword,
			},
		}
	})

	Context("interactive usage", func() {
		Describe("when there are a small number of organizations and spaces", func() {
			var org2 models.Organization
			var space2 models.Space

			BeforeEach(func() {
				org1 := models.Organization{}
				org1.Guid = "some-org-guid"
				org1.Name = "some-org"

				org2 = models.Organization{}
				org2.Guid = "my-new-org-guid"
				org2.Name = "my-new-org"

				space1 := models.Space{}
				space1.Guid = "my-space-guid"
				space1.Name = "my-space"

				space2 = models.Space{}
				space2.Guid = "some-space-guid"
				space2.Name = "some-space"

				orgRepo.ListOrgsReturns([]models.Organization{org1, org2}, nil)
				spaceRepo.Spaces = []models.Space{space1, space2}
			})

			It("lets the user select an org and space by number", func() {
				orgRepo.FindByNameReturns(org2, nil)
				OUT_OF_RANGE_CHOICE := "3"
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Select an org"},
					[]string{"1. some-org"},
					[]string{"2. my-new-org"},
					[]string{"Select a space"},
					[]string{"1. my-space"},
					[]string{"2. some-space"},
				))

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user select an org and space by name", func() {
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-new-org", "my-space"}
				orgRepo.FindByNameReturns(org2, nil)

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Select an org"},
					[]string{"1. some-org"},
					[]string{"2. my-new-org"},
					[]string{"Select a space"},
					[]string{"1. my-space"},
					[]string{"2. some-space"},
				))

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user specify an org and space using flags", func() {
				Flags = []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "my-space"}

				orgRepo.FindByNameReturns(org2, nil)
				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"username": "user@example.com",
						"password": "password",
					},
				}))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("doesn't ask the user for the API url if they have it in their config", func() {
				orgRepo.FindByNameReturns(org, nil)
				Config.SetApiEndpoint("http://api.example.com")

				Flags = []string{"-o", "my-new-org", "-s", "my-space"}
				ui.Inputs = []string{"user@example.com", "password"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(Config.ApiEndpoint()).To(Equal("http://api.example.com"))
				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})
		Describe("when the CLI version is below the minimum required", func() {
			BeforeEach(func() {
				Config.SetMinCliVersion("5.0.0")
				Config.SetMinRecommendedCliVersion("5.5.0")
			})

			It("prompts users to upgrade if CLI version < min cli version requirement", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
				cf.Version = "4.5.0"

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"To upgrade your CLI"},
					[]string{"5.0.0"},
				))
			})
		})

		Describe("when there are too many orgs to show", func() {
			BeforeEach(func() {
				organizations := []models.Organization{}
				for i := 0; i < 60; i++ {
					id := strconv.Itoa(i)
					org := models.Organization{}
					org.Guid = "my-org-guid-" + id
					org.Name = "my-org-" + id
					organizations = append(organizations, org)
				}
				orgRepo.ListOrgsReturns(organizations, nil)
				orgRepo.FindByNameReturns(organizations[1], nil)

				space1 := models.Space{}
				space1.Guid = "my-space-guid"
				space1.Name = "my-space"

				space2 := models.Space{}
				space2.Guid = "some-space-guid"
				space2.Name = "some-space"

				spaceRepo.Spaces = []models.Space{space1, space2}
			})

			It("doesn't display a list of orgs (the user must type the name)", func() {
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-org-1", "my-space"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"my-org-2"}))
				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-org-1"))
				Expect(Config.OrganizationFields().Guid).To(Equal("my-org-guid-1"))
			})
		})

		Describe("when there is only a single org and space", func() {
			It("does not ask the user to select an org/space", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"username": "user@example.com",
						"password": "password",
					},
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("where there are no available orgs", func() {
			BeforeEach(func() {
				orgRepo.ListOrgsReturns([]models.Organization{}, nil)
				spaceRepo.Spaces = []models.Space{}
			})

			It("does not as the user to select an org", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(Config.OrganizationFields().Guid).To(Equal(""))
				Expect(Config.SpaceFields().Guid).To(Equal(""))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"username": "user@example.com",
						"password": "password",
					},
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("when there is only a single org and no spaces", func() {
			BeforeEach(func() {
				orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
				spaceRepo.Spaces = []models.Space{}
			})

			It("does not ask the user to select a space", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal(""))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"username": "user@example.com",
						"password": "password",
					},
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("login prompts", func() {
			BeforeEach(func() {
				authRepo.GetLoginPromptsReturns.Prompts = map[string]core_config.AuthPrompt{
					"account_number": core_config.AuthPrompt{
						DisplayName: "Account Number",
						Type:        core_config.AuthPromptTypeText,
					},
					"username": core_config.AuthPrompt{
						DisplayName: "Username",
						Type:        core_config.AuthPromptTypeText,
					},
					"passcode": core_config.AuthPrompt{
						DisplayName: "It's a passcode, what you want it to be???",
						Type:        core_config.AuthPromptTypePassword,
					},
					"password": core_config.AuthPrompt{
						DisplayName: "Your Password",
						Type:        core_config.AuthPromptTypePassword,
					},
				}
			})

			Context("when the user does not provide the --sso flag", func() {
				It("prompts the user for 'password' prompt and any text type prompt", func() {
					ui.Inputs = []string{"api.example.com", "the-username", "the-account-number", "the-password"}

					testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"API endpoint"},
						[]string{"Account Number"},
						[]string{"Username"},
					))
					Expect(ui.PasswordPrompts).To(ContainSubstrings([]string{"Your Password"}))
					Expect(ui.PasswordPrompts).ToNot(ContainSubstrings(
						[]string{"passcode"},
					))

					Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
						{
							"account_number": "the-account-number",
							"username":       "the-username",
							"password":       "the-password",
						},
					}))
				})
			})

			Context("when the user does provide the --sso flag", func() {
				It("only prompts the user for the passcode type prompts", func() {
					Flags = []string{"--sso", "-a", "api.example.com"}
					ui.Inputs = []string{"the-one-time-code"}

					testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.PasswordPrompts).To(ContainSubstrings([]string{"passcode"}))
					Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
						{
							"passcode": "the-one-time-code",
						},
					}))
				})
			})

			It("takes the password from the -p flag", func() {
				Flags = []string{"-p", "the-password"}
				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number", "the-pin"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(ui.PasswordPrompts).ToNot(ContainSubstrings([]string{"Your Password"}))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"account_number": "the-account-number",
						"username":       "the-username",
						"password":       "the-password",
					},
				}))
			})

			It("tries 3 times for the password-type prompts", func() {
				authRepo.AuthError = true
				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number",
					"the-password-1", "the-password-2", "the-password-3"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"username":       "the-username",
						"account_number": "the-account-number",
						"password":       "the-password-1",
					},
					{
						"username":       "the-username",
						"account_number": "the-account-number",
						"password":       "the-password-2",
					},
					{
						"username":       "the-username",
						"account_number": "the-account-number",
						"password":       "the-password-3",
					},
				}))

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})

			It("prompts user for password again if password given on the cmd line fails", func() {
				authRepo.AuthError = true

				Flags = []string{"-p", "the-password-1"}

				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number",
					"the-password-2", "the-password-3"}

				testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)

				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
					{
						"account_number": "the-account-number",
						"username":       "the-username",
						"password":       "the-password-1",
					},
					{
						"account_number": "the-account-number",
						"username":       "the-username",
						"password":       "the-password-2",
					},
					{
						"account_number": "the-account-number",
						"username":       "the-username",
						"password":       "the-password-3",
					},
				}))

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})
	})

	Describe("updates to the config", func() {
		BeforeEach(func() {
			Config.SetApiEndpoint("api.the-old-endpoint.com")
			Config.SetAccessToken("the-old-access-token")
			Config.SetRefreshToken("the-old-refresh-token")
		})

		JustBeforeEach(func() {
			testcmd.RunCliCommand("login", Flags, nil, updateCommandDependency, false)
		})

		var ItShowsTheTarget = func() {
			It("shows the target", func() {
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		}

		var ItDoesntShowTheTarget = func() {
			It("does not show the target info", func() {
				Expect(ui.ShowConfigurationCalled).To(BeFalse())
			})
		}

		var ItFails = func() {
			It("fails", func() {
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		}

		var ItSucceeds = func() {
			It("runs successfully", func() {
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
				Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
			})
		}

		Describe("when the user is setting an API", func() {
			BeforeEach(func() {
				Flags = []string{"-a", "https://api.the-server.com", "-u", "the-user-name", "-p", "the-password"}
			})

			Describe("when the --skip-ssl-validation flag is provided", func() {
				BeforeEach(func() {
					Flags = append(Flags, "--skip-ssl-validation")
				})

				Describe("setting api endpoint is successful", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(false)
					})

					ItSucceeds()
					ItShowsTheTarget()

					It("stores the API endpoint and the skip-ssl flag", func() {
						Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeTrue())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.UpdateEndpointError = errors.New("API endpoint not found")
					})

					ItFails()
					ItDoesntShowTheTarget()

					It("clears the entire config", func() {
						Expect(Config.ApiEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().Guid).To(BeEmpty())
						Expect(Config.SpaceFields().Guid).To(BeEmpty())
					})
				})
			})

			Describe("when the --skip-ssl-validation flag is not provided", func() {
				Describe("setting api endpoint is successful", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
					})

					ItSucceeds()
					ItShowsTheTarget()

					It("updates the API endpoint and enables SSL validation", func() {
						Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeFalse())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.UpdateEndpointError = errors.New("API endpoint not found")
					})

					ItFails()
					ItDoesntShowTheTarget()

					It("clears the entire config", func() {
						Expect(Config.ApiEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().Guid).To(BeEmpty())
						Expect(Config.SpaceFields().Guid).To(BeEmpty())
					})
				})
			})

			Describe("when there is an invalid SSL cert", func() {
				BeforeEach(func() {
					endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://bobs-burgers.com", "SELF SIGNED SADNESS")
					ui.Inputs = []string{"bobs-burgers.com"}
				})

				It("fails and suggests the user skip SSL validation", func() {
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"SSL Cert", "https://bobs-burgers.com"},
						[]string{"TIP", "login", "--skip-ssl-validation"},
					))
				})

				ItDoesntShowTheTarget()
			})
		})

		Describe("when user is logging in and not setting the api endpoint", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "the-user-name", "-p", "the-password"}
			})

			Describe("when the --skip-ssl-validation flag is provided", func() {
				BeforeEach(func() {
					Flags = append(Flags, "--skip-ssl-validation")
					Config.SetSSLDisabled(false)
				})

				It("disables SSL validation", func() {
					Expect(Config.IsSSLDisabled()).To(BeTrue())
				})
			})

			Describe("when the --skip-ssl-validation flag is not provided", func() {
				BeforeEach(func() {
					Config.SetSSLDisabled(true)
				})

				It("should not change config's SSLDisabled flag", func() {
					Expect(Config.IsSSLDisabled()).To(BeTrue())
				})
			})

			Describe("and the login fails authenticaton", func() {
				BeforeEach(func() {
					authRepo.AuthError = true

					Config.SetSSLDisabled(true)

					Flags = []string{"-u", "user@example.com"}
					ui.Inputs = []string{"password", "password2", "password3", "password4"}
				})

				ItFails()
				ItShowsTheTarget()

				It("does not change the api endpoint or SSL setting in the config", func() {
					Expect(Config.ApiEndpoint()).To(Equal("api.the-old-endpoint.com"))
					Expect(Config.IsSSLDisabled()).To(BeTrue())
				})

				It("clears Access Token, Refresh Token, Org, and Space in the config", func() {
					Expect(Config.AccessToken()).To(BeEmpty())
					Expect(Config.RefreshToken()).To(BeEmpty())
					Expect(Config.OrganizationFields().Guid).To(BeEmpty())
					Expect(Config.SpaceFields().Guid).To(BeEmpty())
				})
			})
		})

		Describe("and the login fails to target an org", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "nonexistentorg", "-s", "my-space"}
				orgRepo.FindByNameReturns(models.Organization{}, errors.New("No org"))
				Config.SetSSLDisabled(true)
			})

			ItFails()
			ItShowsTheTarget()

			It("does not update the api endpoint or ssl setting in the config", func() {
				Expect(Config.ApiEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("clears Org, and Space in the config", func() {
				Expect(Config.OrganizationFields().Guid).To(BeEmpty())
				Expect(Config.SpaceFields().Guid).To(BeEmpty())
			})
		})

		Describe("and the login fails to target a space", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "nonexistent"}
				orgRepo.FindByNameReturns(org, nil)

				Config.SetSSLDisabled(true)
			})

			ItFails()
			ItShowsTheTarget()

			It("does not update the api endpoint or ssl setting in the config", func() {
				Expect(Config.ApiEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("updates the org in the config", func() {
				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
			})

			It("clears the space in the config", func() {
				Expect(Config.SpaceFields().Guid).To(BeEmpty())
			})
		})

		Describe("and the login succeeds", func() {
			BeforeEach(func() {
				orgRepo.FindByNameReturns(models.Organization{
					OrganizationFields: models.OrganizationFields{
						Name: "new-org",
						Guid: "new-org-guid",
					},
				}, nil)
				spaceRepo.Spaces[0].Name = "new-space"
				spaceRepo.Spaces[0].Guid = "new-space-guid"
				authRepo.AccessToken = "new_access_token"
				authRepo.RefreshToken = "new_refresh_token"

				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "new-org", "-s", "new-space"}

				Config.SetApiEndpoint("api.the-old-endpoint.com")
				Config.SetSSLDisabled(true)
			})

			ItSucceeds()
			ItShowsTheTarget()

			It("does not update the api endpoint or SSL setting", func() {
				Expect(Config.ApiEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("updates Access Token, Refresh Token, Org, and Space in the config", func() {
				Expect(Config.AccessToken()).To(Equal("new_access_token"))
				Expect(Config.RefreshToken()).To(Equal("new_refresh_token"))
				Expect(Config.OrganizationFields().Guid).To(Equal("new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("new-space-guid"))
			})
		})
	})
})
