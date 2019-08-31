package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleStackOverride", func() {
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
		transformedManifest, executeErr = HandleStackOverride(originalManifest, overrides)
	})

	When("stack flag is set", func() {
		When("there is a single app in the manifest with a stack specified", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

				originalManifest.Applications = []pushmanifestparser.Application{
					{
						Stack: "cflinuxfs3",
					},
				}
			})

			It("will override the stack in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Stack: "cflinuxfs2",
					},
				))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

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
})
