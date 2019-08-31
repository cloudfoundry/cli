package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDockerUsernameOverride", func() {
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
		transformedManifest, executeErr = HandleDockerUsernameOverride(originalManifest, overrides)
	})

	When("docker username flag is set", func() {
		When("there is a single app in the manifest without any docker info specified", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []pushmanifestparser.Application{
					{},
				}
			})

			It("will populate the docker username in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Docker: &pushmanifestparser.Docker{
							Username: "some-docker-username",
						},
					},
				))
			})
		})

		When("there is a single app in the manifest with a docker username specified", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []pushmanifestparser.Application{
					{
						Docker: &pushmanifestparser.Docker{Username: "old-docker-username"},
					},
				}
			})

			It("will override the docker username in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Docker: &pushmanifestparser.Docker{
							Username: "some-docker-username",
						},
					},
				))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

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
