package buildpack_test

import (
	"cf/commands/buildpack"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callListBuildpacks(reqFactory *testreq.FakeReqFactory, buildpackRepo *testapi.FakeBuildpackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("buildpacks", []string{})
	cmd := buildpack.NewListBuildpacks(ui, buildpackRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("ListBuildpacks", func() {
	It("has the right requirements", func() {
		buildpackRepo := &testapi.FakeBuildpackRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		callListBuildpacks(reqFactory, buildpackRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		callListBuildpacks(reqFactory, buildpackRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})

	It("lists buildpacks", func() {
		p1 := 5
		p2 := 10
		p3 := 15
		t := true
		f := false

		buildpackRepo := &testapi.FakeBuildpackRepository{
			Buildpacks: []models.Buildpack{
				models.Buildpack{Name: "Buildpack-1", Position: &p1, Enabled: &t, Locked: &f},
				models.Buildpack{Name: "Buildpack-2", Position: &p2, Enabled: &f, Locked: &t},
				models.Buildpack{Name: "Buildpack-3", Position: &p3, Enabled: &t, Locked: &f},
			},
		}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		ui := callListBuildpacks(reqFactory, buildpackRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting buildpacks"},
			{"buildpack", "position", "enabled"},
			{"Buildpack-1", "5", "true", "false"},
			{"Buildpack-2", "10", "false", "true"},
			{"Buildpack-3", "15", "true", "false"},
		})
	})

	It("TestListingBuildpacksWhenNoneExist", func() {
		buildpackRepo := &testapi.FakeBuildpackRepository{
			Buildpacks: []models.Buildpack{},
		}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		ui := callListBuildpacks(reqFactory, buildpackRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting buildpacks"},
			{"No buildpacks found"},
		})
	})
})
