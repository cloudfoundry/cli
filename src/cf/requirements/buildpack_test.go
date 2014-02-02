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
		It("TestBuildpackReqExecute", func() {

			buildpack := cf.Buildpack{}
			buildpack.Name = "my-buildpack"
			buildpack.Guid = "my-buildpack-guid"
			buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameBuildpack: buildpack}
			ui := new(testterm.FakeUI)

			buildpackReq := newBuildpackRequirement("foo", ui, buildpackRepo)
			success := buildpackReq.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), buildpackRepo.FindByNameName, "foo")
			assert.Equal(mr.T(), buildpackReq.GetBuildpack(), buildpack)
		})
		It("TestBuildpackReqExecuteWhenBuildpackNotFound", func() {

			buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameNotFound: true}
			ui := new(testterm.FakeUI)

			buildpackReq := newBuildpackRequirement("foo", ui, buildpackRepo)
			success := buildpackReq.Execute()

			assert.False(mr.T(), success)
		})
	})
}
