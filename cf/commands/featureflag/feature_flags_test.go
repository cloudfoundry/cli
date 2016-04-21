package featureflag_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/featureflags/featureflagsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/featureflag"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("feature-flags command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          coreconfig.Repository
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("feature-flags").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("feature-flags", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		Context("when arguments are provided", func() {
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &featureflag.ListFeatureFlags{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(requirementsFactory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			flags := []models.FeatureFlag{
				models.FeatureFlag{
					Name:         "user_org_creation",
					Enabled:      true,
					ErrorMessage: "error",
				},
				models.FeatureFlag{
					Name:    "private_domain_creation",
					Enabled: false,
				},
				models.FeatureFlag{
					Name:    "app_bits_upload",
					Enabled: true,
				},
				models.FeatureFlag{
					Name:    "app_scaling",
					Enabled: true,
				},
				models.FeatureFlag{
					Name:    "route_creation",
					Enabled: false,
				},
			}
			flagRepo.ListReturns(flags, nil)
		})

		It("lists the state of all feature flags", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Retrieving status of all flagged features as my-user..."},
				[]string{"Feature", "State"},
				[]string{"user_org_creation", "enabled"},
				[]string{"private_domain_creation", "disabled"},
				[]string{"app_bits_upload", "enabled"},
				[]string{"app_scaling", "enabled"},
				[]string{"route_creation", "disabled"},
			))
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				flagRepo.ListReturns(nil, errors.New("An error occurred."))
			})

			It("fails with an error", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An error occurred."},
				))
			})
		})
	})
})
