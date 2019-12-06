package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleBuildpacksOverride", func() {
	var (
		originalManifest    manifestparser.Manifest
		transformedManifest manifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = manifestparser.Manifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleBuildpacksOverride(originalManifest, overrides)
	})

	When("buildpacks flag is set", func() {
		When("there is a single app in the manifest with a buildpacks specified", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}

				originalManifest.Applications = []manifestparser.Application{
					{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"buildpack-3"}},
					},
				}
			})

			It("will override the buildpacks in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"buildpack-1", "buildpack-2"}},
					},
				))
			})
		})

		When("there is a single app with --buildpack default override specified", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"default"}

				originalManifest.Applications = []manifestparser.Application{
					{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"buildpack-3"}},
					},
				}
			})

			It("sets the buildpacks list in the manifest to be an empty array", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{}},
					},
				))
			})
		})

		When("there is a single app with --buildpack null override specified", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"null"}

				originalManifest.Applications = []manifestparser.Application{
					{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"buildpack-3"}},
					},
				}
			})

			It("sets the buildpacks list in the manifest to be an empty array", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{}},
					},
				))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}

				originalManifest.Applications = []manifestparser.Application{
					{},
					{},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
			})
		})

		When("when docker is set in the manifest", func() {
			BeforeEach(func() {
				overrides.Buildpacks = []string{"buildpack-1", "buildpack-2"}

				originalManifest.Applications = []manifestparser.Application{
					{
						Name: "some-app",
						Docker: &manifestparser.Docker{
							Image: "nginx:latest",
						},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--buildpack, -b",
					ManifestProperty: "docker",
				}))
			})
		})
	})
})
