package space_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	"code.cloudfoundry.org/cli/plugin/models"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commands/space"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("spaces command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		spaceRepo           *spacesfakes.FakeSpaceRepository

		deps commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("spaces").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
		ui = &testterm.FakeUI{}
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("spaces", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when an org is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(runCommand()).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &space.ListSpaces{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs, err := cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				err = testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	listSpacesStub := func(spaces []models.Space) func(func(models.Space) bool) error {
		return func(cb func(models.Space) bool) error {
			var keepGoing bool
			for _, s := range spaces {
				keepGoing = cb(s)
				if !keepGoing {
					return nil
				}
			}
			return nil
		}
	}

	Describe("when invoked by a plugin", func() {
		var (
			pluginModels []plugin_models.GetSpaces_Model
		)

		BeforeEach(func() {
			pluginModels = []plugin_models.GetSpaces_Model{}
			deps.PluginModels.Spaces = &pluginModels

			space := models.Space{}
			space.Name = "space1"
			space.GUID = "123"
			space2 := models.Space{}
			space2.Name = "space2"
			space2.GUID = "456"
			spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space, space2})

			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		})

		It("populates the plugin models upon execution", func() {
			testcmd.RunCLICommand("spaces", []string{}, requirementsFactory, updateCommandDependency, true, ui)
			runCommand()
			Expect(pluginModels[0].Name).To(Equal("space1"))
			Expect(pluginModels[0].Guid).To(Equal("123"))
			Expect(pluginModels[1].Name).To(Equal("space2"))
			Expect(pluginModels[1].Guid).To(Equal("456"))
		})
	})

	Context("when logged in and an org is targeted", func() {
		BeforeEach(func() {
			space := models.Space{}
			space.Name = "space1"
			space2 := models.Space{}
			space2.Name = "space2"
			space3 := models.Space{}
			space3.Name = "space3"
			spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{space, space2, space3})
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		})

		It("lists all of the spaces", func() {
			runCommand()

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting spaces in org", "my-org", "my-user"},
				[]string{"space1"},
				[]string{"space2"},
				[]string{"space3"},
			))
		})

		Context("when there are no spaces", func() {
			BeforeEach(func() {
				spaceRepo.ListSpacesStub = listSpacesStub([]models.Space{})
			})

			It("politely tells the user", func() {
				runCommand()
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting spaces in org", "my-org", "my-user"},
					[]string{"No spaces found"},
				))
			})
		})
	})
})
