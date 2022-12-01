package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleStackOverride", func() {
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
		transformedManifest, executeErr = HandleStackOverride(originalManifest, overrides)
	})

	When("stack is not set", func() {
		When("there is a single app in the manifest with the stack specified", func() {
			BeforeEach(func() {
				originalManifest.Applications = []manifestparser.Application{
					{
						Stack: "og_cflinuxfs",
					},
				}
			})

			It("will retain the original stack value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications[0].Stack).To(Equal("og_cflinuxfs"))
			})
		})
	})

	When("stack flag is set", func() {
		When("there is a single app in the manifest with a stack specified", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

				originalManifest.Applications = []manifestparser.Application{
					{
						Stack: "cflinuxfs3",
					},
				}
			})

			It("will override the stack in the manifest with the provided flag value", func() {
				Expect(executeErr).To(Not(HaveOccurred()))
				Expect(transformedManifest.Applications).To(ConsistOf(
					manifestparser.Application{
						Stack: "cflinuxfs2",
					},
				))
			})
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				overrides.Stack = "cflinuxfs2"

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
})
