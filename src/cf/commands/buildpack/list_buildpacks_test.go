package buildpack_test

import (
	"cf"
	"cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListBuildpacksRequirements(t *testing.T) {
	buildpackRepo := &testapi.FakeBuildpackRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callListBuildpacks(reqFactory, buildpackRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callListBuildpacks(reqFactory, buildpackRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListBuildpacks(t *testing.T) {
	buildpackBuilder := func(name string, position int) (buildpack cf.Buildpack) {
		buildpack.Name = name
		buildpack.Position = &position
		return
	}

	buildpacks := []cf.Buildpack{
		buildpackBuilder("Buildpack-1", 5),
		buildpackBuilder("Buildpack-2", 10),
		buildpackBuilder("Buildpack-3", 15),
	}

	buildpackRepo := &testapi.FakeBuildpackRepository{
		Buildpacks: buildpacks,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callListBuildpacks(reqFactory, buildpackRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting buildpacks"},
		{"buildpack", "position"},
		{"Buildpack-1", "5"},
		{"Buildpack-2", "10"},
		{"Buildpack-3", "15"},
	})
}

func TestListingBuildpacksWhenNoneExist(t *testing.T) {
	buildpacks := []cf.Buildpack{}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		Buildpacks: buildpacks,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callListBuildpacks(reqFactory, buildpackRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting buildpacks"},
		{"No buildpacks found"},
	})
}

func callListBuildpacks(reqFactory *testreq.FakeReqFactory, buildpackRepo *testapi.FakeBuildpackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("buildpacks", []string{})
	cmd := buildpack.NewListBuildpacks(ui, buildpackRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
