package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleLogRateLimitOverride", func() {
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
		transformedManifest, executeErr = HandleLogRateLimitOverride(originalManifest, overrides)
	})

	When("log rate limit is not set on a flag override", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
						{Type: "worker", LogRateLimit: "1B"},
					},
				},
			}
		})

		It("does not change the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "web"},
						{Type: "worker", LogRateLimit: "1B"},
					},
				},
			))
		})
	})

	When("manifest web process does not specify log rate limit", func() {
		BeforeEach(func() {
			overrides.LogRateLimit = "64K"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		It("changes the log rate limit of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "web", LogRateLimit: "64K"},
					},
				},
			))
		})
	})

	When("manifest app has only non-web processes", func() {
		BeforeEach(func() {
			overrides.LogRateLimit = "32B"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			}
		})

		It("changes the log rate limit of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					LogRateLimit: "32B",
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			))
		})
	})

	When("manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.LogRateLimit = "4MB"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web"},
					},
					LogRateLimit: "1GB",
				},
			}
		})

		It("changes the log rate limit of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web", LogRateLimit: "4MB"},
					},
					LogRateLimit: "1GB",
				},
			))
		})
	})

	When("there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.LogRateLimit = "64M"

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
