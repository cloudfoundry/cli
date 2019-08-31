package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleStartCommandOverride", func() {
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
		transformedManifest, executeErr = HandleStartCommandOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify start command", func() {
		BeforeEach(func() {
			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("start command is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Processes: []pushmanifestparser.Process{
							{Type: "web"},
						},
					},
				))
			})
		})

		When("start command set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}
			})

			It("changes the start command of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Processes: []pushmanifestparser.Process{
							{Type: "web", RemainingManifestFields: map[string]interface{}{"command": "./start.sh"}},
						},
					},
				))
			})
		})
	})

	When("start command flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker"},
					},
				},
			}
		})

		It("changes the start command in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					RemainingManifestFields: map[string]interface{}{"command": "./start.sh"},
					Processes: []pushmanifestparser.Process{
						{Type: "worker"},
					},
				},
			))
		})
	})

	When("start command flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker"},
						{Type: "web"},
					},
				},
			}
		})

		It("changes the start command of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					Processes: []pushmanifestparser.Process{
						{Type: "worker"},
						{Type: "web", RemainingManifestFields: map[string]interface{}{"command": "./start.sh"}},
					},
				},
			))
		})
	})

	When("start command flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "./start.sh", IsSet: true}

			originalManifest.Applications = []pushmanifestparser.Application{
				{},
				{},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
		})
	})

	When("start command set on the flag overrides but is default", func() {
		BeforeEach(func() {
			overrides.StartCommand = types.FilteredString{Value: "", IsSet: true}
			originalManifest.Applications = []pushmanifestparser.Application{
				{},
			}
		})

		It("changes the start command of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications[0].RemainingManifestFields["command"]).To(BeNil())
		})
	})
})
