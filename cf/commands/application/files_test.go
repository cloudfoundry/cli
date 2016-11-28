package application_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/appfiles/appfilesfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {
	var (
		ui           *testterm.FakeUI
		configRepo   coreconfig.Repository
		appFilesRepo *appfilesfakes.FakeAppFilesRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement          requirements.Requirement
		targetedSpaceRequirement  requirements.Requirement
		deaApplicationRequirement *requirementsfakes.FakeDEAApplicationRequirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		appFilesRepo = new(appfilesfakes.FakeAppFilesRepository)
		repoLocator := deps.RepoLocator.SetAppFileRepository(appFilesRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &application.Files{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		deaApplicationRequirement = new(requirementsfakes.FakeDEAApplicationRequirement)
		factory.NewDEAApplicationRequirementReturns(deaApplicationRequirement)
		app := models.Application{}
		app.InstanceCount = 1
		app.GUID = "app-guid"
		app.Name = "app-name"
		deaApplicationRequirement.GetApplicationReturns(app)
	})

	Describe("Requirements", func() {
		Context("when not provided one or two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "the-path", "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
		var (
			args []string
			err  error
		)

		JustBeforeEach(func() {
			err = flagContext.Parse(args...)
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
			err = cmd.Execute(flagContext)
		})

		Context("when given a valid instance", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "0"}
			})

			It("tells the user it is getting the files", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting files for app app-name"},
				))
			})

			It("tries to list the files", func() {
				Expect(err).NotTo(HaveOccurred())
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
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
				})

				Context("when the files are empty", func() {
					BeforeEach(func() {
						appFilesRepo.ListFilesReturns("", nil)
					})

					It("tells the user empty file or no files were found", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Empty file or folder"}))
					})
				})

				Context("when the files are not empty", func() {
					BeforeEach(func() {
						appFilesRepo.ListFilesReturns("the-files", nil)
					})

					It("tells the user which files were found", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"the-files"}))
					})
				})
			})

			Context("when listing the files fails with an error", func() {
				BeforeEach(func() {
					appFilesRepo.ListFilesReturns("", errors.New("list-files-err"))
				})

				It("fails with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("list-files-err"))
				})
			})
		})

		Context("when given a negative instance", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "-1"}
			})

			It("fails with error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Instance must be a positive integer"))
			})
		})

		Context("when given an instance greater than the app's instance count", func() {
			BeforeEach(func() {
				args = []string{"app-name", "-i", "2"}
			})

			It("fails with error", func() {
				Expect(err).To(HaveOccurred())
				errStr := err.Error()
				Expect(errStr).To(ContainSubstring("Invalid instance: 2"))
				Expect(errStr).To(ContainSubstring("Instance must be less than 1"))
			})
		})

		Context("when given a path", func() {
			BeforeEach(func() {
				args = []string{"app-name", "the-path"}
			})

			It("lists the files with the given path", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, path := appFilesRepo.ListFilesArgsForCall(0)
				Expect(path).To(Equal("the-path"))
			})
		})
	})
})
