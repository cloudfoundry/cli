package service_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	cferrors "code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PurgeServiceInstance", func() {
	var (
		ui          *testterm.FakeUI
		configRepo  coreconfig.Repository
		serviceRepo *apifakes.FakeServiceRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		minAPIVersionRequirement requirements.Requirement
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

		cmd = &service.PurgeServiceInstance{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		minAPIVersionRequirement = &passingRequirement{}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires an argument"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))

				expectedVersion, err := semver.Make("2.36.0")
				Expect(err).NotTo(HaveOccurred())

				commandName, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(commandName).To(Equal("purge-service-instance"))
				Expect(requiredVersion).To(Equal(expectedVersion))

				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("service-instance-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		It("finds the instance by name in the service repo", func() {
			err := flagContext.Parse("service-instance-name", "-f")
			Expect(err).NotTo(HaveOccurred())
			cmd.Execute(flagContext)
			Expect(serviceRepo.FindInstanceByNameCallCount()).To(Equal(1))
		})

		Context("when the instance can be found", func() {
			var serviceInstance models.ServiceInstance

			BeforeEach(func() {
				serviceInstance = models.ServiceInstance{}
				serviceInstance.Name = "service-instance-name"
				serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
			})

			It("warns the user", func() {
				ui.Inputs = []string{"n"}
				cmd.Execute(flagContext)
				Eventually(func() []string {
					return ui.Outputs()
				}).Should(ContainSubstrings(
					[]string{"WARNING"},
				))
			})

			It("asks the user if they would like to proceed", func() {
				ui.Inputs = []string{"n"}
				cmd.Execute(flagContext)
				Eventually(func() []string { return ui.Prompts }).Should(ContainSubstrings(
					[]string{"Really purge service instance service-instance-name from Cloud Foundry?"},
				))
			})

			It("purges the service instance when the response is to proceed", func() {
				ui.Inputs = []string{"y"}
				cmd.Execute(flagContext)

				Eventually(serviceRepo.PurgeServiceInstanceCallCount).Should(Equal(1))
				Expect(serviceRepo.PurgeServiceInstanceArgsForCall(0)).To(Equal(serviceInstance))
			})

			It("does not purge the service instance when the response is not to proceed", func() {
				ui.Inputs = []string{"n"}
				cmd.Execute(flagContext)
				Consistently(serviceRepo.PurgeServiceInstanceCallCount).Should(BeZero())
			})

			Context("when force is set", func() {
				BeforeEach(func() {
					err := flagContext.Parse("service-instance-name", "-f")
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not ask the user if they would like to proceed", func() {
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Really purge service instance service-instance-name from Cloud Foundry?"},
					))
				})

				It("purges the service instance", func() {
					cmd.Execute(flagContext)
					Expect(serviceRepo.PurgeServiceInstanceCallCount()).To(Equal(1))
					Expect(serviceRepo.PurgeServiceInstanceArgsForCall(0)).To(Equal(serviceInstance))
				})
			})
		})

		Context("when the instance can not be found", func() {
			BeforeEach(func() {
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, cferrors.NewModelNotFoundError("model-type", "model-name"))
			})

			It("prints a warning", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Service instance service-instance-name not found"},
				))
			})
		})

		Context("when an error occurs fetching the instance", func() {
			var runCLIErr error

			BeforeEach(func() {
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("an-error"))
			})

			JustBeforeEach(func() {
				runCLIErr = cmd.Execute(flagContext)
			})

			It("prints a message with the error", func() {
				Expect(runCLIErr).To(HaveOccurred())
			})
		})
	})
})
