package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleAppNameOverride", func() {
	var (
		originalManifest    manifestparser.Manifest
		transformedManifest manifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest.Applications = []manifestparser.Application{
			{
				Name: "app-1",
			},
			{
				Name: "app-2",
			},
		}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleAppNameOverride(originalManifest, overrides)
	})

	When("app name is not given as arg", func() {
		It("does not change the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Name: "app-1",
				},
				manifestparser.Application{
					Name: "app-2",
				},
			))
		})
	})

	When("a valid app name is set as a manifest override", func() {
		BeforeEach(func() {
			overrides.AppName = "app-2"
		})

		It("removes non-specified apps from manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			numApps := len(transformedManifest.Applications)
			Expect(numApps).To(Equal(1))
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Name: "app-2",
				},
			))
		})
	})

	When("there is only one app, with no name, and a name is given as a manifest override", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{},
			}

			overrides.AppName = "app-2"
		})

		It("gives the app a name in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			numApps := len(transformedManifest.Applications)
			Expect(numApps).To(Equal(1))
			Expect(transformedManifest.Applications).To(ConsistOf(
				manifestparser.Application{
					Name: "app-2",
				},
			))
		})
	})

	When("there are multiple apps and one does not have a name", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{
					Name: "app-1",
				},
				{
					Name: "",
				},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("Found an application with no name specified."))
		})
	})

	When("an invalid app name is set as a manifest override", func() {
		BeforeEach(func() {
			overrides.AppName = "unknown-app"
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(manifestparser.AppNotInManifestError{Name: "unknown-app"}))
		})
	})

	When("the manifest has no name and no appName arg is given", func() {
		BeforeEach(func() {
			originalManifest.Applications = []manifestparser.Application{
				{},
			}
		})

		It("returns a AppNameOrManifestRequiredError", func() {
			Expect(executeErr).To(MatchError(translatableerror.AppNameOrManifestRequiredError{}))
		})
	})
})
