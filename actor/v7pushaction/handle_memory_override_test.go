package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleMemoryOverride", func() {
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
		transformedManifest, executeErr = HandleMemoryOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify memory", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("memory are not set on the flag overrides", func() {
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

		When("memory are set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Memory = "64M"
			})

			It("changes the memory of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Processes: []manifestparser.Process{
							{Type: "web", Memory: "64M"},
						},
					},
				))
			})
		})
	})

	When("memory flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			}
		})

		It("changes the memory of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Memory: "64M",
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			))
		})
	})

	When("memory flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web"},
					},
					Memory: "8M",
				},
			}
		})

		It("changes the memory of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web", Memory: "64M"},
					},
					Memory: "8M",
				},
			))
		})
	})

	When("memory flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Memory = "64M"

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
