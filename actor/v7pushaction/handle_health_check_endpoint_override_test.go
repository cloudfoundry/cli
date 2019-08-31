package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleHealthCheckEndpointOverride", func() {
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
		transformedManifest, executeErr = HandleHealthCheckEndpointOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify health check type", func() {
		BeforeEach(func() {
			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("health check type is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Processes: []pushmanifestparser.Process{
							{Type: "web"},
						},
					},
				))
			})
		})

		When("health check type set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.HealthCheckEndpoint = "/health"
			})

			It("changes the health check type of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Processes: []pushmanifestparser.Process{
							{Type: "web", HealthCheckEndpoint: "/health"},
						},
					},
				))
			})
		})
	})

	When("health check type flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
					},
				},
			}
		})

		It("changes the health check type in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					HealthCheckEndpoint: "/health",
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
					},
				},
			))
		})
	})

	When("health check type flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
						{Type: "web", HealthCheckEndpoint: "/health3"},
					},
					HealthCheckEndpoint: "/health2",
				},
			}
		})

		It("changes the health check type of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					HealthCheckEndpoint: "/health2",
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
						{Type: "web", HealthCheckEndpoint: "/health"},
					},
				},
			))
		})
	})

	When("health check type flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

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
