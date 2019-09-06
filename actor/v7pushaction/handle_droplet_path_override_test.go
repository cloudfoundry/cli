package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDropletPathOverride", func() {
	var (
		transformedManifest pushmanifestparser.Manifest
		executeErr          error

		parsedManifest pushmanifestparser.Manifest
		flagOverrides  FlagOverrides
	)

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleDropletPathOverride(
			parsedManifest,
			flagOverrides,
		)
	})

	When("the droplet path flag override is set", func() {
		BeforeEach(func() {
			flagOverrides = FlagOverrides{DropletPath: "some-droplet-path.tgz"}
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{},
						{},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
			})
		})

		When("there is a single app in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{},
					},
				}
			})

			It("returns the unchanged manifest", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{},
				))
			})
		})

		When("when docker is set in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{
							Name: "some-app",
							Docker: &pushmanifestparser.Docker{
								Image: "nginx:latest",
							},
						},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--droplet",
					ManifestProperty: "docker",
				}))
			})
		})

		When("when buildpacks is set in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{
							Name:                    "some-app",
							RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"ruby_buildpack"}},
						},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--droplet",
					ManifestProperty: "buildpacks",
				}))
			})
		})

		When("when path is set in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{
							Name: "some-app",
							Path: "~",
						},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--droplet",
					ManifestProperty: "path",
				}))
			})
		})
	})

	When("the strategy flag override is not set", func() {
		BeforeEach(func() {
			flagOverrides = FlagOverrides{}
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{},
						{},
					},
				}
			})

			It("does not return an error", func() {
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})
	})
})
