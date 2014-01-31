package buildpack_test

import (
	"cf"
	. "cf/commands/buildpack"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteBuildpackGetRequirements", func() {

			ui := &testterm.FakeUI{Inputs: []string{"y"}}
			buildpackRepo := &testapi.FakeBuildpackRepository{}
			cmd := NewDeleteBuildpack(ui, buildpackRepo)

			ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestDeleteBuildpackSuccess", func() {

			ui := &testterm.FakeUI{Inputs: []string{"y"}}
			buildpack := cf.Buildpack{}
			buildpack.Name = "my-buildpack"
			buildpack.Guid = "my-buildpack-guid"
			buildpackRepo := &testapi.FakeBuildpackRepository{
				FindByNameBuildpack: buildpack,
			}
			cmd := NewDeleteBuildpack(ui, buildpackRepo)

			ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"delete", "my-buildpack"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting buildpack", "my-buildpack"},
				{"OK"},
			})
		})
		It("TestDeleteBuildpackNoConfirmation", func() {

			ui := &testterm.FakeUI{Inputs: []string{"no"}}
			buildpack := cf.Buildpack{}
			buildpack.Name = "my-buildpack"
			buildpack.Guid = "my-buildpack-guid"
			buildpackRepo := &testapi.FakeBuildpackRepository{
				FindByNameBuildpack: buildpack,
			}
			cmd := NewDeleteBuildpack(ui, buildpackRepo)

			ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), buildpackRepo.DeleteBuildpackGuid, "")

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"delete", "my-buildpack"},
			})
		})
		It("TestDeleteBuildpackThatDoesNotExist", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			buildpack := cf.Buildpack{}
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

			assert.Equal(mr.T(), buildpackRepo.FindByNameName, "my-buildpack")
			assert.True(mr.T(), buildpackRepo.FindByNameNotFound)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "my-buildpack"},
				{"OK"},
				{"my-buildpack", "does not exist"},
			})
		})
		It("TestDeleteBuildpackDeleteError", func() {

			ui := &testterm.FakeUI{Inputs: []string{"y"}}
			buildpack := cf.Buildpack{}
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

			assert.Equal(mr.T(), buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting buildpack", "my-buildpack"},
				{"FAILED"},
				{"my-buildpack"},
				{"failed badly"},
			})
		})
		It("TestDeleteBuildpackForceFlagSkipsConfirmation", func() {

			ui := &testterm.FakeUI{}
			buildpack := cf.Buildpack{}
			buildpack.Name = "my-buildpack"
			buildpack.Guid = "my-buildpack-guid"
			buildpackRepo := &testapi.FakeBuildpackRepository{
				FindByNameBuildpack: buildpack,
			}

			cmd := NewDeleteBuildpack(ui, buildpackRepo)

			ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting buildpack", "my-buildpack"},
				{"OK"},
			})
		})
	})
}
