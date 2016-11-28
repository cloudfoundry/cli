package space_test

import (
	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"

	"code.cloudfoundry.org/cli/plugin/models"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commands/space"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("space command", func() {
	var (
		ui               *testterm.FakeUI
		loginReq         *requirementsfakes.FakeRequirement
		targetedOrgReq   *requirementsfakes.FakeTargetedOrgRequirement
		reqFactory       *requirementsfakes.FakeFactory
		deps             commandregistry.Dependency
		cmd              space.ShowSpace
		flagContext      flags.FlagContext
		getSpaceModel    *plugin_models.GetSpace_Model
		spaceRequirement *requirementsfakes.FakeSpaceRequirement
		quotaRepo        *spacequotasfakes.FakeSpaceQuotaRepository
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		repoLocator := api.RepositoryLocator{}
		repoLocator = repoLocator.SetSpaceQuotaRepository(quotaRepo)
		getSpaceModel = new(plugin_models.GetSpace_Model)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      testconfig.NewRepositoryWithDefaults(),
			RepoLocator: repoLocator,
			PluginModels: &commandregistry.PluginModels{
				Space: getSpaceModel,
			},
		}

		reqFactory = new(requirementsfakes.FakeFactory)

		loginReq = new(requirementsfakes.FakeRequirement)
		loginReq.ExecuteReturns(nil)
		reqFactory.NewLoginRequirementReturns(loginReq)

		targetedOrgReq = new(requirementsfakes.FakeTargetedOrgRequirement)
		targetedOrgReq.ExecuteReturns(nil)
		reqFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

		spaceRequirement = new(requirementsfakes.FakeSpaceRequirement)
		spaceRequirement.ExecuteReturns(nil)
		reqFactory.NewSpaceRequirementReturns(spaceRequirement)

		cmd = space.ShowSpace{}
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

			Context("when no flags are provided", func() {
				BeforeEach(func() {
					err := flagContext.Parse("my-space")
					Expect(err).NotTo(HaveOccurred())
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

				It("returns a space requirement", func() {
					Expect(reqFactory.NewSpaceRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(spaceRequirement))
				})
			})
		})
	})

	Describe("Execute", func() {
		var (
			space      models.Space
			spaceQuota models.SpaceQuota
			executeErr error
		)

		BeforeEach(func() {
			org := models.OrganizationFields{
				Name: "my-org",
				GUID: "my-org-guid",
			}

			app := models.ApplicationFields{
				Name: "app1",
				GUID: "app1-guid",
			}

			apps := []models.ApplicationFields{app}

			domain := models.DomainFields{
				Name: "domain1",
				GUID: "domain1-guid",
			}

			domains := []models.DomainFields{domain}

			serviceInstance := models.ServiceInstanceFields{
				Name: "service1",
				GUID: "service1-guid",
			}
			services := []models.ServiceInstanceFields{serviceInstance}

			securityGroup1 := models.SecurityGroupFields{Name: "Nacho Security", Rules: []map[string]interface{}{
				{"protocol": "all", "destination": "0.0.0.0-9.255.255.255", "log": true, "IntTest": 1000},
			}}
			securityGroup2 := models.SecurityGroupFields{Name: "Nacho Prime", Rules: []map[string]interface{}{
				{"protocol": "udp", "ports": "8080-9090", "destination": "198.41.191.47/1"},
			}}
			securityGroups := []models.SecurityGroupFields{securityGroup1, securityGroup2}

			space = models.Space{
				SpaceFields: models.SpaceFields{
					Name: "whose-space-is-it-anyway",
					GUID: "whose-space-is-it-anyway-guid",
				},
				Organization:     org,
				Applications:     apps,
				Domains:          domains,
				ServiceInstances: services,
				SecurityGroups:   securityGroups,
				SpaceQuotaGUID:   "runaway-guid",
			}

			spaceRequirement.GetSpaceReturns(space)

			spaceQuota = models.SpaceQuota{
				Name:                    "runaway",
				GUID:                    "runaway-guid",
				MemoryLimit:             102400,
				InstanceMemoryLimit:     -1,
				RoutesLimit:             111,
				ServicesLimit:           222,
				NonBasicServicesAllowed: false,
				AppInstanceLimit:        7,
				ReservedRoutePortsLimit: "7",
			}

			quotaRepo.FindByGUIDReturns(spaceQuota, nil)
		})

		JustBeforeEach(func() {
			executeErr = cmd.Execute(flagContext)
		})

		Context("when logged in and an org is targeted", func() {
			BeforeEach(func() {
				err := flagContext.Parse("my-space")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(reqFactory, flagContext)
			})

			Context("when the guid flag is passed", func() {
				BeforeEach(func() {
					err := flagContext.Parse("my-space", "--guid")
					Expect(err).NotTo(HaveOccurred())
				})

				It("shows only the space guid", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"whose-space-is-it-anyway-guid"},
					))

					Expect(ui.Outputs()).ToNot(ContainSubstrings(
						[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
					))
				})
			})

			Context("when the security-group-rules flag is passed", func() {
				BeforeEach(func() {
					err := flagContext.Parse("my-space", "--security-group-rules")
					Expect(err).NotTo(HaveOccurred())
				})
				It("it shows space information and security group rules", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Getting rules for the security group", "Nacho Security"},
						[]string{"protocol", "all"},
						[]string{"destination", "0.0.0.0-9.255.255.255"},
						[]string{"Getting rules for the security group", "Nacho Prime"},
						[]string{"protocol", "udp"},
						[]string{"log", "true"},
						[]string{"IntTest", "1000"},
						[]string{"ports", "8080-9090"},
						[]string{"destination", "198.41.191.47/1"},
					))
				})
			})

			Context("when the space has a space quota", func() {
				It("shows information about the given space", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
						[]string{"OK"},
						[]string{"whose-space-is-it-anyway"},
						[]string{"Org", "my-org"},
						[]string{"Apps", "app1"},
						[]string{"Domains", "domain1"},
						[]string{"Services", "service1"},
						[]string{"Security Groups", "Nacho Security", "Nacho Prime"},
						[]string{"Space Quota", "runaway (100G memory limit, unlimited instance memory limit, 111 routes, 222 services, paid services disallowed, 7 app instance limit, 7 route ports)"},
					))
				})

				Context("when the route ports limit is -1", func() {
					BeforeEach(func() {
						spaceQuota.ReservedRoutePortsLimit = "-1"
						quotaRepo.FindByGUIDReturns(spaceQuota, nil)
					})

					It("displays unlimited as the route ports limit", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"unlimited route ports"},
						))
					})
				})

				Context("when the reserved route ports field is not provided by the CC API", func() {
					BeforeEach(func() {
						spaceQuota.ReservedRoutePortsLimit = ""
						quotaRepo.FindByGUIDReturns(spaceQuota, nil)
					})

					It("should not display route ports", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).NotTo(ContainSubstrings(
							[]string{"route ports"},
						))
					})
				})

				Context("when the app instance limit is -1", func() {
					BeforeEach(func() {
						spaceQuota.AppInstanceLimit = -1
						quotaRepo.FindByGUIDReturns(spaceQuota, nil)
					})

					It("displays unlimited as the app instance limit", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"unlimited app instance limit"},
						))
					})
				})
			})

			Context("when the space does not have a space quota", func() {
				BeforeEach(func() {
					space.SpaceQuotaGUID = ""
					spaceRequirement.GetSpaceReturns(space)
				})

				It("shows information without a space quota", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(quotaRepo.FindByGUIDCallCount()).To(Equal(0))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Getting info for space", "whose-space-is-it-anyway", "my-org", "my-user"},
						[]string{"OK"},
						[]string{"whose-space-is-it-anyway"},
						[]string{"Org", "my-org"},
						[]string{"Apps", "app1"},
						[]string{"Domains", "domain1"},
						[]string{"Services", "service1"},
						[]string{"Security Groups", "Nacho Security", "Nacho Prime"},
						[]string{"Space Quota"},
					))
				})
			})

			Context("When called as a plugin", func() {
				BeforeEach(func() {
					cmd.SetDependency(deps, true)
				})

				It("Fills in the PluginModel", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(getSpaceModel.Name).To(Equal("whose-space-is-it-anyway"))
					Expect(getSpaceModel.Guid).To(Equal("whose-space-is-it-anyway-guid"))

					Expect(getSpaceModel.Organization.Name).To(Equal("my-org"))
					Expect(getSpaceModel.Organization.Guid).To(Equal("my-org-guid"))

					Expect(getSpaceModel.Applications).To(HaveLen(1))
					Expect(getSpaceModel.Applications[0].Name).To(Equal("app1"))
					Expect(getSpaceModel.Applications[0].Guid).To(Equal("app1-guid"))

					Expect(getSpaceModel.Domains).To(HaveLen(1))
					Expect(getSpaceModel.Domains[0].Name).To(Equal("domain1"))
					Expect(getSpaceModel.Domains[0].Guid).To(Equal("domain1-guid"))

					Expect(getSpaceModel.ServiceInstances).To(HaveLen(1))
					Expect(getSpaceModel.ServiceInstances[0].Name).To(Equal("service1"))
					Expect(getSpaceModel.ServiceInstances[0].Guid).To(Equal("service1-guid"))

					Expect(getSpaceModel.SecurityGroups).To(HaveLen(2))
					Expect(getSpaceModel.SecurityGroups[0].Name).To(Equal("Nacho Security"))
					Expect(getSpaceModel.SecurityGroups[0].Rules).To(HaveLen(1))
					Expect(getSpaceModel.SecurityGroups[0].Rules[0]).To(HaveLen(4))
					val := getSpaceModel.SecurityGroups[0].Rules[0]["protocol"]
					Expect(val).To(Equal("all"))
					val = getSpaceModel.SecurityGroups[0].Rules[0]["destination"]
					Expect(val).To(Equal("0.0.0.0-9.255.255.255"))

					Expect(getSpaceModel.SecurityGroups[1].Name).To(Equal("Nacho Prime"))
					Expect(getSpaceModel.SecurityGroups[1].Rules).To(HaveLen(1))
					Expect(getSpaceModel.SecurityGroups[1].Rules[0]).To(HaveLen(3))
					val = getSpaceModel.SecurityGroups[1].Rules[0]["protocol"]
					Expect(val).To(Equal("udp"))
					val = getSpaceModel.SecurityGroups[1].Rules[0]["destination"]
					Expect(val).To(Equal("198.41.191.47/1"))
					val = getSpaceModel.SecurityGroups[1].Rules[0]["ports"]
					Expect(val).To(Equal("8080-9090"))

					Expect(getSpaceModel.SpaceQuota.Name).To(Equal("runaway"))
					Expect(getSpaceModel.SpaceQuota.Guid).To(Equal("runaway-guid"))
					Expect(getSpaceModel.SpaceQuota.MemoryLimit).To(Equal(int64(102400)))
					Expect(getSpaceModel.SpaceQuota.InstanceMemoryLimit).To(Equal(int64(-1)))
					Expect(getSpaceModel.SpaceQuota.RoutesLimit).To(Equal(111))
					Expect(getSpaceModel.SpaceQuota.ServicesLimit).To(Equal(222))
					Expect(getSpaceModel.SpaceQuota.NonBasicServicesAllowed).To(BeFalse())
				})
			})
		})
	})
})
