package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleDiskOverride", func() {
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
		transformedManifest, executeErr = HandleDiskOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify disk", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("disk is not set on the flag overrides", func() {
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

		When("disk is set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Disk = "5MB"
			})

			It("changes the disk of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Processes: []manifestparser.Process{
							{
								Type:      "web",
								DiskQuota: "5MB",
							},
						},
					},
				))
			})
		})
	})

	When("disk flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			}
		})

		It("changes the disk of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					DiskQuota: "5MB",
					Processes: []manifestparser.Process{
						{
							Type: "worker",
						},
					},
				},
			))
		})
	})

	When("disk flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker", DiskQuota: "2MB"},
						{Type: "web", DiskQuota: "3MB"},
					},
					DiskQuota: "1MB",
				},
			}
		})

		It("changes the disk of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "worker", DiskQuota: "2MB"},
						{Type: "web", DiskQuota: "5MB"},
					},
					DiskQuota: "1MB",
				},
			))
		})
	})

	When("disk flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Disk = "5MB"

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
