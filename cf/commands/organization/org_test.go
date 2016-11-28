package organization_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/plugin/models"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commands/organization"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("org command", func() {
	var (
		ui             *testterm.FakeUI
		getOrgModel    *plugin_models.GetOrg_Model
		deps           commandregistry.Dependency
		reqFactory     *requirementsfakes.FakeFactory
		loginReq       *requirementsfakes.FakeRequirement
		orgRequirement *requirementsfakes.FakeOrganizationRequirement
		cmd            organization.ShowOrg
		flagContext    flags.FlagContext
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		getOrgModel = new(plugin_models.GetOrg_Model)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      testconfig.NewRepositoryWithDefaults(),
			RepoLocator: api.RepositoryLocator{},
			PluginModels: &commandregistry.PluginModels{
				Organization: getOrgModel,
			},
		}

		reqFactory = new(requirementsfakes.FakeFactory)

		loginReq = new(requirementsfakes.FakeRequirement)
		loginReq.ExecuteReturns(nil)
		reqFactory.NewLoginRequirementReturns(loginReq)

		orgRequirement = new(requirementsfakes.FakeOrganizationRequirement)
		orgRequirement.ExecuteReturns(nil)
		reqFactory.NewOrganizationRequirementReturns(orgRequirement)

		cmd = organization.ShowOrg{}
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		cmd.SetDependency(deps, false)
	})

	Describe("Requirements", func() {
		Context("when the wrong number of args are provided", func() {
			BeforeEach(func() {
				err := flagContext.Parse()
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with no args", func() {
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

			BeforeEach(func() {
				err := flagContext.Parse("my-org")
				Expect(err).NotTo(HaveOccurred())
				actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when no flags are provided", func() {
				It("returns a login requirement", func() {
					Expect(reqFactory.NewLoginRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(loginReq))
				})

				It("returns an organization requirement", func() {
					Expect(reqFactory.NewOrganizationRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(orgRequirement))
				})
			})
		})
	})

	Describe("Execute", func() {
		var (
			org        models.Organization
			executeErr error
		)

		BeforeEach(func() {
			org = models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "my-org",
					GUID: "my-org-guid",
					QuotaDefinition: models.QuotaFields{
						Name:                    "cantina-quota",
						MemoryLimit:             512,
						InstanceMemoryLimit:     256,
						RoutesLimit:             2,
						ServicesLimit:           5,
						NonBasicServicesAllowed: true,
						AppInstanceLimit:        7,
						ReservedRoutePorts:      "7",
					},
				},
				Spaces: []models.SpaceFields{
					models.SpaceFields{
						Name: "development",
						GUID: "dev-space-guid-1",
					},
					models.SpaceFields{
						Name: "staging",
						GUID: "staging-space-guid-1",
					},
				},
				Domains: []models.DomainFields{
					models.DomainFields{
						Name: "cfapps.io",
						GUID: "1111",
						OwningOrganizationGUID: "my-org-guid",
						Shared:                 true,
					},
					models.DomainFields{
						Name: "cf-app.com",
						GUID: "2222",
						OwningOrganizationGUID: "my-org-guid",
						Shared:                 false,
					},
				},
				SpaceQuotas: []models.SpaceQuota{
					{Name: "space-quota-1", GUID: "space-quota-1-guid", MemoryLimit: 512, InstanceMemoryLimit: -1},
					{Name: "space-quota-2", GUID: "space-quota-2-guid", MemoryLimit: 256, InstanceMemoryLimit: 128},
				},
			}

			orgRequirement.GetOrganizationReturns(org)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(flagContext)
		})

		Context("when logged in, and provided the name of an org", func() {
			BeforeEach(func() {
				err := flagContext.Parse("my-org")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(reqFactory, flagContext)
			})

			It("shows the org with the given name", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting info for org", "my-org", "my-user"},
					[]string{"OK"},
					[]string{"my-org"},
					[]string{"domains:", "cfapps.io", "cf-app.com"},
					[]string{"quota: ", "cantina-quota", "512M", "256M instance memory limit", "2 routes", "5 services", "paid services allowed", "7 app instance limit", "7 route ports"},
					[]string{"spaces:", "development", "staging"},
					[]string{"space quotas:", "space-quota-1", "space-quota-2"},
				))
			})

			Context("when ReservedRoutePorts is set to -1", func() {
				BeforeEach(func() {
					org.QuotaDefinition.ReservedRoutePorts = "-1"
					orgRequirement.GetOrganizationReturns(org)
				})

				It("shows unlimited route ports", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"unlimited route ports"},
					))
				})
			})

			Context("when the reserved route ports field is not provided by the CC API", func() {
				BeforeEach(func() {
					org.QuotaDefinition.ReservedRoutePorts = ""
					orgRequirement.GetOrganizationReturns(org)
				})

				It("should not display route ports", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(ui.Outputs()).NotTo(ContainSubstrings(
						[]string{"route ports"},
					))
				})
			})

			Context("when the guid flag is provided", func() {
				BeforeEach(func() {
					err := flagContext.Parse("my-org", "--guid")
					Expect(err).NotTo(HaveOccurred())
				})

				It("shows only the org guid", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"my-org-guid"},
					))
					Expect(ui.Outputs()).ToNot(ContainSubstrings(
						[]string{"Getting info for org", "my-org", "my-user"},
					))
				})
			})

			Context("when invoked by a plugin", func() {
				BeforeEach(func() {
					cmd.SetDependency(deps, true)
				})

				It("populates the plugin model", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(getOrgModel.Guid).To(Equal("my-org-guid"))
					Expect(getOrgModel.Name).To(Equal("my-org"))
					// quota
					Expect(getOrgModel.QuotaDefinition.Name).To(Equal("cantina-quota"))
					Expect(getOrgModel.QuotaDefinition.MemoryLimit).To(Equal(int64(512)))
					Expect(getOrgModel.QuotaDefinition.InstanceMemoryLimit).To(Equal(int64(256)))
					Expect(getOrgModel.QuotaDefinition.RoutesLimit).To(Equal(2))
					Expect(getOrgModel.QuotaDefinition.ServicesLimit).To(Equal(5))
					Expect(getOrgModel.QuotaDefinition.NonBasicServicesAllowed).To(BeTrue())

					// domains
					Expect(getOrgModel.Domains).To(HaveLen(2))
					Expect(getOrgModel.Domains[0].Name).To(Equal("cfapps.io"))
					Expect(getOrgModel.Domains[0].Guid).To(Equal("1111"))
					Expect(getOrgModel.Domains[0].OwningOrganizationGuid).To(Equal("my-org-guid"))
					Expect(getOrgModel.Domains[0].Shared).To(BeTrue())
					Expect(getOrgModel.Domains[1].Name).To(Equal("cf-app.com"))
					Expect(getOrgModel.Domains[1].Guid).To(Equal("2222"))
					Expect(getOrgModel.Domains[1].OwningOrganizationGuid).To(Equal("my-org-guid"))
					Expect(getOrgModel.Domains[1].Shared).To(BeFalse())

					// spaces
					Expect(getOrgModel.Spaces).To(HaveLen(2))
					Expect(getOrgModel.Spaces[0].Name).To(Equal("development"))
					Expect(getOrgModel.Spaces[0].Guid).To(Equal("dev-space-guid-1"))
					Expect(getOrgModel.Spaces[1].Name).To(Equal("staging"))
					Expect(getOrgModel.Spaces[1].Guid).To(Equal("staging-space-guid-1"))

					// space quotas
					Expect(getOrgModel.SpaceQuotas).To(HaveLen(2))
					Expect(getOrgModel.SpaceQuotas[0].Name).To(Equal("space-quota-1"))
					Expect(getOrgModel.SpaceQuotas[0].Guid).To(Equal("space-quota-1-guid"))
					Expect(getOrgModel.SpaceQuotas[0].MemoryLimit).To(Equal(int64(512)))
					Expect(getOrgModel.SpaceQuotas[0].InstanceMemoryLimit).To(Equal(int64(-1)))
					Expect(getOrgModel.SpaceQuotas[1].Name).To(Equal("space-quota-2"))
					Expect(getOrgModel.SpaceQuotas[1].Guid).To(Equal("space-quota-2-guid"))
					Expect(getOrgModel.SpaceQuotas[1].MemoryLimit).To(Equal(int64(256)))
					Expect(getOrgModel.SpaceQuotas[1].InstanceMemoryLimit).To(Equal(int64(128)))
				})
			})
		})
	})
})
