package buildpack_test

import (
	"cf"
	"cf/commands/buildpack"
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListBuildpacksRequirements", func() {
			buildpackRepo := &testapi.FakeBuildpackRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callListBuildpacks(reqFactory, buildpackRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callListBuildpacks(reqFactory, buildpackRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListBuildpacks", func() {

			buildpackBuilder := func(name string, position int, enabled bool, locked bool) (buildpack cf.Buildpack) {
				buildpack.Name = name
				buildpack.Position = &position
				buildpack.Enabled = &enabled
				buildpack.Locked = &locked
				return
			}

			buildpacks := []cf.Buildpack{
				buildpackBuilder("Buildpack-1", 5, true, false),
				buildpackBuilder("Buildpack-2", 10, false, true),
				buildpackBuilder("Buildpack-3", 15, true, false),
			}

			buildpackRepo := &testapi.FakeBuildpackRepository{
				Buildpacks: buildpacks,
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

			buildpacks := []cf.Buildpack{}
			buildpackRepo := &testapi.FakeBuildpackRepository{
				Buildpacks: buildpacks,
			}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			ui := callListBuildpacks(reqFactory, buildpackRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting buildpacks"},
				{"No buildpacks found"},
			})
		})
	})
}
