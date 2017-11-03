package pushaction_test

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
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
		actor = NewActor(nil, nil)
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
					Expect(executeErr).To(MatchError(actionerror.AppNotFoundInManifestError{Name: "app-4"}))
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

	const RealPath = "some-real-path"

	manifestWithMultipleApps := []manifest.Application{
		{Name: "some-name-1"},
		{Name: "some-name-2"},
	}

	DescribeTable("validation errors",
		func(settings CommandLineSettings, apps []manifest.Application, expectedErr error) {
			currentDirectory, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			if settings.ProvidedAppPath == RealPath {
				settings.ProvidedAppPath = currentDirectory
			}

			for i, app := range apps {
				if app.Path == RealPath {
					apps[i].Path = currentDirectory
				}
			}

			_, err = actor.MergeAndValidateSettingsAndManifests(settings, apps)
			Expect(err).To(MatchError(expectedErr))
		},

		Entry("MissingNameError", CommandLineSettings{}, nil, actionerror.MissingNameError{}),
		Entry("MissingNameError", CommandLineSettings{}, []manifest.Application{{}}, actionerror.MissingNameError{}),

		Entry("NonexistentAppPathError",
			CommandLineSettings{
				Name:            "some-name",
				ProvidedAppPath: "does-not-exist",
			}, nil,
			actionerror.NonexistentAppPathError{Path: "does-not-exist"}),
		Entry("NonexistentAppPathError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name: "some-name",
				Path: "does-not-exist",
			}},
			actionerror.NonexistentAppPathError{Path: "does-not-exist"}),

		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				Buildpack: types.FilteredString{IsSet: true},
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				Command: types.FilteredString{IsSet: true},
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				DiskQuota: 4,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				DockerImage: "some-docker-image",
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				HealthCheckTimeout: 4,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				HealthCheckType: "http",
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				Instances: types.NullInt{IsSet: true},
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{Memory: 4},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{ProvidedAppPath: "some-path"},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),

		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{RoutePath: "some-route-path"},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),

		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{StackName: "some-stackname"},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),

		Entry("DockerPasswordNotSetError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:           "some-name-1",
				DockerImage:    "some-image",
				DockerUsername: "some-username",
			}},
			actionerror.DockerPasswordNotSetError{}),

		// The following are premerge PropertyCombinationErrors
		Entry("PropertyCombinationError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:    "some-name-1",
				Routes:  []string{"some-route"},
				NoRoute: true,
				Path:    RealPath,
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"no-route", "routes"},
			}),

		// The following are postmerge PropertyCombinationErrors
		Entry("PropertyCombinationError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:        "some-name-1",
				DockerImage: "some-image",
				Buildpack:   types.FilteredString{IsSet: true},
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"docker", "buildpack"},
			}),
		Entry("PropertyCombinationError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:        "some-name-1",
				DockerImage: "some-image",
				Path:        "some-path",
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"docker", "path"},
			}),
		Entry("PropertyCombinationError",
			CommandLineSettings{
				NoHostname: true,
			},
			[]manifest.Application{{
				NoRoute:     true,
				Name:        "some-name-1",
				DockerImage: "some-docker-image",
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"no-hostname", "no-route"},
			}),
		Entry("PropertyCombinationError",
			CommandLineSettings{
				DefaultRouteHostname: "potato",
			},
			[]manifest.Application{{
				NoRoute:     true,
				Name:        "some-name-1",
				DockerImage: "some-docker-image",
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"hostname", "no-route"},
			}),
		Entry("PropertyCombinationError",
			CommandLineSettings{
				RoutePath: "some-path",
			},
			[]manifest.Application{{
				NoRoute:     true,
				Name:        "some-name-1",
				DockerImage: "some-docker-image",
			}},
			actionerror.PropertyCombinationError{
				AppName:    "some-name-1",
				Properties: []string{"route-path", "no-route"},
			}),
		Entry("HTTPHealthCheckInvalidError",
			CommandLineSettings{
				HealthCheckType: "port",
			},
			[]manifest.Application{{
				Name: "some-name-1",
				HealthCheckHTTPEndpoint: "/some/endpoint",
				Path: RealPath,
			}},
			actionerror.HTTPHealthCheckInvalidError{}),
		Entry("HTTPHealthCheckInvalidError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:                    "some-name-1",
				HealthCheckType:         "port",
				HealthCheckHTTPEndpoint: "/some/endpoint",
				Path: RealPath,
			}},
			actionerror.HTTPHealthCheckInvalidError{}),
		Entry("HTTPHealthCheckInvalidError",
			CommandLineSettings{
				HealthCheckType: "process",
			},
			[]manifest.Application{{
				Name: "some-name-1",
				HealthCheckHTTPEndpoint: "/some/endpoint",
				Path: RealPath,
			}},
			actionerror.HTTPHealthCheckInvalidError{}),
		Entry("HTTPHealthCheckInvalidError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:                    "some-name-1",
				HealthCheckType:         "process",
				HealthCheckHTTPEndpoint: "/some/endpoint",
				Path: RealPath,
			}},
			actionerror.HTTPHealthCheckInvalidError{}),
		Entry("HTTPHealthCheckInvalidError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name: "some-name-1",
				HealthCheckHTTPEndpoint: "/some/endpoint",
				Path: RealPath,
			}},
			actionerror.HTTPHealthCheckInvalidError{}),
	)
})
