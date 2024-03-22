package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDockerUsernameOverride", func() {
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
		transformedManifest, executeErr = HandleDockerUsernameOverride(originalManifest, overrides)
	})

	When("docker username flag is set", func() {
		When("there is a single app in the manifest without any docker info specified", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []manifestparser.Application{
					{},
				}
			})

			It("will populate the docker username in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Docker: &manifestparser.Docker{
							Username: "some-docker-username",
						},
					},
				))
			})
		})

		When("there is a single app in the manifest with a docker username specified", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []manifestparser.Application{
					{
						Docker: &manifestparser.Docker{Username: "old-docker-username"},
					},
				}
			})

			It("will override the docker username in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Docker: &manifestparser.Docker{
							Username: "some-docker-username",
						},
					},
				))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.DockerUsername = "some-docker-username"

				originalManifest.Applications = []manifestparser.Application{
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
