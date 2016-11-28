package service_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PurgeServiceOffering", func() {
	var (
		ui          *testterm.FakeUI
		configRepo  coreconfig.Repository
		serviceRepo *apifakes.FakeServiceRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		maxAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceRepo = new(apifakes.FakeServiceRepository)
		repoLocator := deps.RepoLocator.SetServiceRepository(serviceRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.PurgeServiceOffering{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = new(requirementsfakes.FakeFactory)

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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})
		})

		Context("when the -p flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service", "-p", "provider-name")
			})

			It("returns a MaxAPIVersion requirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).To(ContainElement(maxAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var runCLIErr error

		BeforeEach(func() {
			err := flagContext.Parse("service-name")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
			ui.Inputs = []string{"n"}
			serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{{}}, nil)
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		It("tries to find the service offering by label", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
			Expect(serviceRepo.FindServiceOfferingsByLabelCallCount()).To(Equal(1))
			name := serviceRepo.FindServiceOfferingsByLabelArgsForCall(0)
			Expect(name).To(Equal("service-name"))
		})

		Context("when finding the service offering succeeds", func() {
			BeforeEach(func() {
				serviceOffering := models.ServiceOffering{}
				serviceOffering.GUID = "service-offering-guid"
				serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{serviceOffering}, nil)
				ui.Inputs = []string{"n"}
			})

			It("asks the user to confirm", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"WARNING"}))
				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service offering service-name from Cloud Foundry?"}))
			})

			Context("when the user confirms", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"y"}
				})

				It("tells the user it will purge the service offering", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Purging service service-name..."}))
				})

				It("tries to purge the service offering", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(Equal(1))
				})

				Context("when purging succeeds", func() {
					BeforeEach(func() {
						serviceRepo.PurgeServiceOfferingReturns(nil)
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
					})
				})

				Context("when purging fails", func() {
					BeforeEach(func() {
						serviceRepo.PurgeServiceOfferingReturns(errors.New("purge-err"))
					})

					It("fails with error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("purge-err"))
					})
				})
			})

			Context("when the user does not confirm", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"n"}
				})

				It("does not try to purge the service offering", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
				})
			})
		})

		Context("when finding the service offering fails with an error other than 404", func() {
			BeforeEach(func() {
				serviceRepo.FindServiceOfferingsByLabelReturns([]models.ServiceOffering{}, errors.New("find-err"))
			})

			It("fails with error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("find-err"))
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
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Service offering does not exist"},
				))
			})

			It("does not try to purge the service offering", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
			})
		})

		Context("when the -p flag is passed", func() {
			var origAPIVersion string

			BeforeEach(func() {
				origAPIVersion = configRepo.APIVersion()
				configRepo.SetAPIVersion("2.46.0")

				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("service-name", "-p", "provider-name")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
				ui.Inputs = []string{"n"}
			})

			AfterEach(func() {
				configRepo.SetAPIVersion(origAPIVersion)
			})

			It("tries to find the service offering by label and provider", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceRepo.FindServiceOfferingByLabelAndProviderCallCount()).To(Equal(1))
				name, provider := serviceRepo.FindServiceOfferingByLabelAndProviderArgsForCall(0)
				Expect(name).To(Equal("service-name"))
				Expect(provider).To(Equal("provider-name"))
			})

			Context("when finding the service offering succeeds", func() {
				BeforeEach(func() {
					serviceOffering := models.ServiceOffering{}
					serviceOffering.GUID = "service-offering-guid"
					serviceRepo.FindServiceOfferingByLabelAndProviderReturns(serviceOffering, nil)
					ui.Inputs = []string{"n"}
				})

				It("asks the user to confirm", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"WARNING"}))
					Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service offering service-name from Cloud Foundry?"}))
				})

				Context("when the user confirms", func() {
					BeforeEach(func() {
						ui.Inputs = []string{"y"}
					})

					It("tells the user it will purge the service offering", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Purging service service-name..."}))
					})

					It("tries to purge the service offering", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(Equal(1))
					})

					Context("when purging succeeds", func() {
						BeforeEach(func() {
							serviceRepo.PurgeServiceOfferingReturns(nil)
						})

						It("says OK", func() {
							Expect(runCLIErr).NotTo(HaveOccurred())
							Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
						})
					})

					Context("when purging fails", func() {
						BeforeEach(func() {
							serviceRepo.PurgeServiceOfferingReturns(errors.New("purge-err"))
						})

						It("fails with error", func() {
							Expect(runCLIErr).To(HaveOccurred())
							Expect(runCLIErr.Error()).To(Equal("purge-err"))
						})
					})
				})

				Context("when the user does not confirm", func() {
					BeforeEach(func() {
						ui.Inputs = []string{"n"}
					})

					It("does not try to purge the service offering", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
					})
				})
			})

			Context("when finding the service offering fails with an error other than 404", func() {
				BeforeEach(func() {
					serviceRepo.FindServiceOfferingByLabelAndProviderReturns(models.ServiceOffering{}, errors.New("find-err"))
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("find-err"))
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
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Service offering does not exist"},
					))
				})

				It("does not try to purge the service offering", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
				})
			})
		})
	})
})
