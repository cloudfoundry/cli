package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpaceReqExecute(t *testing.T) {
	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameSpace: space}
	ui := new(testterm.FakeUI)

	spaceReq := NewSpaceRequirement("foo", ui, spaceRepo)
	success := spaceReq.Execute()

	assert.True(t, success)
	assert.Equal(t, spaceRepo.FindByNameName, "foo")
	assert.Equal(t, spaceReq.GetSpace(), space)
}

func TestSpaceReqExecuteWhenSpaceNotFound(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	spaceReq := NewSpaceRequirement("foo", ui, spaceRepo)
	success := spaceReq.Execute()

	assert.False(t, success)
}
