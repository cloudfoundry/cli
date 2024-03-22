package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleHealthCheckTimeoutOverride", func() {
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
		transformedManifest, executeErr = HandleHealthCheckTimeoutOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify health check timeout", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("health check timeout is not set on the flag overrides", func() {
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

		When("health check timeout set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.HealthCheckTimeout = 50
			})

			It("changes the health check timeout of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Processes: []manifestparser.Process{
							{Type: "web", HealthCheckTimeout: 50},
						},
					},
				))
			})
		})
	})

	When("health check timeout flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckTimeout: 10},
					},
				},
			}
		})

		It("changes the health check timeout in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					HealthCheckTimeout: 50,
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckTimeout: 10},
					},
				},
			))
		})
	})

	When("health check timeout flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckTimeout: 10},
						{Type: "web", HealthCheckTimeout: 20},
					},
					HealthCheckTimeout: 30,
				},
			}
		})

		It("changes the health check timeout of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "worker", HealthCheckTimeout: 10},
						{Type: "web", HealthCheckTimeout: 50},
					},
					HealthCheckTimeout: 30,
				},
			))
		})
	})

	When("health check timeout flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckTimeout = 50

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
