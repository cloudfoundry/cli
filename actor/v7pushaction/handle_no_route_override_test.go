package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleNoRouteOverride", func() {
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
		transformedManifest, executeErr = HandleNoRouteOverride(originalManifest, overrides)
	})

	When("manifest app does not specify no-route", func() {
		BeforeEach(func() {
			originalManifest.Applications = []pushmanifestparser.Application{
				{},
			}
		})

		When("no-route is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{},
				))
			})
		})

		When("no-route is set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.NoRoute = true
			})

			It("changes the no-route of the only app in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						NoRoute: true,
					},
				))
			})
		})
	})

	When("no-route flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.NoRoute = true

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
