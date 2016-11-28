package servicebroker_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/servicebroker"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateServiceBroker", func() {
	var (
		ui                *testterm.FakeUI
		configRepo        coreconfig.Repository
		serviceBrokerRepo *apifakes.FakeServiceBrokerRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceBrokerRepo = new(apifakes.FakeServiceBrokerRepository)
		repoLocator := deps.RepoLocator.SetServiceBrokerRepository(serviceBrokerRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &servicebroker.CreateServiceBroker{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = new(requirementsfakes.FakeFactory)

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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})
		})

		Context("when the --space-scoped flag is provided", func() {
			BeforeEach(func() {
				flagContext.Parse("service-broker", "username", "password", "url", "--space-scoped")
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var runCLIErr error

		BeforeEach(func() {
			err := flagContext.Parse("service-broker", "username", "password", "url")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		It("tells the user it is creating the service broker", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service broker", "service-broker", "my-user"},
				[]string{"OK"},
			))
		})

		It("tries to create the service broker", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
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
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(serviceBrokerRepo.CreateCallCount()).To(Equal(1))
				name, url, username, password, spaceGUID := serviceBrokerRepo.CreateArgsForCall(0)
				Expect(name).To(Equal("service-broker"))
				Expect(url).To(Equal("url"))
				Expect(username).To(Equal("username"))
				Expect(password).To(Equal("password"))
				Expect(spaceGUID).To(Equal("my-space-guid"))
			})

			It("tells the user it is creating the service broker in the targeted org and space", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
			})
		})

		Context("when creating the service broker fails", func() {
			BeforeEach(func() {
				serviceBrokerRepo.CreateReturns(errors.New("create-err"))
			})

			It("returns an error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("create-err"))
			})
		})
	})
})
