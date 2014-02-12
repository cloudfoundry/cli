package buildpack_test

import (
	. "cf/commands/buildpack"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteBuildpackGetRequirements", func() {

		ui := &testterm.FakeUI{Inputs: []string{"y"}}
		buildpackRepo := &testapi.FakeBuildpackRepository{}
		cmd := NewDeleteBuildpack(ui, buildpackRepo)

		ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestDeleteBuildpackSuccess", func() {

		ui := &testterm.FakeUI{Inputs: []string{"y"}}
		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		buildpackRepo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
		}
		cmd := NewDeleteBuildpack(ui, buildpackRepo)

		ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"delete", "my-buildpack"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting buildpack", "my-buildpack"},
			{"OK"},
		})
	})
	It("TestDeleteBuildpackNoConfirmation", func() {

		ui := &testterm.FakeUI{Inputs: []string{"no"}}
		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		buildpackRepo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
		}
		cmd := NewDeleteBuildpack(ui, buildpackRepo)

		ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal(""))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"delete", "my-buildpack"},
		})
	})
	It("TestDeleteBuildpackThatDoesNotExist", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		buildpackRepo := &testapi.FakeBuildpackRepository{
			FindByNameNotFound:  true,
			FindByNameBuildpack: buildpack,
		}

		ui := &testterm.FakeUI{}
		ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})

		cmd := NewDeleteBuildpack(ui, buildpackRepo)
		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
		Expect(buildpackRepo.FindByNameNotFound).To(BeTrue())

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting", "my-buildpack"},
			{"OK"},
			{"my-buildpack", "does not exist"},
		})
	})
	It("TestDeleteBuildpackDeleteError", func() {

		ui := &testterm.FakeUI{Inputs: []string{"y"}}
		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		buildpackRepo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
			DeleteApiResponse:   net.NewApiResponseWithMessage("failed badly"),
		}

		cmd := NewDeleteBuildpack(ui, buildpackRepo)

		ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting buildpack", "my-buildpack"},
			{"FAILED"},
			{"my-buildpack"},
			{"failed badly"},
		})
	})
	It("TestDeleteBuildpackForceFlagSkipsConfirmation", func() {

		ui := &testterm.FakeUI{}
		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		buildpackRepo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
		}

		cmd := NewDeleteBuildpack(ui, buildpackRepo)

		ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		testcmd.RunCommand(cmd, ctxt, reqFactory)

		Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting buildpack", "my-buildpack"},
			{"OK"},
		})
	})
})
