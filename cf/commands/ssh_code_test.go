package commands_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneTimeSSHCode", func() {
	var (
		ui           *testterm.FakeUI
		configRepo   coreconfig.Repository
		authRepo     *authenticationfakes.FakeRepository
		endpointRepo *coreconfigfakes.FakeEndpointRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		endpointRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAPIEndpoint("fake-api-endpoint")
		endpointRepo = new(coreconfigfakes.FakeEndpointRepository)
		repoLocator := deps.RepoLocator.SetEndpointRepository(endpointRepo)
		authRepo = new(authenticationfakes.FakeRepository)
		repoLocator = repoLocator.SetAuthenticationRepository(authRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &commands.OneTimeSSHCode{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		endpointRequirement = &passingRequirement{Name: "endpoint-requirement"}
		factory.NewAPIEndpointRequirementReturns(endpointRequirement)
	})

	Describe("Requirements", func() {
		It("returns an EndpointRequirement", func() {
			actualRequirements, err := cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewAPIEndpointRequirementCallCount()).To(Equal(1))
			Expect(actualRequirements).To(ContainElement(endpointRequirement))
		})

		Context("when not provided exactly zero args", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("fails with usage", func() {
				var firstErr error

				reqs, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				for _, req := range reqs {
					err := req.Execute()
					if err != nil {
						firstErr = err
						break
					}
				}

				Expect(firstErr.Error()).To(ContainSubstring("Incorrect Usage. No argument required"))
			})
		})
	})

	Describe("Execute", func() {
		var runCLIerr error

		BeforeEach(func() {
			cmd.Requirements(factory, flagContext)

			endpointRepo.GetCCInfoReturns(
				&coreconfig.CCInfo{
					LoggregatorEndpoint: "loggregator/endpoint",
				},
				"some-endpoint",
				nil,
			)
		})

		JustBeforeEach(func() {
			runCLIerr = cmd.Execute(flagContext)
		})

		It("tries to update the endpoint", func() {
			Expect(runCLIerr).NotTo(HaveOccurred())
			Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
			Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("fake-api-endpoint"))
		})

		Context("when updating the endpoint succeeds", func() {
			ccInfo := &coreconfig.CCInfo{
				APIVersion:               "some-version",
				AuthorizationEndpoint:    "auth/endpoint",
				LoggregatorEndpoint:      "loggregator/endpoint",
				MinCLIVersion:            "min-cli-version",
				MinRecommendedCLIVersion: "min-rec-cli-version",
				SSHOAuthClient:           "some-client",
				RoutingAPIEndpoint:       "routing/endpoint",
			}
			BeforeEach(func() {
				endpointRepo.GetCCInfoReturns(
					ccInfo,
					"updated-endpoint",
					nil,
				)
			})

			It("tries to refresh the auth token", func() {
				Expect(runCLIerr).NotTo(HaveOccurred())
				Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
			})

			Context("when refreshing the token fails with an error", func() {
				BeforeEach(func() {
					authRepo.RefreshAuthTokenReturns("", errors.New("auth-error"))
				})

				It("fails with error", func() {
					Expect(runCLIerr).To(HaveOccurred())
					Expect(runCLIerr.Error()).To(Equal("Error refreshing oauth token: auth-error"))
				})
			})

			Context("when refreshing the token succeeds", func() {
				BeforeEach(func() {
					authRepo.RefreshAuthTokenReturns("auth-token", nil)
				})

				It("tries to get the ssh-code", func() {
					Expect(runCLIerr).NotTo(HaveOccurred())
					Expect(authRepo.AuthorizeCallCount()).To(Equal(1))
					Expect(authRepo.AuthorizeArgsForCall(0)).To(Equal("auth-token"))
				})

				Context("when getting the ssh-code succeeds", func() {
					BeforeEach(func() {
						authRepo.AuthorizeReturns("some-code", nil)
					})

					It("displays the token", func() {
						Expect(runCLIerr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"some-code"},
						))
					})
				})

				Context("when getting the ssh-code fails", func() {
					BeforeEach(func() {
						authRepo.AuthorizeReturns("", errors.New("auth-err"))
					})

					It("fails with error", func() {
						Expect(runCLIerr).To(HaveOccurred())
						Expect(runCLIerr.Error()).To(Equal("Error getting SSH code: auth-err"))
					})
				})
			})
		})
	})
})
