/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package space_test

import (
	. "cf/commands/space"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("delete-space command", func() {
	var (
		ui                  *testterm.FakeUI
		space               models.Space
		config              configuration.ReadWriter
		spaceRepo           *testapi.FakeSpaceRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	var deleteSpace = func(args ...string) {
		ctxt := testcmd.NewContext("delete-space", args)
		cmd := NewDeleteSpace(ui, config, spaceRepo)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)
		return
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		config = testconfig.NewRepositoryWithDefaults()

		space = models.Space{}
		space.Name = "space-to-delete"
		space.Guid = "space-to-delete-guid"

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
			deleteSpace("my-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when not targeting a space", func() {
			requirementsFactory.TargetedOrgSuccess = false
			deleteSpace("my-space")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("deletes a space, given its name", func() {
		ui.Inputs = []string{"yes"}
		deleteSpace("space-to-delete")

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "space-to-delete"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting space", "space-to-delete", "my-org", "my-user"},
			{"OK"},
		})
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("does not prompt when the -f flag is given", func() {
		deleteSpace("-f", "space-to-delete")

		Expect(ui.Prompts).To(BeEmpty())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting", "space-to-delete"},
			{"OK"},
		})
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
	})

	It("clears the space from the config, when deleting the space currently targeted", func() {
		config.SetSpaceFields(space.SpaceFields)
		deleteSpace("-f", "space-to-delete")

		Expect(config.HasSpace()).To(Equal(false))
	})
})
