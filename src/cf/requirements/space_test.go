package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSpaceReqExecute(t *testing.T) {
	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByName: space}
	ui := new(testhelpers.FakeUI)

	spaceReq := NewSpaceRequirement("foo", ui, spaceRepo)
	success := spaceReq.Execute()

	assert.True(t, success)
	assert.Equal(t, spaceRepo.SpaceName, "foo")
	assert.Equal(t, spaceReq.GetSpace(), space)
}

func TestSpaceReqExecuteWhenSpaceNotFound(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	ui := new(testhelpers.FakeUI)

	spaceReq := NewSpaceRequirement("foo", ui, spaceRepo)
	success := spaceReq.Execute()

	assert.False(t, success)
}
