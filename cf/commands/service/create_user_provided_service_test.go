package service_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
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

var _ = Describe("CreateUserProvidedService", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		serviceInstanceRepo *testapi.FakeUserProvidedServiceInstanceRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceInstanceRepo = &testapi.FakeUserProvidedServiceInstanceRepository{}
		repoLocator := deps.RepoLocator.SetUserProvidedServiceInstanceRepository(serviceInstanceRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.CreateUserProvidedService{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)

		targetedSpaceRequirement = &passingRequirement{Name: "targeted-space-requirement"}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "extra-arg")
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
				flagContext.Parse("service-instance")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})
		})

		Context("when provided the -r flag", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-r", "route-service-url")
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
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
		BeforeEach(func() {
			err := flagContext.Parse("service-instance")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		It("tells the user it will create the user provided service", func() {
			cmd.Execute(flagContext)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating user provided service service-instance in org"},
			))
		})

		It("tries to create the user provided service instance", func() {
			cmd.Execute(flagContext)
			Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
			name, drainURL, routeServiceURL, credentialsMap := serviceInstanceRepo.CreateArgsForCall(0)
			Expect(name).To(Equal("service-instance"))
			Expect(drainURL).To(Equal(""))
			Expect(routeServiceURL).To(Equal(""))
			Expect(credentialsMap).To(Equal(map[string]interface{}{}))
		})

		Context("when creating the user provided service instance succeeds", func() {
			BeforeEach(func() {
				serviceInstanceRepo.CreateReturns(nil)
			})

			It("tells the user OK", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
				))
			})
		})

		Context("when creating the user provided service instance fails", func() {
			BeforeEach(func() {
				serviceInstanceRepo.CreateReturns(errors.New("create-err"))
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"create-err"},
				))
			})
		})

		Context("when the -l flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-l", "drain-url")
			})

			It("tries to create the user provided service instance with the drain url", func() {
				cmd.Execute(flagContext)
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, drainURL, _, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(drainURL).To(Equal("drain-url"))
			})
		})

		Context("when the -r flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-r", "route-service-url")
			})

			It("tries to create the user provided service instance with the route service url", func() {
				cmd.Execute(flagContext)
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, routeServiceURL, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(routeServiceURL).To(Equal("route-service-url"))
			})
		})

		Context("when the -p flag is passed with inline JSON", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-p", `"{"some":"json"}"`)
			})

			It("tries to create the user provided service instance with the credentials", func() {
				cmd.Execute(flagContext)
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
					"some": "json",
				}))
			})
		})

		Context("when the -p flag is passed with a file containing JSON", func() {
			BeforeEach(func() {
				tempfile, err := ioutil.TempFile("", "create-user-provided-service-test")
				Expect(err).NotTo(HaveOccurred())
				jsonData := `{"some":"json"}`
				ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)
				flagContext.Parse("service-instance", "-p", tempfile.Name())
			})

			It("tries to create the user provided service instance with the credentials", func() {
				cmd.Execute(flagContext)
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
					"some": "json",
				}))
			})
		})

		Context("when the -p flag is passed with inline JSON", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-p", `key1,key2`)
			})

			It("prompts the user for the values", func() {
				ui.Inputs = []string{"value1", "value2"}
				cmd.Execute(flagContext)
				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"key1"},
					[]string{"key2"},
				))
			})

			It("tries to create the user provided service instance with the credentials", func() {
				ui.Inputs = []string{"value1", "value2"}
				cmd.Execute(flagContext)

				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				}))
			})
		})
	})
})
