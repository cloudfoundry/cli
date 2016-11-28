package spacequota_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("update-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *spacequotasfakes.FakeSpaceQuotaRepository
		requirementsFactory *requirementsfakes.FakeFactory

		quota            models.SpaceQuota
		quotaPaidService models.SpaceQuota
		configRepo       coreconfig.Repository
		deps             commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-space-quota").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("update-space-quota", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-quota", "-m", "50G")).NotTo(HavePassedRequirements())
		})

		It("fails when the user does not have an org targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			orgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			orgReq.ExecuteReturns(errors.New("not targeting org"))
			requirementsFactory.NewTargetedOrgRequirementReturns(orgReq)
			Expect(runCommand()).NotTo(HavePassedRequirements())
			Expect(runCommand("my-quota", "-m", "50G")).NotTo(HavePassedRequirements())
		})

		It("fails with usage if space quota name is not provided", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			runCommand()

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		Context("the minimum API version requirement", func() {
			BeforeEach(func() {
				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
				requirementsFactory.NewMinAPIVersionRequirementReturns(requirements.Failing{Message: "not min api"})
			})

			It("fails when the -a option is provided", func() {
				Expect(runCommand("my-quota", "-a", "10")).To(BeFalse())
				Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				option, version := requirementsFactory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(option).To(Equal("Option '-a'"))
				Expect(version).To(Equal(cf.SpaceAppInstanceLimitMinimumAPIVersion))
			})

			It("does not fail when the -a option is not provided", func() {
				Expect(runCommand("my-quota", "-m", "10G")).To(BeTrue())
			})
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			quota = models.SpaceQuota{
				GUID:                    "my-quota-guid",
				Name:                    "my-quota",
				MemoryLimit:             1024,
				InstanceMemoryLimit:     512,
				RoutesLimit:             111,
				ServicesLimit:           222,
				AppInstanceLimit:        333,
				NonBasicServicesAllowed: false,
				OrgGUID:                 "my-org-guid",
			}

			quotaPaidService = models.SpaceQuota{NonBasicServicesAllowed: true}

			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			requirementsFactory.NewMinAPIVersionRequirementReturns(requirements.Passing{})
		})

		JustBeforeEach(func() {
			quotaRepo.FindByNameReturns(quota, nil)
		})

		Context("when the -m flag is provided", func() {
			It("updates the memory limit", func() {
				runCommand("-m", "15G", "my-quota")
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("my-quota"))
				Expect(quotaRepo.UpdateArgsForCall(0).MemoryLimit).To(Equal(int64(15360)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-m", "whoops", "wit mah hussle", "my-org")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the -i flag is provided", func() {
			It("sets the memory limit", func() {
				runCommand("-i", "50G", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(51200)))
			})

			It("sets the memory limit to -1", func() {
				runCommand("-i", "-1", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
			})

			It("alerts the user when parsing the memory limit fails", func() {
				runCommand("-i", "whoops", "my-quota")
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the -a flag is provided", func() {
			It("sets the instance limit", func() {
				runCommand("-a", "50", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).AppInstanceLimit).To(Equal(50))
			})

			It("does not override the value if it's not provided", func() {
				runCommand("-s", "5", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).AppInstanceLimit).To(Equal(333))
			})
		})

		Context("when the -r flag is provided", func() {
			It("sets the route limit", func() {
				runCommand("-r", "12", "ecstatic")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).RoutesLimit).To(Equal(12))
			})
		})

		Context("when the -s flag is provided", func() {
			It("sets the service instance limit", func() {
				runCommand("-s", "42", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).ServicesLimit).To(Equal(42))
			})
		})

		Context("when the -n flag is provided", func() {
			It("sets the service instance name", func() {
				runCommand("-n", "foo", "my-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).Name).To(Equal("foo"))
			})
		})

		Context("when --allow-non-basic-services is provided", func() {
			It("updates the quota to allow paid service plans", func() {
				runCommand("--allow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeTrue())
			})
		})

		Context("when --disallow-non-basic-services is provided", func() {
			It("updates the quota to disallow paid service plans", func() {
				quotaRepo.FindByNameReturns(quotaPaidService, nil)

				runCommand("--disallow-paid-service-plans", "my-for-profit-quota")
				Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
				Expect(quotaRepo.UpdateArgsForCall(0).NonBasicServicesAllowed).To(BeFalse())
			})
		})

		Context("when --reserved-route-ports is provided", func() {
			DescribeTable("updates the quota to the given number of reserved route ports",
				func(numberOfReservedRoutes string) {
					quotaRepo.FindByNameReturns(quotaPaidService, nil)

					runCommand("--reserved-route-ports", numberOfReservedRoutes, "my-for-profit-quota")
					Expect(quotaRepo.UpdateCallCount()).To(Equal(1))
					Expect(quotaRepo.UpdateArgsForCall(0).ReservedRoutePortsLimit).To(Equal(json.Number(numberOfReservedRoutes)))
				},
				Entry("for positive values", "42"),
				Entry("for 0", "0"),
				Entry("for -1", "-1"),
			)
		})

		Context("when updating a quota returns an error", func() {
			It("alerts the user when creating the quota fails", func() {
				quotaRepo.UpdateReturns(errors.New("WHOOP THERE IT IS"))
				runCommand("my-quota")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating space quota", "my-quota", "my-user"},
					[]string{"FAILED"},
				))
			})

			It("fails if the allow and disallow flag are both passed", func() {
				runCommand("--disallow-paid-service-plans", "--allow-paid-service-plans", "my-for-profit-quota")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
