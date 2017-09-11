package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var (
		actor       *Actor
		cmdSettings CommandLineSettings

		currentDirectory string
	)

	BeforeEach(func() {
		actor = NewActor(nil)
		currentDirectory = getCurrentDir()
	})

	Context("when only passed command line settings", func() {
		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: currentDirectory,
				DockerImage:      "some-image",
				Name:             "some-app",
			}
		})

		It("returns a manifest made from the command line settings", func() {
			manifests, err := actor.MergeAndValidateSettingsAndManifests(cmdSettings, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(manifests).To(Equal([]manifest.Application{{
				DockerImage: "some-image",
				Name:        "some-app",
				Path:        currentDirectory,
			}}))
		})
	})

	Context("when passed command line settings and a single manifest application", func() {
		var (
			apps       []manifest.Application
			mergedApps []manifest.Application
			executeErr error
		)

		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: currentDirectory,
				Name:             "steve",
			}

			apps = []manifest.Application{
				{Name: "app-1"},
			}
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		It("merges command line settings and manifest apps", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(mergedApps).To(ConsistOf(
				manifest.Application{
					Name: "steve",
					Path: currentDirectory,
				},
			))
		})
	})

	Context("when passed command line settings and multiple manifest applications", func() {
		var (
			apps       []manifest.Application
			mergedApps []manifest.Application
			executeErr error
		)

		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: currentDirectory,
			}

			apps = []manifest.Application{
				{Name: "app-1"},
				{Name: "app-2"},
			}
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		It("merges command line settings and manifest apps", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(mergedApps).To(ConsistOf(
				manifest.Application{
					Name: "app-1",
					Path: currentDirectory,
				},
				manifest.Application{
					Name: "app-2",
					Path: currentDirectory,
				},
			))
		})

		Context("when CommandLineSettings specify an app in the manifests", func() {
			Context("when the app exists in the manifest", func() {
				BeforeEach(func() {
					cmdSettings.Name = "app-1"
				})

				It("returns just the specified app manifest", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(mergedApps).To(ConsistOf(
						manifest.Application{
							Name: "app-1",
							Path: currentDirectory,
						},
					))
				})
			})

			Context("when the app does *not* exist in the manifest", func() {
				BeforeEach(func() {
					cmdSettings.Name = "app-4"
				})

				It("returns just the specified app manifest", func() {
					Expect(executeErr).To(MatchError(AppNotFoundInManifestError{Name: "app-4"}))
				})
			})
		})
	})

	Describe("defaulting values", func() {
		var (
			apps       []manifest.Application
			mergedApps []manifest.Application
			executeErr error
		)

		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: currentDirectory,
			}

			apps = []manifest.Application{
				{Name: "app-1"},
				{Name: "app-2"},
			}
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		Context("when HealthCheckType is set to http and no endpoint is set", func() {
			BeforeEach(func() {
				apps[0].HealthCheckType = "http"
				apps[1].HealthCheckType = "http"
				apps[1].HealthCheckHTTPEndpoint = "/banana"
			})

			It("sets health-check-http-endpoint to '/'", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(mergedApps[0].HealthCheckHTTPEndpoint).To(Equal("/"))
				Expect(mergedApps[1].HealthCheckHTTPEndpoint).To(Equal("/banana"))
			})
		})
	})

	DescribeTable("validation errors",
		func(settings CommandLineSettings, apps []manifest.Application, expectedErr error) {
			_, err := actor.MergeAndValidateSettingsAndManifests(settings, apps)
			Expect(err).To(MatchError(expectedErr))
		},

		Entry("MissingNameError", CommandLineSettings{}, nil, MissingNameError{}),
		Entry("MissingNameError", CommandLineSettings{}, []manifest.Application{{}}, MissingNameError{}),
		Entry("NonexistentAppPathError", CommandLineSettings{Name: "some-name", ProvidedAppPath: "does-not-exist"}, nil, NonexistentAppPathError{Path: "does-not-exist"}),
		Entry("NonexistentAppPathError", CommandLineSettings{}, []manifest.Application{{Name: "some-name", Path: "does-not-exist"}}, NonexistentAppPathError{Path: "does-not-exist"}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{Buildpack: types.FilteredString{IsSet: true}},
			[]manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}},
			CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{Command: types.FilteredString{IsSet: true}},
			[]manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{DiskQuota: 4}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{DockerImage: "some-docker-image"}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{HealthCheckTimeout: 4}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{HealthCheckType: "http"}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{Instances: types.NullInt{IsSet: true}}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{Memory: 4}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{ProvidedAppPath: "some-path"}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError", CommandLineSettings{StackName: "some-stackname"}, []manifest.Application{{Name: "some-name-1"}, {Name: "some-name-2"}}, CommandLineOptionsWithMultipleAppsError{}),
	)
})
