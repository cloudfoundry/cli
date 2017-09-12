package spacequota_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commands/spacequota"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-space-quota", func() {
	var (
		ui        *testterm.FakeUI
		quotaRepo *spacequotasfakes.FakeSpaceQuotaRepository
		config    *coreconfigfakes.FakeRepository

		loginReq         *requirementsfakes.FakeRequirement
		targetedOrgReq   *requirementsfakes.FakeTargetedOrgRequirement
		minApiVersionReq *requirementsfakes.FakeRequirement
		reqFactory       *requirementsfakes.FakeFactory

		deps        commandregistry.Dependency
		cmd         spacequota.CreateSpaceQuota
		flagContext flags.FlagContext
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		config = new(coreconfigfakes.FakeRepository)

		repoLocator := api.RepositoryLocator{}
		repoLocator = repoLocator.SetSpaceQuotaRepository(quotaRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      config,
			RepoLocator: repoLocator,
		}

		reqFactory = new(requirementsfakes.FakeFactory)

		loginReq = new(requirementsfakes.FakeRequirement)
		loginReq.ExecuteReturns(nil)
		reqFactory.NewLoginRequirementReturns(loginReq)

		targetedOrgReq = new(requirementsfakes.FakeTargetedOrgRequirement)
		targetedOrgReq.ExecuteReturns(nil)
		reqFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

		minApiVersionReq = new(requirementsfakes.FakeRequirement)
		minApiVersionReq.ExecuteReturns(nil)
		reqFactory.NewMinAPIVersionRequirementReturns(minApiVersionReq)

		cmd = spacequota.CreateSpaceQuota{}
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	Describe("Requirements", func() {
		BeforeEach(func() {
			cmd.SetDependency(deps, false)
		})
		Context("when not exactly one arg is provided", func() {
			It("fails", func() {
				flagContext.Parse()
				_, err := cmd.Requirements(reqFactory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			var actualRequirements []requirements.Requirement
			var err error

			Context("when no flags are provided", func() {
				BeforeEach(func() {
					flagContext.Parse("myquota")
					actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns a login requirement", func() {
					Expect(reqFactory.NewLoginRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(loginReq))
				})

				It("returns a targeted org requirement", func() {
					Expect(reqFactory.NewTargetedOrgRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(targetedOrgReq))
				})

				It("does not return a min api requirement", func() {
					Expect(reqFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(0))
				})
			})

			Context("when the -a flag is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("myquota", "-a", "2")
					actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns a min api version requirement", func() {
					Expect(reqFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					commandName, requiredVersion := reqFactory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(commandName).To(Equal("Option '-a'"))
					expectVersion, _ := semver.Make("2.40.0")
					Expect(requiredVersion).To(Equal(expectVersion))
					Expect(actualRequirements).To(ContainElement(minApiVersionReq))
				})
			})

			Context("when the --reserved-route-ports is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("myquota", "--reserved-route-ports", "2")
					actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
					Expect(err).NotTo(HaveOccurred())
				})

				It("return a minimum api version requirement", func() {
					Expect(reqFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					commandName, requiredVersion := reqFactory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(commandName).To(Equal("Option '--reserved-route-ports'"))
					expectVersion, _ := semver.Make("2.55.0")
					Expect(requiredVersion).To(Equal(expectVersion))
					Expect(actualRequirements).To(ContainElement(minApiVersionReq))
				})
			})
		})
	})

	Describe("Execute", func() {
		var runCLIErr error

		BeforeEach(func() {
			orgFields := models.OrganizationFields{
				Name: "my-org",
				GUID: "my-org-guid",
			}

			config.OrganizationFieldsReturns(orgFields)
			config.UsernameReturns("my-user")
		})

		JustBeforeEach(func() {
			cmd.SetDependency(deps, false)
			runCLIErr = cmd.Execute(flagContext)
		})

		Context("when creating a quota succeeds", func() {
			Context("without any flags", func() {
				BeforeEach(func() {
					flagContext.Parse("my-quota")
				})

				It("creates a quota with a given name", func() {
					Expect(quotaRepo.CreateArgsForCall(0).Name).To(Equal("my-quota"))
					Expect(quotaRepo.CreateArgsForCall(0).OrgGUID).To(Equal("my-org-guid"))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Creating space quota", "my-quota", "my-org", "my-user", "..."},
						[]string{"OK"},
					))
				})

				It("sets the instance memory limit to unlimiited", func() {
					Expect(quotaRepo.CreateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
				})

				It("sets the instance limit to unlimited", func() {
					Expect(quotaRepo.CreateArgsForCall(0).AppInstanceLimit).To(Equal(resources.UnlimitedAppInstances))
				})

				It("defaults to not allowing paid service plans", func() {
					Expect(quotaRepo.CreateArgsForCall(0).NonBasicServicesAllowed).To(BeFalse())
				})
			})

			Context("when the -m flag is provided with valid value", func() {
				BeforeEach(func() {
					flagContext.Parse("-m", "50G", "erryday makin fitty jeez")
				})

				It("sets the memory limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).MemoryLimit).To(Equal(int64(51200)))
				})
			})

			Context("when the -i flag is provided with positive value", func() {
				BeforeEach(func() {
					flagContext.Parse("-i", "50G", "erryday makin fitty jeez")
				})

				It("sets the memory limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(51200)))
				})
			})

			Context("when the -i flag is provided with -1", func() {
				BeforeEach(func() {
					flagContext.Parse("-i", "-1", "wit mah hussle")
				})

				It("accepts it as an appropriate value", func() {
					Expect(quotaRepo.CreateArgsForCall(0).InstanceMemoryLimit).To(Equal(int64(-1)))
				})
			})

			Context("when the -a flag is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("-a", "50", "my special quota")
				})

				It("sets the instance limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).AppInstanceLimit).To(Equal(50))
				})
			})

			Context("when the -r flag is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("-r", "12", "ecstatic")
				})

				It("sets the route limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).RoutesLimit).To(Equal(12))
				})
			})

			Context("when the -s flag is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("-s", "42", "black star")
				})

				It("sets the service instance limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).ServicesLimit).To(Equal(42))
				})
			})

			Context("when the --reserved-route-ports flag is provided", func() {
				BeforeEach(func() {
					flagContext.Parse("--reserved-route-ports", "5", "square quota")
				})

				It("sets the quotas TCP route limit", func() {
					Expect(quotaRepo.CreateArgsForCall(0).ReservedRoutePortsLimit).To(Equal(json.Number("5")))
				})
			})

			Context("when requesting to allow paid service plans", func() {
				BeforeEach(func() {
					flagContext.Parse("--allow-paid-service-plans", "my-for-profit-quota")
				})

				It("creates the quota with paid service plans allowed", func() {
					Expect(quotaRepo.CreateArgsForCall(0).NonBasicServicesAllowed).To(BeTrue())
				})
			})
		})

		Context("when the -i flag is provided with invalid value", func() {
			BeforeEach(func() {
				flagContext.Parse("-i", "whoops", "yo", "12")
				cmd.SetDependency(deps, false)
			})

			It("alerts the user when parsing the memory limit fails", func() {
				Expect(runCLIErr).To(HaveOccurred())
				runCLIErrStr := runCLIErr.Error()
				Expect(runCLIErrStr).To(ContainSubstring("Invalid instance memory limit: whoops"))
				Expect(runCLIErrStr).To(ContainSubstring("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
			})
		})

		Context("when the -m flag is provided with invalid value", func() {
			BeforeEach(func() {
				flagContext.Parse("-m", "whoops", "wit mah hussle")
				cmd.SetDependency(deps, false)
			})

			It("alerts the user when parsing the memory limit fails", func() {
				Expect(runCLIErr).To(HaveOccurred())
				runCLIErrStr := runCLIErr.Error()
				Expect(runCLIErrStr).To(ContainSubstring("Invalid memory limit: whoops"))
				Expect(runCLIErrStr).To(ContainSubstring("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
			})
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				flagContext.Parse("my-quota")
				quotaRepo.CreateReturns(errors.New("WHOOP THERE IT IS"))
				cmd.SetDependency(deps, false)
			})

			It("alets the user when creating the quota fails", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating space quota", "my-quota", "my-org"},
				))
				Expect(runCLIErr.Error()).To(Equal("WHOOP THERE IT IS"))
			})
		})

		Context("when the quota already exists", func() {
			BeforeEach(func() {
				flagContext.Parse("my-quota")
				quotaRepo.CreateReturns(errors.NewHTTPError(400, errors.QuotaDefinitionNameTaken, "Quota Definition is taken: quota-sct"))
			})

			It("warns the user", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"already exists"}))
			})
		})
	})
})
