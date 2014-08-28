package featureflag_test

import (
	"errors"

	fakeflag "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/featureflag"
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
		flagRepo            *fakeflag.FakeFeatureFlagRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewShowFeatureFlag(ui, testconfig.NewRepositoryWithDefaults(), flagRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
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
