package commands_test

import (
	"strconv"

	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("Login Command", func() {
	var (
		Flags        []string
		Config       coreconfig.Repository
		ui           *testterm.FakeUI
		authRepo     *authenticationfakes.FakeRepository
		endpointRepo *coreconfigfakes.FakeEndpointRepository
		orgRepo      *organizationsfakes.FakeOrganizationRepository
		spaceRepo    *spacesfakes.FakeSpaceRepository

		org  models.Organization
		deps commandregistry.Dependency

		minCLIVersion            string
		minRecommendedCLIVersion string
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = Config
		deps.RepoLocator = deps.RepoLocator.SetEndpointRepository(endpointRepo)
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("login").SetDependency(deps, pluginCall))
	}

	listSpacesStub := func(spaces []models.Space) func(func(models.Space) bool) error {
		return func(cb func(models.Space) bool) error {
			var keepGoing bool
			for _, s := range spaces {
				keepGoing = cb(s)
				if !keepGoing {
					return nil
				}
			}
			return nil
		}
	}

	BeforeEach(func() {
		Flags = []string{}
		Config = testconfig.NewRepository()
		ui = &testterm.FakeUI{}
		authRepo = new(authenticationfakes.FakeRepository)
		authRepo.AuthenticateStub = func(credentials map[string]string) error {
			Config.SetAccessToken("my_access_token")
			Config.SetRefreshToken("my_refresh_token")
			return nil
		}
		endpointRepo = new(coreconfigfakes.FakeEndpointRepository)
		minCLIVersion = "1.0.0"
		minRecommendedCLIVersion = "1.0.0"

		org = models.Organization{}
		org.Name = "my-new-org"
		org.GUID = "my-new-org-guid"

		orgRepo = &organizationsfakes.FakeOrganizationRepository{}
		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)

		space := models.Space{}
		space.GUID = "my-space-guid"
		space.Name = "my-space"

		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space})

		authRepo.GetLoginPromptsAndSaveUAAServerURLReturns(map[string]coreconfig.AuthPrompt{
			"username": {
				DisplayName: "Username",
				Type:        coreconfig.AuthPromptTypeText,
			},
			"password": {
				DisplayName: "Password",
				Type:        coreconfig.AuthPromptTypePassword,
			},
		}, nil)
	})

	Context("interactive usage", func() {
		JustBeforeEach(func() {
			endpointRepo.GetCCInfoStub = func(endpoint string) (*coreconfig.CCInfo, string, error) {
				return &coreconfig.CCInfo{
					APIVersion:               "some-version",
					AuthorizationEndpoint:    "auth/endpoint",
					LoggregatorEndpoint:      "loggregator/endpoint",
					MinCLIVersion:            minCLIVersion,
					MinRecommendedCLIVersion: minRecommendedCLIVersion,
					SSHOAuthClient:           "some-client",
					RoutingAPIEndpoint:       "routing/endpoint",
				}, endpoint, nil
			}
		})

		Describe("when there are a small number of organizations and spaces", func() {
			var org2 models.Organization
			var space2 models.Space

			BeforeEach(func() {
				org1 := models.Organization{}
				org1.GUID = "some-org-guid"
				org1.Name = "some-org"

				org2 = models.Organization{}
				org2.GUID = "my-new-org-guid"
				org2.Name = "my-new-org"

				space1 := models.Space{}
				space1.GUID = "my-space-guid"
				space1.Name = "my-space"

				space2 = models.Space{}
				space2.GUID = "some-space-guid"
				space2.Name = "some-space"

				orgRepo.ListOrgsReturns([]models.Organization{org1, org2}, nil)
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space1, space2})
				spaceRepo.FindByNameStub = func(name string) (models.Space, error) {
					m := map[string]models.Space{
						space1.Name: space1,
						space2.Name: space2,
					}
					return m[name], nil
				}
			})

			It("lets the user select an org and space by number", func() {
				orgRepo.FindByNameReturns(org2, nil)
				OUT_OF_RANGE_CHOICE := "3"
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Select an org"},
					[]string{"1. some-org"},
					[]string{"2. my-new-org"},
					[]string{"Select a space"},
					[]string{"1. my-space"},
					[]string{"2. some-space"},
				))

				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(Config.APIEndpoint()).To(Equal("api.example.com"))
				Expect(Config.APIVersion()).To(Equal("some-version"))
				Expect(Config.AuthenticationEndpoint()).To(Equal("auth/endpoint"))
				Expect(Config.SSHOAuthClient()).To(Equal("some-client"))
				Expect(Config.MinCLIVersion()).To(Equal("1.0.0"))
				Expect(Config.MinRecommendedCLIVersion()).To(Equal("1.0.0"))
				Expect(Config.LoggregatorEndpoint()).To(Equal("loggregator/endpoint"))
				Expect(Config.DopplerEndpoint()).To(Equal("doppler/endpoint"))
				Expect(Config.RoutingAPIEndpoint()).To(Equal("routing/endpoint"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("api.example.com"))

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user select an org and space by name", func() {
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-new-org", "my-space"}
				orgRepo.FindByNameReturns(org2, nil)

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Select an org"},
					[]string{"1. some-org"},
					[]string{"2. my-new-org"},
					[]string{"Select a space"},
					[]string{"1. my-space"},
					[]string{"2. some-space"},
				))

				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("api.example.com"))

				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameArgsForCall(0)).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user specify an org and space using flags", func() {
				Flags = []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "my-space"}

				orgRepo.FindByNameReturns(org2, nil)
				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("api.example.com"))
				Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("doesn't ask the user for the API url if they have it in their config", func() {
				orgRepo.FindByNameReturns(org, nil)
				Config.SetAPIEndpoint("http://api.example.com")

				Flags = []string{"-o", "my-new-org", "-s", "my-space"}
				ui.Inputs = []string{"user@example.com", "password"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(Config.APIEndpoint()).To(Equal("http://api.example.com"))
				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("http://api.example.com"))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		It("displays an update notification", func() {
			ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
			testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)
			Expect(ui.NotifyUpdateIfNeededCallCount).To(Equal(1))
		})

		It("tries to get the organizations", func() {
			Flags = []string{}
			ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-org-1", "my-space"}
			testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)
			Expect(orgRepo.ListOrgsCallCount()).To(Equal(1))
			Expect(orgRepo.ListOrgsArgsForCall(0)).To(Equal(50))
		})

		Describe("when there are too many orgs to show", func() {
			BeforeEach(func() {
				organizations := []models.Organization{}
				for i := 0; i < 60; i++ {
					id := strconv.Itoa(i)
					org := models.Organization{}
					org.GUID = "my-org-guid-" + id
					org.Name = "my-org-" + id
					organizations = append(organizations, org)
				}
				orgRepo.ListOrgsReturns(organizations, nil)
				orgRepo.FindByNameReturns(organizations[1], nil)

				space1 := models.Space{}
				space1.GUID = "my-space-guid"
				space1.Name = "my-space"

				space2 := models.Space{}
				space2.GUID = "some-space-guid"
				space2.Name = "some-space"

				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space1, space2})
			})

			It("doesn't display a list of orgs (the user must type the name)", func() {
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-org-1", "my-space"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"my-org-2"}))
				Expect(orgRepo.FindByNameArgsForCall(0)).To(Equal("my-org-1"))
				Expect(Config.OrganizationFields().GUID).To(Equal("my-org-guid-1"))
			})
		})

		Describe("when there is only a single org and space", func() {
			It("does not ask the user to select an org/space", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("where there are no available orgs", func() {
			BeforeEach(func() {
				orgRepo.ListOrgsReturns([]models.Organization{}, nil)
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{})
			})

			It("does not as the user to select an org", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(Config.OrganizationFields().GUID).To(Equal(""))
				Expect(Config.SpaceFields().GUID).To(Equal(""))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("when there is only a single org and no spaces", func() {
			BeforeEach(func() {
				orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{})
			})

			It("does not ask the user to select a space", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal(""))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("login prompts", func() {
			BeforeEach(func() {
				authRepo.GetLoginPromptsAndSaveUAAServerURLReturns(map[string]coreconfig.AuthPrompt{
					"account_number": {
						DisplayName: "Account Number",
						Type:        coreconfig.AuthPromptTypeText,
					},
					"username": {
						DisplayName: "Username",
						Type:        coreconfig.AuthPromptTypeText,
					},
					"passcode": {
						DisplayName: "It's a passcode, what you want it to be???",
						Type:        coreconfig.AuthPromptTypePassword,
					},
					"password": {
						DisplayName: "Your Password",
						Type:        coreconfig.AuthPromptTypePassword,
					},
				}, nil)
			})

			Context("when the user does not provide the --sso flag", func() {
				It("prompts the user for 'password' prompt and any text type prompt", func() {
					ui.Inputs = []string{"api.example.com", "the-username", "the-account-number", "the-password"}

					testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"API endpoint"},
						[]string{"Account Number"},
						[]string{"Username"},
					))
					Expect(ui.PasswordPrompts).To(ContainSubstrings([]string{"Your Password"}))
					Expect(ui.PasswordPrompts).ToNot(ContainSubstrings(
						[]string{"passcode"},
					))

					Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
					Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
						"account_number": "the-account-number",
						"username":       "the-username",
						"password":       "the-password",
					}))
				})
			})

			Context("when the user does provide the --sso flag", func() {
				It("only prompts the user for the passcode type prompts", func() {
					Flags = []string{"--sso", "-a", "api.example.com"}
					ui.Inputs = []string{"the-one-time-code"}

					testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.PasswordPrompts).To(ContainSubstrings([]string{"passcode"}))
					Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
					Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
						"passcode": "the-one-time-code",
					}))
				})
			})

			It("takes the password from the -p flag", func() {
				Flags = []string{"-p", "the-password"}
				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number", "the-pin"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(ui.PasswordPrompts).ToNot(ContainSubstrings([]string{"Your Password"}))
				Expect(authRepo.AuthenticateCallCount()).To(Equal(1))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"account_number": "the-account-number",
					"username":       "the-username",
					"password":       "the-password",
				}))
			})

			It("tries 3 times for the password-type prompts", func() {
				authRepo.AuthenticateReturns(errors.New("Error authenticating."))
				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number",
					"the-password-1", "the-password-2", "the-password-3"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(authRepo.AuthenticateCallCount()).To(Equal(3))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-1",
				}))
				Expect(authRepo.AuthenticateArgsForCall(1)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-2",
				}))
				Expect(authRepo.AuthenticateArgsForCall(2)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-3",
				}))

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})

			It("prompts user for password again if password given on the cmd line fails", func() {
				authRepo.AuthenticateReturns(errors.New("Error authenticating."))

				Flags = []string{"-p", "the-password-1"}

				ui.Inputs = []string{"api.example.com", "the-username", "the-account-number",
					"the-password-2", "the-password-3"}

				testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)

				Expect(authRepo.AuthenticateCallCount()).To(Equal(3))
				Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-1",
				}))
				Expect(authRepo.AuthenticateArgsForCall(1)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-2",
				}))
				Expect(authRepo.AuthenticateArgsForCall(2)).To(Equal(map[string]string{
					"username":       "the-username",
					"account_number": "the-account-number",
					"password":       "the-password-3",
				}))

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})
		})
	})

	Describe("updates to the config", func() {
		BeforeEach(func() {
			Config.SetAPIEndpoint("api.the-old-endpoint.com")
			Config.SetAccessToken("the-old-access-token")
			Config.SetRefreshToken("the-old-refresh-token")
			endpointRepo.GetCCInfoStub = func(endpoint string) (*coreconfig.CCInfo, string, error) {
				return &coreconfig.CCInfo{
					APIVersion:               "some-version",
					AuthorizationEndpoint:    "auth/endpoint",
					LoggregatorEndpoint:      "loggregator/endpoint",
					MinCLIVersion:            minCLIVersion,
					MinRecommendedCLIVersion: minRecommendedCLIVersion,
					SSHOAuthClient:           "some-client",
					RoutingAPIEndpoint:       "routing/endpoint",
				}, endpoint, nil
			}

		})

		JustBeforeEach(func() {
			testcmd.RunCLICommand("login", Flags, nil, updateCommandDependency, false, ui)
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
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})
		}

		var ItSucceeds = func() {
			It("runs successfully", func() {
				Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
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
						Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
						Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeTrue())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.GetCCInfoReturns(nil, "", errors.New("API endpoint not found"))
					})

					ItFails()
					ItDoesntShowTheTarget()

					It("clears the entire config", func() {
						Expect(Config.APIEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().GUID).To(BeEmpty())
						Expect(Config.SpaceFields().GUID).To(BeEmpty())
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
						Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
						Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeFalse())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.GetCCInfoReturns(nil, "", errors.New("API endpoint not found"))
					})

					ItFails()
					ItDoesntShowTheTarget()

					It("clears the entire config", func() {
						Expect(Config.APIEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().GUID).To(BeEmpty())
						Expect(Config.SpaceFields().GUID).To(BeEmpty())
					})
				})
			})

			Describe("when there is an invalid SSL cert", func() {
				BeforeEach(func() {
					endpointRepo.GetCCInfoReturns(nil, "", errors.NewInvalidSSLCert("https://bobs-burgers.com", "SELF SIGNED SADNESS"))
					ui.Inputs = []string{"bobs-burgers.com"}
				})

				It("fails and suggests the user skip SSL validation", func() {
					Expect(ui.Outputs()).To(ContainSubstrings(
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
					authRepo.AuthenticateReturns(errors.New("Error authenticating."))

					Config.SetSSLDisabled(true)

					Flags = []string{"-u", "user@example.com"}
					ui.Inputs = []string{"password", "password2", "password3", "password4"}
				})

				ItFails()
				ItShowsTheTarget()

				It("does not change the api endpoint or SSL setting in the config", func() {
					Expect(Config.APIEndpoint()).To(Equal("api.the-old-endpoint.com"))
					Expect(Config.IsSSLDisabled()).To(BeTrue())
				})

				It("clears Access Token, Refresh Token, Org, and Space in the config", func() {
					Expect(Config.AccessToken()).To(BeEmpty())
					Expect(Config.RefreshToken()).To(BeEmpty())
					Expect(Config.OrganizationFields().GUID).To(BeEmpty())
					Expect(Config.SpaceFields().GUID).To(BeEmpty())
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
				Expect(Config.APIEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("clears Org, and Space in the config", func() {
				Expect(Config.OrganizationFields().GUID).To(BeEmpty())
				Expect(Config.SpaceFields().GUID).To(BeEmpty())
			})
		})

		Describe("and the login fails to target a space", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "nonexistent"}
				orgRepo.FindByNameReturns(org, nil)
				spaceRepo.FindByNameReturns(models.Space{}, errors.New("find-by-name-err"))

				Config.SetSSLDisabled(true)
			})

			ItFails()
			ItShowsTheTarget()

			It("does not update the api endpoint or ssl setting in the config", func() {
				Expect(Config.APIEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("updates the org in the config", func() {
				Expect(Config.OrganizationFields().GUID).To(Equal("my-new-org-guid"))
			})

			It("clears the space in the config", func() {
				Expect(Config.SpaceFields().GUID).To(BeEmpty())
			})
		})

		Describe("and the login succeeds", func() {
			BeforeEach(func() {
				orgRepo.FindByNameReturns(models.Organization{
					OrganizationFields: models.OrganizationFields{
						Name: "new-org",
						GUID: "new-org-guid",
					},
				}, nil)

				space1 := models.Space{}
				space1.GUID = "new-space-guid"
				space1.Name = "new-space-name"
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space1})
				spaceRepo.FindByNameReturns(space1, nil)

				authRepo.AuthenticateStub = func(credentials map[string]string) error {
					Config.SetAccessToken("new_access_token")
					Config.SetRefreshToken("new_refresh_token")
					return nil
				}

				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "new-org", "-s", "new-space"}

				Config.SetAPIEndpoint("api.the-old-endpoint.com")
				Config.SetSSLDisabled(true)
			})

			ItSucceeds()
			ItShowsTheTarget()

			It("does not update the api endpoint or SSL setting", func() {
				Expect(Config.APIEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("updates the config", func() {
				Expect(Config.AccessToken()).To(Equal("new_access_token"))
				Expect(Config.RefreshToken()).To(Equal("new_refresh_token"))
				Expect(Config.OrganizationFields().GUID).To(Equal("new-org-guid"))
				Expect(Config.SpaceFields().GUID).To(Equal("new-space-guid"))

				Expect(Config.APIVersion()).To(Equal("some-version"))
				Expect(Config.AuthenticationEndpoint()).To(Equal("auth/endpoint"))
				Expect(Config.SSHOAuthClient()).To(Equal("some-client"))
				Expect(Config.MinCLIVersion()).To(Equal("1.0.0"))
				Expect(Config.MinRecommendedCLIVersion()).To(Equal("1.0.0"))
				Expect(Config.LoggregatorEndpoint()).To(Equal("loggregator/endpoint"))
				Expect(Config.DopplerEndpoint()).To(Equal("doppler/endpoint"))
				Expect(Config.RoutingAPIEndpoint()).To(Equal("routing/endpoint"))

			})
		})
	})
})
