package servicebroker_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/servicebroker"
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

var _ = Describe("CreateServiceBroker", func() {
	var (
		ui                *testterm.FakeUI
		configRepo        core_config.Repository
		serviceBrokerRepo *testapi.FakeServiceBrokerRepository

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
		serviceBrokerRepo = &testapi.FakeServiceBrokerRepository{}
		repoLocator := deps.RepoLocator.SetServiceBrokerRepository(serviceBrokerRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &servicebroker.CreateServiceBroker{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{Name: "targeted-space-requirement"}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	It("has an alias of `csb`", func() {
		cmd := &servicebroker.CreateServiceBroker{}

		Expect(cmd.MetaData().ShortName).To(Equal("csb"))
	})

	Describe("Requirements", func() {
		Context("when not provided exactly four args", func() {
			BeforeEach(func() {
				flagContext.Parse("service-broker")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires SERVICE_BROKER, USERNAME, PASSWORD, URL as arguments"},
				))
			})
		})

		Context("when provided exactly four args", func() {
			BeforeEach(func() {
				flagContext.Parse("service-broker", "username", "password", "url")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})
		})

		Context("when the --space-scoped flag is provided", func() {
			BeforeEach(func() {
				flagContext.Parse("service-broker", "username", "password", "url", "--space-scoped")
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("service-broker", "username", "password", "url")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		It("tells the user it is creating the service broker", func() {
			cmd.Execute(flagContext)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service broker", "service-broker", "my-user"},
				[]string{"OK"},
			))
		})

		It("tries to create the service broker", func() {
			cmd.Execute(flagContext)
			Expect(serviceBrokerRepo.CreateCallCount()).To(Equal(1))
			name, url, username, password, spaceGUID := serviceBrokerRepo.CreateArgsForCall(0)
			Expect(name).To(Equal("service-broker"))
			Expect(url).To(Equal("url"))
			Expect(username).To(Equal("username"))
			Expect(password).To(Equal("password"))
			Expect(spaceGUID).To(Equal(""))
		})

		Context("when the --space-scoped flag is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("service-broker", "username", "password", "url", "--space-scoped")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create the service broker with the targeted space guid", func() {
				cmd.Execute(flagContext)
				Expect(serviceBrokerRepo.CreateCallCount()).To(Equal(1))
				name, url, username, password, spaceGUID := serviceBrokerRepo.CreateArgsForCall(0)
				Expect(name).To(Equal("service-broker"))
				Expect(url).To(Equal("url"))
				Expect(username).To(Equal("username"))
				Expect(password).To(Equal("password"))
				Expect(spaceGUID).To(Equal("my-space-guid"))
			})

			It("tells the user it is creating the service broker in the targeted org and space", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating service broker service-broker in org my-org / space my-space as my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when creating the service broker succeeds", func() {
			BeforeEach(func() {
				serviceBrokerRepo.CreateReturns(nil)
			})

			It("says OK", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
			})
		})

		Context("when creating the service broker fails", func() {
			BeforeEach(func() {
				serviceBrokerRepo.CreateReturns(errors.New("create-err"))
			})

			It("says OK", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"create-err"},
				))
			})
		})
	})
})
