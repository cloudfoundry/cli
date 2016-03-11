package service_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PurgeServiceOffering", func() {
	var (
		ui          *testterm.FakeUI
		configRepo  core_config.Repository
		serviceRepo *testapi.FakeServiceRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		maxAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepository{}
		repoLocator := deps.RepoLocator.SetServiceRepository(serviceRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.PurgeServiceOffering{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		maxAPIVersionRequirement = &passingRequirement{Name: "max-api-version-requirement"}
		factory.NewMaxAPIVersionRequirementReturns(maxAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("service", "extra-arg")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("service")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})
		})

		Context("when the -p flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service", "-p", "provider-name")
			})

			It("returns a MaxAPIVersion requirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(actualRequirements).To(ContainElement(maxAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("service-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		It("tries to find the service offering by label", func() {
			ui.Inputs = []string{"n"}
			serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{{}}, nil)
			cmd.Execute(flagContext)
			Expect(serviceRepo.FindServiceOfferingsByLabelCallCount()).To(Equal(1))
			name := serviceRepo.FindServiceOfferingsByLabelArgsForCall(0)
			Expect(name).To(Equal("service-name"))
		})

		Context("when finding the service offering succeeds", func() {
			BeforeEach(func() {
				serviceOffering := models.ServiceOffering{}
				serviceOffering.Guid = "service-offering-guid"
				serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{serviceOffering}, nil)
			})

			It("asks the user to confirm", func() {
				ui.Inputs = []string{"n"}
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings([]string{"WARNING"}))
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service offering service-name from Cloud Foundry?"}))
			})

			Context("when the user confirms", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"y"}
				})

				It("tells the user it will purge the service offering", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings([]string{"Purging service service-name..."}))
				})

				It("tries to purge the service offering", func() {
					cmd.Execute(flagContext)
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(Equal(1))
				})

				Context("when purging succeeds", func() {
					BeforeEach(func() {
						serviceRepo.PurgeServiceOfferingReturns(nil)
					})

					It("says OK", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
					})
				})

				Context("when purging fails", func() {
					BeforeEach(func() {
						serviceRepo.PurgeServiceOfferingReturns(errors.New("purge-err"))
					})

					It("fails with error", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"purge-err"},
						))
					})
				})
			})

			Context("when the user does not confirm", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"n"}
				})

				It("does not try to purge the service offering", func() {
					cmd.Execute(flagContext)
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
				})
			})
		})

		Context("when finding the service offering fails with an error other than 404", func() {
			BeforeEach(func() {
				serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{}, errors.New("find-err"))
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when finding the service offering fails with 404 not found", func() {
			BeforeEach(func() {
				serviceRepo.FindServiceOfferingsByLabelReturns(
					[]models.ServiceOffering{{}},
					errors.NewModelNotFoundError("model-type", "find-err"),
				)
			})

			It("warns the user", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service offering does not exist"},
				))
			})

			It("does not try to purge the service offering", func() {
				cmd.Execute(flagContext)
				Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
			})
		})

		Context("when the -p flag is passed", func() {
			var origAPIVersion string
			BeforeEach(func() {
				origAPIVersion = configRepo.ApiVersion()
				configRepo.SetApiVersion("2.46.0")

				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("service-name", "-p", "provider-name")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			AfterEach(func() {
				configRepo.SetApiVersion(origAPIVersion)
			})

			It("tries to find the service offering by label and provider", func() {
				ui.Inputs = []string{"n"}
				cmd.Execute(flagContext)
				Expect(serviceRepo.FindServiceOfferingByLabelAndProviderCallCount()).To(Equal(1))
				name, provider := serviceRepo.FindServiceOfferingByLabelAndProviderArgsForCall(0)
				Expect(name).To(Equal("service-name"))
				Expect(provider).To(Equal("provider-name"))
			})

			Context("when finding the service offering succeeds", func() {
				BeforeEach(func() {
					serviceOffering := models.ServiceOffering{}
					serviceOffering.Guid = "service-offering-guid"
					serviceRepo.FindServiceOfferingByLabelAndProviderReturns(serviceOffering, nil)
				})

				It("asks the user to confirm", func() {
					ui.Inputs = []string{"n"}
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings([]string{"WARNING"}))
					Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service offering service-name from Cloud Foundry?"}))
				})

				Context("when the user confirms", func() {
					BeforeEach(func() {
						ui.Inputs = []string{"y"}
					})

					It("tells the user it will purge the service offering", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings([]string{"Purging service service-name..."}))
					})

					It("tries to purge the service offering", func() {
						cmd.Execute(flagContext)
						Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(Equal(1))
					})

					Context("when purging succeeds", func() {
						BeforeEach(func() {
							serviceRepo.PurgeServiceOfferingReturns(nil)
						})

						It("says OK", func() {
							cmd.Execute(flagContext)
							Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
						})
					})

					Context("when purging fails", func() {
						BeforeEach(func() {
							serviceRepo.PurgeServiceOfferingReturns(errors.New("purge-err"))
						})

						It("fails with error", func() {
							Expect(func() { cmd.Execute(flagContext) }).To(Panic())
							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"purge-err"},
							))
						})
					})
				})

				Context("when the user does not confirm", func() {
					BeforeEach(func() {
						ui.Inputs = []string{"n"}
					})

					It("does not try to purge the service offering", func() {
						cmd.Execute(flagContext)
						Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
					})
				})
			})

			Context("when finding the service offering fails with an error other than 404", func() {
				BeforeEach(func() {
					serviceRepo.FindServiceOfferingByLabelAndProviderReturns(models.ServiceOffering{}, errors.New("find-err"))
				})

				It("fails with error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
				})
			})

			Context("when finding the service offering fails with 404 not found", func() {
				BeforeEach(func() {
					serviceRepo.FindServiceOfferingByLabelAndProviderReturns(
						models.ServiceOffering{},
						errors.NewModelNotFoundError("model-type", "find-err"),
					)
				})

				It("warns the user", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Service offering does not exist"},
					))
				})

				It("does not try to purge the service offering", func() {
					cmd.Execute(flagContext)
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
				})
			})
		})
	})
})
