package space_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/space"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-space command", func() {
	var (
		ui                  *testterm.FakeUI
		space               models.Space
		config              configuration.ReadWriter
		spaceRepo           *testapi.FakeSpaceRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	runCommand := func(args ...string) {
		ctxt := testcmd.NewContext("delete-space", args)
		cmd := NewDeleteSpace(ui, config, spaceRepo)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)
		return
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		config = testconfig.NewRepositoryWithDefaults()

		space = maker.NewSpace(maker.Overrides{
			"name": "space-to-delete",
			"guid": "space-to-delete-guid",
		})

		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:       true,
			TargetedOrgSuccess: true,
			Space:              space,
		}
	})

	Describe("requirements", func() {
		BeforeEach(func() {
			ui.Inputs = []string{"y"}
		})
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("my-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when not targeting a space", func() {
			requirementsFactory.TargetedOrgSuccess = false
			runCommand("my-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("deletes a space, given its name", func() {
		ui.Inputs = []string{"yes"}
		runCommand("space-to-delete")

		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the space space-to-delete"}))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Deleting space", "space-to-delete", "my-org", "my-user"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("does not prompt when the -f flag is given", func() {
		runCommand("-f", "space-to-delete")

		Expect(ui.Prompts).To(BeEmpty())
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Deleting", "space-to-delete"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
	})

	It("clears the space from the config, when deleting the space currently targeted", func() {
		config.SetSpaceFields(space.SpaceFields)
		runCommand("-f", "space-to-delete")

		Expect(config.HasSpace()).To(Equal(false))
	})
})
