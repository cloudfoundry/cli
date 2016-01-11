package application_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	fakeappfiles "github.com/cloudfoundry/cli/cf/api/app_files/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {
	var (
		ui           *testterm.FakeUI
		configRepo   core_config.Repository
		appFilesRepo *fakeappfiles.FakeAppFilesRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement          requirements.Requirement
		targetedSpaceRequirement  requirements.Requirement
		deaApplicationRequirement *fakerequirements.FakeDEAApplicationRequirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		appFilesRepo = &fakeappfiles.FakeAppFilesRepository{}
		repoLocator := deps.RepoLocator.SetAppFileRepository(appFilesRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &application.Files{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		deaApplicationRequirement = &fakerequirements.FakeDEAApplicationRequirement{}
		factory.NewDEAApplicationRequirementReturns(deaApplicationRequirement)
		app := models.Application{}
		app.InstanceCount = 1
		app.Guid = "app-guid"
		app.Name = "app-name"
		deaApplicationRequirement.GetApplicationReturns(app)
	})

	Describe("Requirements", func() {
		Context("when not provided one or two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "the-path", "extra-arg")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns an DEAApplicationRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDEAApplicationRequirementCallCount()).To(Equal(1))
				Expect(factory.NewDEAApplicationRequirementArgsForCall(0)).To(Equal("app-name"))
				Expect(actualRequirements).To(ContainElement(deaApplicationRequirement))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "the-path")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns an DEAApplicationRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDEAApplicationRequirementCallCount()).To(Equal(1))
				Expect(factory.NewDEAApplicationRequirementArgsForCall(0)).To(Equal("app-name"))
				Expect(actualRequirements).To(ContainElement(deaApplicationRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var args []string

		JustBeforeEach(func() {
			err := flagContext.Parse(args...)
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when given a valid instance", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "0"}
			})

			It("tells the user it is getting the files", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting files for app app-name"},
				))
			})

			It("tries to list the files", func() {
				cmd.Execute(flagContext)
				Expect(appFilesRepo.ListFilesCallCount()).To(Equal(1))
				appGUID, instance, path := appFilesRepo.ListFilesArgsForCall(0)
				Expect(appGUID).To(Equal("app-guid"))
				Expect(instance).To(Equal(0))
				Expect(path).To(Equal("/"))
			})

			Context("when listing the files succeeds", func() {
				BeforeEach(func() {
					appFilesRepo.ListFilesReturns("files", nil)
				})

				It("tells the user OK", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
				})

				Context("when the files are empty", func() {
					BeforeEach(func() {
						appFilesRepo.ListFilesReturns("", nil)
					})

					It("tells the user no files were found", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings([]string{"No files found"}))
					})
				})

				Context("when the files are not empty", func() {
					BeforeEach(func() {
						appFilesRepo.ListFilesReturns("the-files", nil)
					})

					It("tells the user which files were found", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings([]string{"the-files"}))
					})
				})
			})

			Context("when listing the files fails with an error", func() {
				BeforeEach(func() {
					appFilesRepo.ListFilesReturns("", errors.New("list-files-err"))
				})

				It("fails with error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"list-files-err"},
					))
				})
			})
		})

		Context("when given a negative instance", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "-1"}
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Instance must be a positive integer"},
				))
			})
		})

		Context("when given an instance greater than the app's instance count", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "2"}
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Invalid instance: 2"},
					[]string{"Instance must be less than 1"},
				))
			})
		})

		Context("when given a path", func() {
			BeforeEach(func() {
				args = []string{"app-name", "the-path"}
			})

			It("lists the files with the given path", func() {
				cmd.Execute(flagContext)
				_, _, path := appFilesRepo.ListFilesArgsForCall(0)
				Expect(path).To(Equal("the-path"))
			})
		})
	})
})
