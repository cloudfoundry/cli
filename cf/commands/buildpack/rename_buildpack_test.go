package buildpack_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/buildpack"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
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
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"OK"},
			))
		})

		It("fails when the buildpack does not exist", func() {
			fakeRepo.FindByNameNotFound = true

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"FAILED"},
				[]string{"Buildpack my-buildpack1 not found"},
			))
		})

		It("fails when there is an error updating the buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				Guid: "my-buildpack-guid",
			}
			fakeRepo.UpdateBuildpackReturns.Error = errors.New("SAD TROMBONE")

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"SAD TROMBONE"},
			))
		})
	})
})
