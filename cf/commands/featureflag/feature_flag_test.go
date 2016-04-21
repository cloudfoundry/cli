package featureflag_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/featureflags/featureflagsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("feature-flag command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          coreconfig.Repository
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("feature-flag").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("feature-flag", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("foo")).ToNot(HavePassedRequirements())
		})

		It("requires the user to provide a feature flag", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			flag := models.FeatureFlag{
				Name:    "route_creation",
				Enabled: false,
			}
			flagRepo.FindByNameReturns(flag, nil)
		})

		It("lists the state of the specified feature flag", func() {
			runCommand("route_creation")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Retrieving status of route_creation as my-user..."},
				[]string{"Feature", "State"},
				[]string{"route_creation", "disabled"},
			))
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("An error occurred."))
			})

			It("fails with an error", func() {
				runCommand("route_creation")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An error occurred."},
				))
			})
		})
	})
})
