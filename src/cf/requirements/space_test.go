package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpaceReqExecute(t *testing.T) {
	space := cf.Space{}
	space.Name = "awesome-sauce-space"
	space.Guid = "my-space-guid"
	spaceRepo := &testapi.FakeSpaceRepository{Spaces: []cf.Space{space}}
	ui := new(testterm.FakeUI)

	spaceReq := newSpaceRequirement("awesome-sauce-space", ui, spaceRepo)
	success := spaceReq.Execute()

	assert.True(t, success)
	assert.Equal(t, spaceRepo.FindByNameName, "awesome-sauce-space")
	assert.Equal(t, spaceReq.GetSpace(), space)
}

func TestSpaceReqExecuteWhenSpaceNotFound(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
	ui := new(testterm.FakeUI)

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		newSpaceRequirement("foo", ui, spaceRepo).Execute()
	})
}
