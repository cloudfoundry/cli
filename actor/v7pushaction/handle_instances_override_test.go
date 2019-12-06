package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleInstancesOverride", func() {
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
		transformedManifest, executeErr = HandleInstancesOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify instances", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("instances are not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("instances are set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.Instances = types.NullInt{IsSet: true, Value: 4}
			})

			It("changes the instances of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Processes: []manifestparser.Process{
							{Type: "web", Instances: &overrides.Instances.Value},
						},
					},
				))
			})
		})
	})

	When("instances flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			}
		})

		It("changes the instances of the app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Instances: &overrides.Instances.Value,
					Processes: []manifestparser.Process{
						{Type: "worker"},
					},
				},
			))
		})
	})

	When("instances flag is set, and manifest app has web and non-web processes", func() {
		var instances = 5

		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

			originalManifest.Applications = []manifestparser.Application{
				{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web"},
					},
					Instances: &instances,
				},
			}
		})

		It("changes the instances of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Processes: []manifestparser.Process{
						{Type: "worker"},
						{Type: "web", Instances: &overrides.Instances.Value},
					},
					Instances: &instances,
				},
			))
		})
	})

	When("instances flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Instances = types.NullInt{IsSet: true, Value: 4}

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
