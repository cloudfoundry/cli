package commands_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	authenticationfakes "github.com/cloudfoundry/cli/cf/api/authentication/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneTimeSSHCode", func() {
	var (
		ui           *testterm.FakeUI
		configRepo   core_config.Repository
		authRepo     *authenticationfakes.FakeAuthenticationRepository
		endpointRepo *testapi.FakeEndpointRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		endpointRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint("fake-api-endpoint")
		endpointRepo = &testapi.FakeEndpointRepository{}
		repoLocator := deps.RepoLocator.SetEndpointRepository(endpointRepo)
		authRepo = &authenticationfakes.FakeAuthenticationRepository{}
		repoLocator = repoLocator.SetAuthenticationRepository(authRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &commands.OneTimeSSHCode{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		endpointRequirement = &passingRequirement{Name: "endpoint-requirement"}
		factory.NewApiEndpointRequirementReturns(endpointRequirement)
	})

	Describe("Requirements", func() {
		It("returns an EndpointRequirement", func() {
			actualRequirements, err := cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewApiEndpointRequirementCallCount()).To(Equal(1))
			Expect(actualRequirements).To(ContainElement(endpointRequirement))
		})

		Context("when not provided exactly zero args", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. No argument required"},
				))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			_, err := cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		It("tries to update the endpoint", func() {
			cmd.Execute(flagContext)
			Expect(endpointRepo.UpdateEndpointCallCount()).To(Equal(1))
			Expect(endpointRepo.UpdateEndpointArgsForCall(0)).To(Equal("fake-api-endpoint"))
		})

		Context("when updating the endpoint succeeds", func() {
			BeforeEach(func() {
				endpointRepo.UpdateEndpointReturns("updated-endpoint", nil)
			})

			It("tries to refresh the auth token", func() {
				cmd.Execute(flagContext)
				Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
			})

			Context("when refreshing the token fails with an error", func() {
				BeforeEach(func() {
					authRepo.RefreshAuthTokenReturns("", errors.New("auth-error"))
				})

				It("fails with error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Error refreshing oauth token"},
					))
				})
			})

			Context("when refreshing the token succeeds", func() {
				BeforeEach(func() {
					authRepo.RefreshAuthTokenReturns("auth-token", nil)
				})

				It("displays the token", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"auth-token"},
					))
				})
			})
		})
	})
})
