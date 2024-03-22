package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleRandomRouteOverride", func() {
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
		transformedManifest, executeErr = HandleRandomRouteOverride(originalManifest, overrides)
	})

	When("manifest app does not specify random-route", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{},
			}
		})

		When("random-route is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{},
				))
			})
		})

		When("random-route is set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.RandomRoute = true
			})

			It("changes the random-route field of the only app in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						RandomRoute: true,
					},
				))
			})
		})
	})

	When("random-route flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.RandomRoute = true

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
