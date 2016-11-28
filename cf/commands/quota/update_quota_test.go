package quota_test

import (
	"code.cloudfoundry.org/cli/cf/api/quotas/quotasfakes"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	cmdsQuota "code.cloudfoundry.org/cli/cf/commands/quota"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"encoding/json"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app Command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		quotaRepo           *quotasfakes.FakeQuotaRepository
		quota               models.QuotaFields
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetQuotaRepository(quotaRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewMinAPIVersionRequirementReturns(requirements.Passing{})
		quotaRepo = new(quotasfakes.FakeQuotaRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("update-quota", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("Help text", func() {
		var usage string

		BeforeEach(func() {
			uq := &cmdsQuota.UpdateQuota{}
			up := commandregistry.CLICommandUsagePresenter(uq)
			usage = up.Usage()
		})

		It("has an instance memory flag", func() {
			Expect(usage).To(MatchRegexp(`-i\s+Maximum amount of memory an application instance can have \(e.g. 1024M, 1G, 10G\)`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-i INSTANCE_MEMORY\]`))
		})

		It("has a total memory flag", func() {
			Expect(usage).To(MatchRegexp(`-m\s+Total amount of memory \(e.g. 1024M, 1G, 10G\)`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-m TOTAL_MEMORY\]`))
		})

		It("has a new name flag", func() {
			Expect(usage).To(MatchRegexp(`-n\s+New name`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-n NEW_NAME\]`))
		})

		It("has a routes flag", func() {
			Expect(usage).To(MatchRegexp(`-r\s+Total number of routes`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-r ROUTES\]`))
		})

		It("has a service instances flag", func() {
			Expect(usage).To(MatchRegexp(`-s\s+Total number of service instances`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-s SERVICE_INSTANCES\]`))
		})

		It("has an app instances flag", func() {
			Expect(usage).To(MatchRegexp(`-a\s+Total number of application instances. -1 represents an unlimited amount.`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[-a APP_INSTANCES\]`))
		})

		It("has an allow-paid-service-plans flag", func() {
			Expect(usage).To(MatchRegexp(`--allow-paid-service-plans\s+Can provision instances of paid service plans`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\[--allow-paid-service-plans`))
		})

		It("has a disallow-paid-service-plans flag", func() {
			Expect(usage).To(MatchRegexp(`--disallow-paid-service-plans\s+Cannot provision instances of paid service plans`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\--disallow-paid-service-plans\]`))
		})

		It("has a --reserved-route-ports flag", func() {
			Expect(usage).To(MatchRegexp(`--reserved-route-ports\s+Maximum number of routes that may be created with reserved ports`))

			Expect(usage).To(MatchRegexp(`cf update-quota.*\--reserved-route-ports RESERVED_ROUTE_PORTS\]`))
		})
	})

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		})

		It("fails requirements", func() {
			Expect(runCommand("my-quota", "-m", "50G")).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			quota = models.QuotaFields{
				GUID:             "quota-guid",
				Name:             "quota-name",
				MemoryLimit:      1024,
				RoutesLimit:      111,
				ServicesLimit:    222,
				AppInstanceLimit: 333,
			}
		})

		JustBeforeEach(func() {
			quotaRepo.FindByNameReturns(quota, nil)
		})

		Context("when the -i flag is provided", func() {
			It("updates the instance memory limit", func() {
				runCommand("-i", "15G", "quota-name")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("quota-name"))
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(15360)))
			})

			It("totally accepts -1 as a value because it means unlimited", func() {
				runCommand("-i", "-1", "quota-name")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("quota-name"))
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
			})

			It("fails with usage when the value cannot be parsed", func() {
				runCommand("-m", "blasé", "le-tired")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage"},
				))
			})
		})

		Context("when the -a flag is provided", func() {
			It("updated the total number of application instances limit", func() {
				runCommand("-a", "2", "quota-name")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).AppInstanceLimit).To(Equal(2))
			})

			It("totally accepts -1 as a value because it means unlimited", func() {
				runCommand("-a", "-1", "quota-name")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).AppInstanceLimit).To(Equal(resources.UnlimitedAppInstances))
			})

			It("does not override the value if a different field is updated", func() {
				runCommand("-s", "5", "quota-name")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).AppInstanceLimit).To(Equal(333))
			})
		})

		Context("when the -m flag is provided", func() {
			It("updates the memory limit", func() {
				runCommand("-m", "15G", "quota-name")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("quota-name"))
				Expect(quotaRepo.UpdateArgsForCall(0).MemoryLimit).To(Equal(int64(15360)))
			})

			It("fails with usage when the value cannot be parsed", func() {
				runCommand("-m", "blasé", "le-tired")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage"},
				))
			})
		})

		Context("when the -n flag is provided", func() {
			It("updates the quota name", func() {
				runCommand("-n", "quota-new-name", "quota-name")

				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("quota-new-name"))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating quota", "quota-name", "as", "my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when the --reserved-route-ports flag is provided", func() {
			It("updates the route port limit", func() {
				runCommand("--reserved-route-ports", "5", "quota-name")

				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).ReservedRoutePorts).To(Equal(json.Number("5")))
			})

			It("can update the route port limit to be -1, infinity", func() {
				runCommand("--reserved-route-ports", "-1", "quota-name")

				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).ReservedRoutePorts).To(Equal(json.Number("-1")))
			})
		})

		It("updates the total allowed services", func() {
			runCommand("-s", "9000", "quota-name")
			Expect(quotaRepo.UpdateArgsForCall(0).ServicesLimit).To(Equal(9000))
		})

		It("updates the total allowed routes", func() {
			runCommand("-r", "9001", "quota-name")
			Expect(quotaRepo.UpdateArgsForCall(0).RoutesLimit).To(Equal(9001))
		})

		Context("update paid service plans", func() {
			BeforeEach(func() {
				quota.NonBasicServicesAllowed = false
			})

			It("changes to paid service plan when --allow flag is provided", func() {
				runCommand("--allow-paid-service-plans", "quota-name")
				Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeTrue())
			})

			It("shows an error when both --allow and --disallow flags are provided", func() {
				runCommand("--allow-paid-service-plans", "--disallow-paid-service-plans", "quota-name")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Both flags are not permitted"},
				))
			})

			Context("when paid services are allowed", func() {
				BeforeEach(func() {
					quota.NonBasicServicesAllowed = true
				})
				It("changes to non-paid service plan when --disallow flag is provided", func() {
					quotaRepo.FindByNameReturns(quota, nil) // updating an existing quota

					runCommand("--disallow-paid-service-plans", "quota-name")
					Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeFalse())
				})
			})
		})

		It("shows an error when updating fails", func() {
			quotaRepo.UpdateReturns(errors.New("I accidentally a quota"))
			runCommand("-m", "1M", "dead-serious")
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
		})

		It("shows a message explaining the update", func() {
			quota.Name = "i-love-ui"
			quotaRepo.FindByNameReturns(quota, nil)

			runCommand("-m", "50G", "i-love-ui")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Updating quota", "i-love-ui", "as", "my-user"},
				[]string{"OK"},
			))
		})

		It("shows the user an error when finding the quota fails", func() {
			quotaRepo.FindByNameReturns(models.QuotaFields{}, errors.New("i can't believe it's not quotas!"))

			runCommand("-m", "50Somethings", "what-could-possibly-go-wrong?")
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
		})
	})

	Describe("Requirements", func() {
		var (
			requirementsFactory *requirementsfakes.FakeFactory

			ui   *testterm.FakeUI
			cmd  commandregistry.Command
			deps commandregistry.Dependency

			quotaRepo   *quotasfakes.FakeQuotaRepository
			flagContext flags.FlagContext

			loginRequirement         requirements.Requirement
			minAPIVersionRequirement requirements.Requirement
		)

		BeforeEach(func() {
			ui = &testterm.FakeUI{}

			configRepo = testconfig.NewRepositoryWithDefaults()
			quotaRepo = new(quotasfakes.FakeQuotaRepository)
			repoLocator := deps.RepoLocator.SetQuotaRepository(quotaRepo)

			deps = commandregistry.Dependency{
				UI:          ui,
				Config:      configRepo,
				RepoLocator: repoLocator,
			}

			requirementsFactory = new(requirementsfakes.FakeFactory)

			cmd = &cmdsQuota.UpdateQuota{}
			cmd.SetDependency(deps, false)

			flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

			loginRequirement = &passingRequirement{Name: "login-requirement"}
			requirementsFactory.NewLoginRequirementReturns(loginRequirement)

			minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
			requirementsFactory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
		})

		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("quota", "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("quota")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(requirementsFactory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("does not return a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(0))
				Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
			})

			Context("when an app instance limit is passed", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					flagContext.Parse("domain-name", "-a", "2")
				})

				It("returns a MinAPIVersionRequirement as the second requirement", func() {
					actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					expectedVersion, err := semver.Make("2.33.0")
					Expect(err).NotTo(HaveOccurred())

					Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := requirementsFactory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '-a'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements[1]).To(Equal(minAPIVersionRequirement))
				})
			})

			Context("when reserved route ports limit is passed", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					flagContext.Parse("domain-name", "--reserved-route-ports", "3")
				})

				It("returns a MinAPIVersionRequirement as the second requirement", func() {
					actualRequirements, err := cmd.Requirements(requirementsFactory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					expectedVersion, err := semver.Make("2.55.0")
					Expect(err).NotTo(HaveOccurred())

					Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := requirementsFactory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '--reserved-route-ports'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements[1]).To(Equal(minAPIVersionRequirement))
				})
			})
		})
	})
})
