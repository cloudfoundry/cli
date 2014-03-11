package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("Login Command", func() {
	var (
		Flags        []string
		Config       configuration.ReadWriter
		ui           *testterm.FakeUI
		authRepo     *testapi.FakeAuthenticationRepository
		endpointRepo *testapi.FakeEndpointRepo
		orgRepo      *testapi.FakeOrgRepository
		spaceRepo    *testapi.FakeSpaceRepository
	)

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

		org := models.Organization{}
		org.Name = "my-new-org"
		org.Guid = "my-new-org-guid"

		orgRepo = &testapi.FakeOrgRepository{
			Organizations:          []models.Organization{org},
			FindByNameOrganization: models.Organization{},
		}

		space := models.Space{}
		space.Guid = "my-space-guid"
		space.Name = "my-space"

		spaceRepo = &testapi.FakeSpaceRepository{
			Spaces: []models.Space{space},
		}

		authRepo.GetLoginPromptsReturns.Prompts = map[string]configuration.AuthPrompt{
			"username": configuration.AuthPrompt{
				DisplayName: "Username",
				Type:        configuration.AuthPromptTypeText,
			},
			"password": configuration.AuthPrompt{
				DisplayName: "Password",
				Type:        configuration.AuthPromptTypePassword,
			},
		}
	})

	Describe("when the user passed in the skip-ssl-validation flag", func() {
		BeforeEach(func() {
			Flags = []string{"--skip-ssl-validation", "-a", "api.example.com", "-u", "username", "-p", "password"}
		})

		It("disables SSL validation in the config", func() {
			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
			Expect(Config.IsSSLDisabled()).To(BeTrue())
		})
	})

	Describe("SSL validation", func() {
		Context("when the SSL certificate is invalid", func() {
			BeforeEach(func() {
				endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://example.com", "it don't work")
				Flags = []string{"-a", "api.example.com", "-u", "username", "-p", "password"}
			})

			It("does not set the endpoint", func() {
				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Invalid SSL Cert", "https://example.com"},
					{"TIP"},
				})
			})
		})

		Context("when SSL certificate is valid", func() {
			BeforeEach(func() {
				Flags = []string{"-a", "api.example.com", "-u", "username", "-p", "password"}
			})

			It("sets the endpoint", func() {
				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
			})
		})
	})

	Context("when user is not logged in and config is empty", func() {
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

				orgRepo.Organizations = []models.Organization{org1, org2}
				spaceRepo.Spaces = []models.Space{space1, space2}
			})

			It("lets the user select an org and space by number", func() {
				OUT_OF_RANGE_CHOICE := "3"

				ui.Inputs = []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"}

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Select an org"},
					{"1. some-org"},
					{"2. my-new-org"},
					{"Select a space"},
					{"1. my-space"},
					{"2. some-space"},
				})

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))

				Expect(orgRepo.FindByNameName).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user select an org and space by name", func() {
				ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-new-org", "my-space"}

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Select an org"},
					{"1. some-org"},
					{"2. my-new-org"},
					{"Select a space"},
					{"1. my-space"},
					{"2. some-space"},
				})

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))

				Expect(orgRepo.FindByNameName).To(Equal("my-new-org"))
				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("lets the user specify an org and space using flags", func() {
				Flags = []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "my-space"}

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})

			It("doesn't ask the user for the API url if they have it in their config", func() {
				Config.SetApiEndpoint("http://api.example.com")

				Flags = []string{"-o", "my-new-org", "-s", "my-space"}
				ui.Inputs = []string{"user@example.com", "password"}

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				Expect(Config.ApiEndpoint()).To(Equal("http://api.example.com"))
				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))

				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("when there are too many orgs to show", func() {
			BeforeEach(func() {
				for i := 0; i < 60; i++ {
					id := strconv.Itoa(i)
					org := models.Organization{}
					org.Guid = "my-org-guid-" + id
					org.Name = "my-org-" + id
					orgRepo.Organizations = append(orgRepo.Organizations, org)
				}

				orgRepo.FindByNameOrganization = orgRepo.Organizations[1]

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

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
					{"my-org-2"},
				})
				Expect(orgRepo.FindByNameName).To(Equal("my-org-1"))
				Expect(Config.OrganizationFields().Guid).To(Equal("my-org-guid-1"))
			})
		})

		Describe("when there is only a single org and space", func() {
			It("does not ask the user to select an org/space", func() {
				ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}

				l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

				Expect(Config.OrganizationFields().Guid).To(Equal("my-new-org-guid"))
				Expect(Config.SpaceFields().Guid).To(Equal("my-space-guid"))
				Expect(Config.AccessToken()).To(Equal("my_access_token"))
				Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://api.example.com"))
				Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
					"username": "user@example.com",
					"password": "password",
				}))
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		It("asks the user for fields given by the login info API", func() {
			authRepo.GetLoginPromptsReturns.Prompts = map[string]configuration.AuthPrompt{
				"pin": configuration.AuthPrompt{
					DisplayName: "PIN Number",
					Type:        configuration.AuthPromptTypePassword,
				},
				"account_number": configuration.AuthPrompt{
					DisplayName: "Account Number",
					Type:        configuration.AuthPromptTypeText,
				},
				"department_number": configuration.AuthPrompt{
					DisplayName: "Dept Number",
					Type:        configuration.AuthPromptTypeText,
				},
			}

			ui.Inputs = []string{"api.example.com", "the-account-number", "the-department-number", "the-pin"}

			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Account Number>"},
				{"Dept Number>"},
			})
			testassert.SliceContains(ui.PasswordPrompts, testassert.Lines{
				{"PIN Number>"},
			})

			Expect(authRepo.AuthenticateArgs.Credentials).To(Equal(map[string]string{
				"account_number":    "the-account-number",
				"department_number": "the-department-number",
				"pin":               "the-pin",
			}))
		})

		It("enables SSL by default", func() {
			ui.Inputs = []string{"api.example.com", "the-account-number", "the-department-number", "the-pin"}
			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			Expect(Config.IsSSLDisabled()).To(BeFalse())
		})

		It("fails when the user enters invalid credentials", func() {
			authRepo.AuthError = true

			Flags = []string{"-u", "user@example.com"}
			ui.Inputs = []string{"api.example.com", "password", "password2", "password3"}

			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
			Expect(Config.OrganizationFields().Guid).To(BeEmpty())
			Expect(Config.SpaceFields().Guid).To(BeEmpty())
			Expect(Config.AccessToken()).To(BeEmpty())
			Expect(Config.RefreshToken()).To(BeEmpty())

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("fails when the /v2/info API returns an error", func() {
			endpointRepo.UpdateEndpointError = errors.NewErrorWithMessage("Server error")

			ui.Inputs = []string{"api.example.com"}

			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			Expect(Config.ApiEndpoint()).To(BeEmpty())
			Expect(Config.OrganizationFields().Guid).To(BeEmpty())
			Expect(Config.SpaceFields().Guid).To(BeEmpty())
			Expect(Config.AccessToken()).To(BeEmpty())
			Expect(Config.RefreshToken()).To(BeEmpty())

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("fails when the server response is an invalid ssl cert error", func() {
			endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://bobs-burgers.com", "SELF SIGNED SADNESS")

			ui.Inputs = []string{"bobs-burgers.com"}

			cmd := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("login", Flags), nil)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"SSL cert", "https://bobs-burgers.com"},
				{"TIP", "--skip-ssl-validation"},
			})

			Expect(ui.ShowConfigurationCalled).To(BeFalse())
		})

		It("fails when there is an error fetching the organization", func() {
			orgRepo.FindByNameErr = true

			Flags = []string{"-u", "user@example.com", "-o", "my-new-org", "-s", "my-space"}
			ui.Inputs = []string{"api.example.com", "user@example.com", "password"}

			l := NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)

			Expect(endpointRepo.UpdateEndpointReceived).To(Equal("api.example.com"))
			Expect(Config.OrganizationFields().Guid).To(BeEmpty())
			Expect(Config.SpaceFields().Guid).To(BeEmpty())
			Expect(Config.AccessToken()).To(Equal("my_access_token"))
			Expect(Config.RefreshToken()).To(Equal("my_refresh_token"))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})
	})

	Describe("updates to the config", func() {
		var l Login

		BeforeEach(func() {
			Config.SetApiEndpoint("api.the-old-endpoint.com")
			Config.SetAccessToken("the-old-access-token")
			Config.SetRefreshToken("the-old-refresh-token")
			l = NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
		})

		JustBeforeEach(func() {
			testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)
		})

		var ItShowsTheTarget = func() {
			It("shows the target", func() {
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		}

		Describe("when the user is setting an API", func() {
			BeforeEach(func() {
				l = NewLogin(ui, Config, authRepo, endpointRepo, orgRepo, spaceRepo)
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

					It("stores the API endpoint and the skip-ssl flag", func() {
						Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeTrue())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.UpdateEndpointError = errors.NewErrorWithMessage("API endpoint not found")
					})

					It("clears the entire config", func() {
						Expect(Config.ApiEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().Guid).To(BeEmpty())
						Expect(Config.SpaceFields().Guid).To(BeEmpty())
					})

					ItShowsTheTarget()
				})
			})

			Describe("when the --skip-ssl-validation flag is not provided", func() {
				Describe("setting api endpoint is successful", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
					})

					It("updates the API endpoint and enables SSL validation", func() {
						Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://api.the-server.com"))
						Expect(Config.IsSSLDisabled()).To(BeFalse())
					})
				})

				Describe("setting api endpoint failed", func() {
					BeforeEach(func() {
						Config.SetSSLDisabled(true)
						endpointRepo.UpdateEndpointError = errors.NewErrorWithMessage("API endpoint not found")
					})

					It("clears the entire config", func() {
						Expect(Config.ApiEndpoint()).To(BeEmpty())
						Expect(Config.IsSSLDisabled()).To(BeFalse())
						Expect(Config.AccessToken()).To(BeEmpty())
						Expect(Config.RefreshToken()).To(BeEmpty())
						Expect(Config.OrganizationFields().Guid).To(BeEmpty())
						Expect(Config.SpaceFields().Guid).To(BeEmpty())
					})

					ItShowsTheTarget()
				})
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
					ui.Inputs = []string{"password", "password2", "password3"}
				})

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

				ItShowsTheTarget()
			})
		})

		Describe("and the login fails to target an org", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "nonexistentorg", "-s", "my-space"}

				Config.SetSSLDisabled(true)
			})

			JustBeforeEach(func() {
				testcmd.RunCommand(l, testcmd.NewContext("login", Flags), nil)
			})

			It("does not update the api endpoint or ssl setting in the config", func() {
				Expect(Config.ApiEndpoint()).To(Equal("api.the-old-endpoint.com"))
				Expect(Config.IsSSLDisabled()).To(BeTrue())
			})

			It("clears Org, and Space in the config", func() {
				Expect(Config.OrganizationFields().Guid).To(BeEmpty())
				Expect(Config.SpaceFields().Guid).To(BeEmpty())
			})

			It("shows target", func() {
				Expect(ui.ShowConfigurationCalled).To(BeTrue())
			})
		})

		Describe("and the login fails to target a space", func() {
			BeforeEach(func() {
				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "my-new-org", "-s", "nonexistent"}

				Config.SetSSLDisabled(true)
			})

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

			ItShowsTheTarget()
		})

		Describe("and the login succeeds", func() {
			BeforeEach(func() {
				orgRepo.Organizations[0].Name = "new-org"
				orgRepo.Organizations[0].Guid = "new-org-guid"
				spaceRepo.Spaces[0].Name = "new-space"
				spaceRepo.Spaces[0].Guid = "new-space-guid"
				authRepo.AccessToken = "new_access_token"
				authRepo.RefreshToken = "new_refresh_token"

				Flags = []string{"-u", "user@example.com", "-p", "password", "-o", "new-org", "-s", "new-space"}

				Config.SetApiEndpoint("api.the-old-endpoint.com")
				Config.SetSSLDisabled(true)
			})

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

			ItShowsTheTarget()
		})
	})
})
