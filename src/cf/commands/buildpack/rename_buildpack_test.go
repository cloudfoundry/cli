package buildpack_test

import (
	. "cf/commands/buildpack"
	"cf/models"
	"cf/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callRenameBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-buildpack", args)
	cmd := NewRenameBuildpack(ui, fakeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("rename-buildpack command", func() {
	It("TestRenameBuildpackRequirements", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo := &testapi.FakeBuildpackRepository{}

		ui := callRenameBuildpack([]string{}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Failed"},
		})

		ui = callRenameBuildpack([]string{"my-buildpack"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Failed"},
		})

		ui = callRenameBuildpack([]string{"my-buildpack", "renamed-buildpack"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeFalse())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Ok"},
		})
	})

	It("TestRenameBuildpack", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		repo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
		}

		ui := callRenameBuildpack([]string{"my-buildpack", "new-buildpack"}, reqFactory, repo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming buildpack", "my-buildpack"},
			{"OK"},
		})
	})

	It("TestRenameBuildpackThatDoesNotExist", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		repo := &testapi.FakeBuildpackRepository{
			FindByNameNotFound:  true,
			FindByNameBuildpack: buildpack,
		}

		ui := callRenameBuildpack([]string{"my-buildpack1", "new-buildpack"}, reqFactory, repo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming buildpack", "my-buildpack"},
			{"OK"},
			{"Buildpack my-buildpack1 does not exist"},
		})
	})
	
	It("TestRenameBuildpackToOneThatExist", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

		buildpack := models.Buildpack{}
		buildpack.Name = "my-buildpack"
		buildpack.Guid = "my-buildpack-guid"
		repo := &testapi.FakeBuildpackRepository{
			FindByNameBuildpack: buildpack,
			RenameApiResponse:   errors.New("failed, target build pack exists"),
		}

		ui := callRenameBuildpack([]string{"my-buildpack1", "new-buildpack"}, reqFactory, repo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming buildpack", "my-buildpack"},
			{"failed, target build pack exists"},
		})
	})	
})
