package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleTaskOverride", func() {
	var (
		originalManifest    manifestparser.Manifest
		transformedManifest manifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = manifestparser.Manifest{
			Applications: []manifestparser.Application{{}},
		}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleTaskOverride(originalManifest, overrides)
	})

	When("task is set on the flag overrides", func() {
		BeforeEach(func() {
			overrides.Task = true
		})

		It("changes the no-route of the only app in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					NoRoute: true,
				},
			))
		})
	})

	When("task flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.Task = true

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
