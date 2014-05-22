/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package buildpack_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/buildpack"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callUpdateBuildpack(args []string, requirementsFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	cmd := NewUpdateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

func getRepositories() (*testapi.FakeBuildpackRepository, *testapi.FakeBuildpackBitsRepository) {
	return &testapi.FakeBuildpackRepository{}, &testapi.FakeBuildpackBitsRepository{}
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateBuildpackRequirements", func() {
		repo, bitsRepo := getRepositories()

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
		callUpdateBuildpack([]string{"my-buildpack", "-p", "buildpack.zip", "extraArg"}, requirementsFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
		callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, BuildpackSuccess: true}
		callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestUpdateBuildpack", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating buildpack", "my-buildpack"},
			[]string{"OK"},
		))
	})

	It("TestUpdateBuildpackPosition", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"-i", "999", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(*repo.UpdateBuildpackArgs.Buildpack.Position).To(Equal(999))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating buildpack", "my-buildpack"},
			[]string{"OK"},
		))
	})

	It("TestUpdateBuildpackNoPosition", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(repo.UpdateBuildpackArgs.Buildpack.Position).To(BeNil())
	})

	It("TestUpdateBuildpackEnabled", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		fakeUI := callUpdateBuildpack([]string{"--enable", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
		Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(true))

		Expect(fakeUI.Outputs[0]).To(ContainSubstring("Updating buildpack"))
		Expect(fakeUI.Outputs[0]).To(ContainSubstring("my-buildpack"))
		Expect(fakeUI.Outputs[1]).To(ContainSubstring("OK"))
	})

	It("TestUpdateBuildpackDisabled", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"--disable", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
		Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(false))
	})

	It("TestUpdateBuildpackNoEnable", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).To(BeNil())
	})

	It("TestUpdateBuildpackPath", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"-p", "buildpack.zip", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(bitsRepo.UploadBuildpackPath).To(Equal("buildpack.zip"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating buildpack", "my-buildpack"},
			[]string{"OK"},
		))
	})

	It("TestUpdateBuildpackWithInvalidPath", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()
		bitsRepo.UploadBuildpackErr = true

		ui := callUpdateBuildpack([]string{"-p", "bogus/path", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating buildpack", "my-buildpack"},
			[]string{"FAILED"},
		))
	})

	It("TestUpdateBuildpackLock", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--lock", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(repo.UpdateBuildpackArgs.Buildpack.Locked).NotTo(BeNil())
		Expect(*repo.UpdateBuildpackArgs.Buildpack.Locked).To(Equal(true))

		Expect(ui.Outputs[0]).To(ContainSubstring("Updating buildpack"))
		Expect(ui.Outputs[0]).To(ContainSubstring("my-buildpack"))
		Expect(ui.Outputs[1]).To(ContainSubstring("OK"))
	})

	It("TestUpdateBuildpackUnlock", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--unlock", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(ui.Outputs[0]).To(ContainSubstring("Updating buildpack"))
		Expect(ui.Outputs[0]).To(ContainSubstring("my-buildpack"))
		Expect(ui.Outputs[1]).To(ContainSubstring("OK"))
	})

	It("TestUpdateBuildpackInvalidLockWithBits", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--lock", "-p", "buildpack.zip", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(ui.Outputs[1]).To(ContainSubstring("FAILED"))
	})

	It("TestUpdateBuildpackInvalidUnlockWithBits", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--unlock", "-p", "buildpack.zip", "my-buildpack"}, requirementsFactory, repo, bitsRepo)

		Expect(ui.Outputs[1]).To(ContainSubstring("FAILED"))
	})
})
