package service_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
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

var _ = Describe("UpdateUserProvidedService", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		serviceInstanceRepo *apifakes.FakeUserProvidedServiceInstanceRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement           requirements.Requirement
		minAPIVersionRequirement   requirements.Requirement
		serviceInstanceRequirement *requirementsfakes.FakeServiceInstanceRequirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceInstanceRepo = new(apifakes.FakeUserProvidedServiceInstanceRepository)
		repoLocator := deps.RepoLocator.SetUserProvidedServiceInstanceRepository(serviceInstanceRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.UpdateUserProvidedService{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)

		serviceInstanceRequirement = new(requirementsfakes.FakeServiceInstanceRequirement)
		factory.NewServiceInstanceRequirementReturns(serviceInstanceRequirement)
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
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
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
		})

		Context("when provided the -r flag", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-r", "route-service-url")
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))

				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("Option '-r'"))
				expectedRequiredVersion, err := semver.Make("2.51.0")
				Expect(err).NotTo(HaveOccurred())
				Expect(requiredVersion).To(Equal(expectedRequiredVersion))
			})
		})
	})

	Describe("Execute", func() {
		var runCLIErr error

		BeforeEach(func() {
			err := flagContext.Parse("service-instance")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		Context("when the service instance is not user-provided", func() {
			BeforeEach(func() {
				serviceInstanceRequirement.GetServiceInstanceReturns(models.ServiceInstance{
					ServicePlan: models.ServicePlanFields{
						GUID: "service-plan-guid",
					},
				})
			})

			It("fails with error", func() {
				Expect(runCLIErr).To(HaveOccurred())
			})
		})

		Context("when the service instance is user-provided", func() {
			var serviceInstance models.ServiceInstance

			BeforeEach(func() {
				serviceInstance = models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Name:   "service-instance",
						Params: map[string]interface{}{},
					},
					ServicePlan: models.ServicePlanFields{
						GUID:        "",
						Description: "service-plan-description",
					},
				}
				serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
			})

			It("tells the user it is updating the user provided service", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating user provided service service-instance in org"},
				))
			})

			It("tries to update the service instance", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.UpdateCallCount()).To(Equal(1))
				Expect(serviceInstanceRepo.UpdateArgsForCall(0)).To(Equal(serviceInstance.ServiceInstanceFields))
			})

			It("tells the user no changes were made", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"No flags specified. No changes were made."},
				))
			})

			Context("when the -p flag is passed with inline JSON", func() {
				BeforeEach(func() {
					flagContext.Parse("service-instance", "-p", `"{"some":"json"}"`)
				})

				It("tries to update the user provided service instance with the credentials", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(serviceInstanceRepo.UpdateCallCount()).To(Equal(1))
					serviceInstanceFields := serviceInstanceRepo.UpdateArgsForCall(0)
					Expect(serviceInstanceFields.Params).To(Equal(map[string]interface{}{
						"some": "json",
					}))
				})
			})

			Context("when the -p flag is passed with a file containing JSON", func() {
				BeforeEach(func() {
					tempfile, err := ioutil.TempFile("", "update-user-provided-service-test")
					Expect(err).NotTo(HaveOccurred())
					jsonData := `{"some":"json"}`
					ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)
					flagContext.Parse("service-instance", "-p", tempfile.Name())
				})

				It("tries to update the user provided service instance with the credentials", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(serviceInstanceRepo.UpdateCallCount()).To(Equal(1))
					serviceInstanceFields := serviceInstanceRepo.UpdateArgsForCall(0)
					Expect(serviceInstanceFields.Params).To(Equal(map[string]interface{}{
						"some": "json",
					}))
				})
			})

			Context("when the -p flag is passed with inline JSON", func() {
				BeforeEach(func() {
					flagContext.Parse("service-instance", "-p", `key1,key2`)
					ui.Inputs = []string{"value1", "value2"}
				})

				It("prompts the user for the values", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"key1"},
						[]string{"key2"},
					))
				})

				It("tries to update the user provided service instance with the credentials", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())

					Expect(serviceInstanceRepo.UpdateCallCount()).To(Equal(1))
					serviceInstanceFields := serviceInstanceRepo.UpdateArgsForCall(0)
					Expect(serviceInstanceFields.Params).To(Equal(map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					}))
				})
			})

			Context("when updating succeeds", func() {
				BeforeEach(func() {
					serviceInstanceRepo.UpdateReturns(nil)
				})

				It("tells the user OK", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
					))
				})

				It("prints a tip", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"TIP"},
					))
				})
			})

			Context("when updating fails", func() {
				BeforeEach(func() {
					serviceInstanceRepo.UpdateReturns(errors.New("update-err"))
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("update-err"))
				})
			})
		})
	})
})
