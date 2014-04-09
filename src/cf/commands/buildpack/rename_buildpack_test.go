package buildpack_test

import (
	. "cf/commands/buildpack"
	"cf/errors"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("rename-buildpack command", func() {
	var (
		cmd                 *RenameBuildpack
		fakeRepo            *testapi.FakeBuildpackRepository
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		ui = new(testterm.FakeUI)
		fakeRepo = &testapi.FakeBuildpackRepository{}
		cmd = NewRenameBuildpack(ui, fakeRepo)
	})

	runCommand := func(args ...string) {
		ctxt := testcmd.NewContext("rename-buildpack", args)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	}

	It("fails requirements when called without the current name and the new name to use", func() {
		runCommand("my-buildpack-name")
		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("renames a buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				Guid: "my-buildpack-guid",
			}

			runCommand("my-buildpack", "new-buildpack")
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Renaming buildpack", "my-buildpack"},
				{"OK"},
			})
		})

		It("fails when the buildpack does not exist", func() {
			fakeRepo.FindByNameNotFound = true

			runCommand("my-buildpack1", "new-buildpack")
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Renaming buildpack", "my-buildpack"},
				{"FAILED"},
				{"Buildpack my-buildpack1 not found"},
			})
		})

		It("fails when there is an error updating the buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				Guid: "my-buildpack-guid",
			}
			fakeRepo.UpdateBuildpackReturns.Error = errors.New("SAD TROMBONE")

			runCommand("my-buildpack1", "new-buildpack")
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Renaming buildpack", "my-buildpack"},
				{"SAD TROMBONE"},
			})
		})
	})
})
