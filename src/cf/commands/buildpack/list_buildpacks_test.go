package buildpack_test

import (
	"cf"
	"cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
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
	buildpackBuilder := func(name string, position int, enabled bool) (buildpack cf.Buildpack) {
		buildpack.Name = name
		buildpack.Position = &position
		buildpack.Enabled = &enabled
		return
	}

	buildpacks := []cf.Buildpack{
		buildpackBuilder("Buildpack-1", 5, true),
		buildpackBuilder("Buildpack-2", 10, false),
		buildpackBuilder("Buildpack-3", 15, true),
	}

	buildpackRepo := &testapi.FakeBuildpackRepository{
		Buildpacks: buildpacks,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callListBuildpacks(reqFactory, buildpackRepo)

	assert.Contains(t, ui.Outputs[0], "Getting buildpacks")

	assert.Contains(t, ui.Outputs[1], "buildpack")
	assert.Contains(t, ui.Outputs[1], "position")

	assert.Contains(t, ui.Outputs[2], "Buildpack-1")
	assert.Contains(t, ui.Outputs[2], "5")
	assert.Contains(t, ui.Outputs[2], "true")

	assert.Contains(t, ui.Outputs[3], "Buildpack-2")
	assert.Contains(t, ui.Outputs[3], "10")
	assert.Contains(t, ui.Outputs[3], "false")

	assert.Contains(t, ui.Outputs[4], "Buildpack-3")
	assert.Contains(t, ui.Outputs[4], "15")
	assert.Contains(t, ui.Outputs[4], "true")
}

func TestListingBuildpacksWhenNoneExist(t *testing.T) {
	buildpacks := []cf.Buildpack{}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		Buildpacks: buildpacks,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callListBuildpacks(reqFactory, buildpackRepo)

	assert.Contains(t, ui.Outputs[0], "Getting buildpacks")
	assert.Contains(t, ui.Outputs[1], "No buildpacks found")
}

func callListBuildpacks(reqFactory *testreq.FakeReqFactory, buildpackRepo *testapi.FakeBuildpackRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("buildpacks", []string{})
	cmd := buildpack.NewListBuildpacks(fakeUI, buildpackRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
