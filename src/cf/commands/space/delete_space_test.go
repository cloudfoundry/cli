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

func defaultDeleteSpaceSpace() models.Space {
	space := models.Space{}
	space.Name = "space-to-delete"
	space.Guid = "space-to-delete-guid"
	return space
}

func defaultDeleteSpaceReqFactory() *testreq.FakeReqFactory {
	return &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: defaultDeleteSpaceSpace()}
}

func deleteSpace(inputs []string, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, spaceRepo *testapi.FakeSpaceRepository) {
	spaceRepo = &testapi.FakeSpaceRepository{}

	ui = &testterm.FakeUI{
		Inputs: inputs,
	}
	ctxt := testcmd.NewContext("delete-space", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewDeleteSpace(ui, configRepo, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestDeleteSpaceRequirements", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		deleteSpace([]string{"y"}, []string{"my-space"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		deleteSpace([]string{"y"}, []string{"my-space"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		deleteSpace([]string{"y"}, []string{"my-space"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.SpaceName).To(Equal("my-space"))
	})

	It("TestDeleteSpaceConfirmingWithY", func() {

		ui, spaceRepo := deleteSpace([]string{"y"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "space-to-delete"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting space", "space-to-delete", "my-org", "my-user"},
			{"OK"},
		})
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
	})

	It("TestDeleteSpaceConfirmingWithYes", func() {

		ui, spaceRepo := deleteSpace([]string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "space-to-delete"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting space", "space-to-delete", "my-org", "my-user"},
			{"OK"},
		})
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
	})

	It("TestDeleteSpaceWithForceOption", func() {

		ui, spaceRepo := deleteSpace([]string{}, []string{"-f", "space-to-delete"}, defaultDeleteSpaceReqFactory())

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting", "space-to-delete"},
			{"OK"},
		})
		Expect(spaceRepo.DeletedSpaceGuid).To(Equal("space-to-delete-guid"))
	})

	It("TestDeleteSpaceWhenSpaceIsTargeted", func() {

		reqFactory := defaultDeleteSpaceReqFactory()
		spaceRepo := &testapi.FakeSpaceRepository{}

		config := testconfig.NewRepository()
		config.SetSpaceFields(defaultDeleteSpaceSpace().SpaceFields)

		ui := &testterm.FakeUI{}
		ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

		cmd := NewDeleteSpace(ui, config, spaceRepo)
		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(config.HasSpace()).To(Equal(false))
	})

	It("TestDeleteSpaceWhenSpaceNotTargeted", func() {
		reqFactory := defaultDeleteSpaceReqFactory()
		spaceRepo := &testapi.FakeSpaceRepository{}

		otherSpace := models.SpaceFields{}
		otherSpace.Name = "do-not-delete"
		otherSpace.Guid = "do-not-delete-guid"

		config := testconfig.NewRepository()
		config.SetSpaceFields(otherSpace)

		ui := &testterm.FakeUI{}
		ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

		cmd := NewDeleteSpace(ui, config, spaceRepo)
		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(config.HasSpace()).To(Equal(true))
	})

	It("TestDeleteSpaceCommandWith", func() {

		ui, _ := deleteSpace([]string{"Yes"}, []string{}, defaultDeleteSpaceReqFactory())
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui, _ = deleteSpace([]string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
})
