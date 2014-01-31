package requirements

import (
	"cf"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSpaceReqExecute", func() {

			space := cf.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			spaceRepo := &testapi.FakeSpaceRepository{FindByNameSpace: space}
			ui := new(testterm.FakeUI)

			spaceReq := newSpaceRequirement("foo", ui, spaceRepo)
			success := spaceReq.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), spaceRepo.FindByNameName, "foo")
			assert.Equal(mr.T(), spaceReq.GetSpace(), space)
		})
		It("TestSpaceReqExecuteWhenSpaceNotFound", func() {

			spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
			ui := new(testterm.FakeUI)

			spaceReq := newSpaceRequirement("foo", ui, spaceRepo)
			success := spaceReq.Execute()

			assert.False(mr.T(), success)
		})
	})
}
