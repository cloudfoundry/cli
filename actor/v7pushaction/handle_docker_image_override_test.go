package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDockerImageOverride", func() {
	var (
		originalManifest    pushmanifestparser.Manifest
		transformedManifest pushmanifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = pushmanifestparser.Manifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleDockerImageOverride(originalManifest, overrides)
	})

	When("docker image flag is set", func() {
		When("there is a single app in the manifest without any docker info specified", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []pushmanifestparser.Application{
					{},
				}
			})

			It("will populate the docker image in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Docker: &pushmanifestparser.Docker{
							Image: "some-docker-image",
						},
					},
				))
			})
		})

		When("there is a single app in the manifest with a docker image specified", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []pushmanifestparser.Application{
					{
						Docker: &pushmanifestparser.Docker{Image: "old-docker-image"},
					},
				}
			})

			It("will override the docker image in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Docker: &pushmanifestparser.Docker{
							Image: "some-docker-image",
						},
					},
				))
			})
		})

		When("when buildpacks are set in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []pushmanifestparser.Application{
					{
						Name:                    "some-app",
						RemainingManifestFields: map[string]interface{}{"buildpacks": []string{"buildpack-1"}},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--docker-image, -o",
					ManifestProperty: "buildpacks",
				}))
			})
		})

		When("a path is set in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []pushmanifestparser.Application{
					{
						Name: "some-app",
						Path: "/some/path",
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
					Arg:              "--docker-image, -o",
					ManifestProperty: "path",
				}))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerImage = "some-docker-image"

				originalManifest.Applications = []pushmanifestparser.Application{
					{},
					{},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
			})
		})
	})
})
