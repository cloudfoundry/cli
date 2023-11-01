package service_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver/v4"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateUserProvidedService", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		serviceInstanceRepo *apifakes.FakeUserProvidedServiceInstanceRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceInstanceRepo = new(apifakes.FakeUserProvidedServiceInstanceRepository)
		repoLocator := deps.RepoLocator.SetUserProvidedServiceInstanceRepository(serviceInstanceRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.CreateUserProvidedService{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = new(requirementsfakes.FakeFactory)

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

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})
		})

		Context("when provided the -t flag", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-t", "tag,a,service")
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))

				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("Option '-t'"))
				expectedRequiredVersion, err := semver.Make("2.104.0")
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

		It("tells the user it will create the user provided service", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating user provided service service-instance in org"},
			))
		})

		It("tries to create the user provided service instance", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
			Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
			name, drainURL, routeServiceURL, credentialsMap, tags := serviceInstanceRepo.CreateArgsForCall(0)
			Expect(name).To(Equal("service-instance"))
			Expect(drainURL).To(Equal(""))
			Expect(routeServiceURL).To(Equal(""))
			Expect(credentialsMap).To(Equal(map[string]interface{}{}))
			Expect(tags).To(Equal([]string{}))
		})

		Context("when creating the user provided service instance succeeds", func() {
			BeforeEach(func() {
				serviceInstanceRepo.CreateReturns(nil)
			})

			It("tells the user OK", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"OK"},
				))
			})
		})

		Context("when creating the user provided service instance fails", func() {
			BeforeEach(func() {
				serviceInstanceRepo.CreateReturns(errors.New("create-err"))
			})

			It("fails with error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("create-err"))
			})
		})

		Context("when the -l flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-l", "drain-url")
			})

			It("tries to create the user provided service instance with the drain url", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, drainURL, _, _, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(drainURL).To(Equal("drain-url"))
			})
		})

		Context("when the -r flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-r", "route-service-url")
			})

			It("tries to create the user provided service instance with the route service url", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, routeServiceURL, _, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(routeServiceURL).To(Equal("route-service-url"))
			})
		})

		Context("when the -p flag is passed with inline JSON", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-p", `"{"some":"json"}"`)
			})

			It("tries to create the user provided service instance with the credentials", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
					"some": "json",
				}))
			})
		})

		Context("when the -p flag is passed with a file containing JSON", func() {
			var filename string

			BeforeEach(func() {
				tempfile, err := ioutil.TempFile("", "create-user-provided-service-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(tempfile.Close()).NotTo(HaveOccurred())
				filename = tempfile.Name()

				jsonData := `{"some":"json"}`
				ioutil.WriteFile(filename, []byte(jsonData), os.ModePerm)
				flagContext.Parse("service-instance", "-p", filename)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(filename)).NotTo(HaveOccurred())
			})

			It("tries to create the user provided service instance with the credentials", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
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

			It("tries to create the user provided service instance with the credentials", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())

				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, credentialsMap, _ := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(credentialsMap).To(Equal(map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				}))
			})
		})

		Context("when passing in tags", func() {
			BeforeEach(func() {
				flagContext.Parse("service-instance", "-t", "tag1, tag2, tag3, tag4")
			})

			It("successfully creates a service instance and passes the tags as json", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceInstanceRepo.CreateCallCount()).To(Equal(1))
				_, _, _, _, tags := serviceInstanceRepo.CreateArgsForCall(0)
				Expect(tags).To(ConsistOf("tag1", "tag2", "tag3", "tag4"))
			})
		})
	})
})
