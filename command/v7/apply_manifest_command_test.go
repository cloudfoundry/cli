package v7_test

import (
	"errors"
	"regexp"

	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("apply-manifest Command", func() {
	var (
		cmd               ApplyManifestCommand
		testUI            *ui.UI
		fakeConfig        *commandfakes.FakeConfig
		fakeSharedActor   *commandfakes.FakeSharedActor
		fakeActor         *v7fakes.FakeActor
		fakeParser        *v7fakes.FakeManifestParser
		fakeLocator       *v7fakes.FakeManifestLocator
		fakeDiffDisplayer *v7fakes.FakeDiffDisplayer
		binaryName        string
		executeErr        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeParser = new(v7fakes.FakeManifestParser)
		fakeLocator = new(v7fakes.FakeManifestLocator)
		fakeDiffDisplayer = new(v7fakes.FakeDiffDisplayer)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = ApplyManifestCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			ManifestParser:  fakeParser,
			ManifestLocator: fakeLocator,
			DiffDisplayer:   fakeDiffDisplayer,
			CWD:             "fake-directory",
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the user is logged in", func() {
		var (
			providedPath string
		)

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		When("the manifest location is specified with `-f`", func() {
			BeforeEach(func() {
				providedPath = "some-manifest-path"
				cmd.PathToManifest = flag.ManifestPathWithExistenceCheck(providedPath)
			})

			It("tries locate the manifest file at the given path", func() {
				Expect(fakeLocator.PathCallCount()).To(Equal(1))
				Expect(fakeLocator.PathArgsForCall(0)).To(Equal(providedPath))
			})
		})

		When("the manifest location is not specified with `-f`", func() {
			When("looking for the manifest file errors", func() {
				BeforeEach(func() {
					fakeLocator.PathReturns("", false, errors.New("some-error"))
				})

				It("returns the error", func() {
					Expect(fakeLocator.PathCallCount()).To(Equal(1))
					Expect(fakeLocator.PathArgsForCall(0)).To(Equal(cmd.CWD))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})

			When("the manifest file does not exist in the current directory", func() {
				BeforeEach(func() {
					fakeLocator.PathReturns("", false, nil)
				})

				It("returns a descriptive error", func() {
					Expect(executeErr).To(MatchError(translatableerror.ManifestFileNotFoundInDirectoryError{
						PathToManifest: cmd.CWD,
					}))
				})
			})

			When("the manifest file exists in the current directory", func() {
				var resolvedPath = "/fake/manifest.yml"

				BeforeEach(func() {
					cmd.PathsToVarsFiles = []flag.PathWithExistenceCheck{"vars.yml"}
					cmd.Vars = []template.VarKV{{Name: "o", Value: "nice"}}
					fakeLocator.PathReturns(resolvedPath, true, nil)
				})

				When("the manifest is successfully parsed", func() {
					var expectedDiff resources.ManifestDiff

					BeforeEach(func() {
						expectedDiff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{Op: resources.AddOperation, Path: "/path/to/field", Value: "hello"},
							},
						}

						fakeActor.SetSpaceManifestReturns(
							v7action.Warnings{"some-manifest-warning"},
							nil,
						)
						fakeParser.InterpolateManifestReturns([]byte("interpolated!"), nil)
						fakeParser.MarshalManifestReturns([]byte("manifesto"), nil)
						fakeActor.DiffSpaceManifestReturns(expectedDiff, nil, nil)
					})

					It("displays the success text", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Applying manifest %s in org some-org / space some-space as steve...", regexp.QuoteMeta(resolvedPath)))
						Expect(testUI.Err).To(Say("some-manifest-warning"))
						Expect(testUI.Out).To(Say("OK"))

						Expect(fakeParser.InterpolateManifestCallCount()).To(Equal(1))
						path, varsFiles, vars := fakeParser.InterpolateManifestArgsForCall(0)
						Expect(path).To(Equal(resolvedPath))
						Expect(varsFiles).To(Equal([]string{"vars.yml"}))
						Expect(vars).To(Equal([]template.VarKV{{Name: "o", Value: "nice"}}))

						Expect(fakeParser.ParseManifestCallCount()).To(Equal(1))
						path, rawManifest := fakeParser.ParseManifestArgsForCall(0)
						Expect(path).To(Equal(resolvedPath))
						Expect(rawManifest).To(Equal([]byte("interpolated!")))

						Expect(fakeActor.DiffSpaceManifestCallCount()).To(Equal(1))
						spaceGUID, manifestBytes := fakeActor.DiffSpaceManifestArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(manifestBytes).To(Equal([]byte("manifesto")))

						Expect(fakeDiffDisplayer.DisplayDiffCallCount()).To(Equal(1))
						manifestBytes, diff := fakeDiffDisplayer.DisplayDiffArgsForCall(0)
						Expect(manifestBytes).To(Equal([]byte("manifesto")))
						Expect(diff).To(Equal(expectedDiff))

						Expect(fakeActor.SetSpaceManifestCallCount()).To(Equal(1))
						spaceGUIDArg, actualBytes := fakeActor.SetSpaceManifestArgsForCall(0)
						Expect(actualBytes).To(Equal([]byte("manifesto")))
						Expect(spaceGUIDArg).To(Equal("some-space-guid"))
					})
				})

				When("the manifest is unparseable", func() {
					BeforeEach(func() {
						fakeParser.ParseManifestReturns(manifestparser.Manifest{}, &yaml.TypeError{
							Errors: []string{"oooooh nooooos"},
						})
					})

					It("returns back the parse error", func() {
						Expect(executeErr).To(MatchError(errors.New("Unable to apply manifest because its format is invalid.")))

						Expect(fakeActor.SetSpaceManifestCallCount()).To(Equal(0))
					})
				})

				When("retrieving the manifest diff 500s", func() {
					BeforeEach(func() {
						fakeActor.DiffSpaceManifestReturns(resources.ManifestDiff{}, v7action.Warnings{}, ccerror.V3UnexpectedResponseError{})
					})

					It("reports the 500, does not display the diff, but still applies the manifest", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Err).To(Say("Unable to generate diff. Continuing to apply manifest..."))
						Expect(fakeDiffDisplayer.DisplayDiffCallCount()).To(Equal(0))
						Expect(fakeActor.SetSpaceManifestCallCount()).To(Equal(1))
					})
				})

				When("displaying the manifest diff fails", func() {
					BeforeEach(func() {
						fakeDiffDisplayer.DisplayDiffReturns(errors.New("diff failed"))
					})

					It("returns the diff error", func() {
						Expect(executeErr).To(MatchError("diff failed"))
						Expect(fakeActor.SetSpaceManifestCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
