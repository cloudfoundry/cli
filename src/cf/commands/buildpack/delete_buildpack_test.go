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

var _ = Describe("delete-buildpack command", func() {
	var (
		ui                  *testterm.FakeUI
		buildpackRepo       *testapi.FakeBuildpackRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		buildpackRepo = &testapi.FakeBuildpackRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	var runCommand = func(args ...string) {
		cmd := NewDeleteBuildpack(ui, buildpackRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-buildpack", args), requirementsFactory)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("fails requirements", func() {
			runCommand("-f", "my-buildpack")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the buildpack exists", func() {
			BeforeEach(func() {
				buildpackRepo.FindByNameBuildpack = models.Buildpack{
					Name: "my-buildpack",
					Guid: "my-buildpack-guid",
				}
			})

			It("deletes the buildpack", func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}

				runCommand("my-buildpack")

				Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

				testassert.SliceContains(ui.Prompts, testassert.Lines{
					{"delete the buildpack my-buildpack"},
				})
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting buildpack", "my-buildpack"},
					{"OK"},
				})
			})

			Context("when the force flag is provided", func() {
				It("does not prompt the user to delete the buildback", func() {
					runCommand("-f", "my-buildpack")

					Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

					Expect(len(ui.Prompts)).To(Equal(0))
					testassert.SliceContains(ui.Outputs, testassert.Lines{
						{"Deleting buildpack", "my-buildpack"},
						{"OK"},
					})
				})
			})
		})

		Context("when the buildpack provided is not found", func() {
			BeforeEach(func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}
				buildpackRepo.FindByNameNotFound = true
			})

			It("warns the user", func() {
				runCommand("my-buildpack")

				Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
				Expect(buildpackRepo.FindByNameNotFound).To(BeTrue())

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "my-buildpack"},
					{"OK"},
				})

				testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
					{"my-buildpack", "does not exist"},
				})
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}

				buildpackRepo.FindByNameBuildpack = models.Buildpack{
					Name: "my-buildpack",
					Guid: "my-buildpack-guid",
				}
				buildpackRepo.DeleteApiResponse = errors.New("failed badly")
			})

			It("fails with the error", func() {
				runCommand("my-buildpack")

				Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting buildpack", "my-buildpack"},
					{"FAILED"},
					{"my-buildpack"},
					{"failed badly"},
				})
			})
		})
	})
})
