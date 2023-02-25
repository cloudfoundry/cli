package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleHealthCheckEndpointOverride", func() {
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
		transformedManifest, executeErr = HandleHealthCheckEndpointOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify health check type", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("health check type is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Processes: []manifestparser.Process{
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
					manifestparser.Application{
						Processes: []manifestparser.Process{
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

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
					},
				},
			}
		})

		It("changes the health check type in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					HealthCheckEndpoint: "/health",
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckEndpoint: "/health2"},
					},
				},
			))
		})
	})

	When("health check type flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
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
				manifestparser.Application{
					HealthCheckEndpoint: "/health2",
					Processes: []manifestparser.Process{
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

			originalManifest.Applications = []manifestparser.Application{
				{},
				{},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
		})
	})

	When("manifest health check type is not set to http (app level), and health check endpoint flag is set", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []manifestparser.Application{
				{
					HealthCheckType: "port",
				},
			}
		})

		It("returns an error ", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
				Arg:              "--endpoint",
				ManifestProperty: "health-check-type",
				ManifestValue:    "port",
			}))
		})
	})

	When("manifest health check type is not set to http (process level), and health check endpoint flag is set", func() {
		BeforeEach(func() {
			overrides.HealthCheckEndpoint = "/health"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web", HealthCheckType: "port"},
					},
				},
			}
		})

		It("returns an error ", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
				Arg:              "--endpoint",
				ManifestProperty: "health-check-type",
				ManifestValue:    "port",
			}))
		})
	})

})
