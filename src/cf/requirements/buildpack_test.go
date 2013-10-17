package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestBuildpackReqExecute(t *testing.T) {
	buildpack := cf.Buildpack{Name: "my-buildpack", Guid: "my-buildpack-guid"}
	buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameBuildpack: buildpack}
	ui := new(testterm.FakeUI)

	buildpackReq := newBuildpackRequirement("foo", ui, buildpackRepo)
	success := buildpackReq.Execute()

	assert.True(t, success)
	assert.Equal(t, buildpackRepo.FindByNameName, "foo")
	assert.Equal(t, buildpackReq.GetBuildpack(), buildpack)
}

func TestBuildpackReqExecuteWhenBuildpackNotFound(t *testing.T) {
	buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	buildpackReq := newBuildpackRequirement("foo", ui, buildpackRepo)
	success := buildpackReq.Execute()

	assert.False(t, success)
}
