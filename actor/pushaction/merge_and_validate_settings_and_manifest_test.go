package pushaction_test

import (
	"io/ioutil"
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
				{
					Name:   "app-1",
					Routes: []string{"google.com"},
				},
			}
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		It("merges command line settings and manifest apps", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(mergedApps).To(ConsistOf(
				manifest.Application{
					Name:   "steve",
					Path:   currentDirectory,
					Routes: []string{"google.com"},
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

	Describe("sanitizing values", func() {
		var (
			tempDir string

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
			}

			var err error
			tempDir, err = ioutil.TempDir("", "merge-push-settings-")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		Context("when app path is set from the command line", func() {
			BeforeEach(func() {
				cmdSettings.ProvidedAppPath = tempDir
			})

			It("sets the app path to the provided path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(mergedApps[0].Path).To(Equal(tempDir))
			})
		})

		Context("when app path is set from the manifest", func() {
			BeforeEach(func() {
				apps[0].Path = tempDir
			})

			It("sets the app path to the provided path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(mergedApps[0].Path).To(Equal(tempDir))
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
			if expectedErr == nil {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(expectedErr))
			}
		},

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
				DefaultRouteDomain: "some-domain",
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
				DockerUsername: "some-docker-username",
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
				DefaultRouteHostname: "some-hostname",
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
			CommandLineSettings{
				Memory: 4,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				ProvidedAppPath: "some-path",
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),

		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				NoHostname: true,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				NoRoute: true,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				ProvidedAppPath: "some-app-path",
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				RandomRoute: true,
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),
		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				RoutePath: "some-route-path",
			},
			manifestWithMultipleApps,
			actionerror.CommandLineOptionsWithMultipleAppsError{}),

		Entry("CommandLineOptionsWithMultipleAppsError",
			CommandLineSettings{
				StackName: "some-stackname",
			},
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

		Entry("ValidRoute",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:   "some-name-1",
				Path:   RealPath,
				Routes: []string{"www.hardknox.cli.fun:1234/foo_1+2.html"},
			}},
			nil),

		Entry("InvalidRoute",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:   "some-name-1",
				Path:   RealPath,
				Routes: []string{"http:/www.hardknox.com"},
			}},
			actionerror.InvalidRouteError{Route: "http:/www.hardknox.com"}),

		Entry("InvalidRoute",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:   "some-name-1",
				Path:   RealPath,
				Routes: []string{"I R ROUTE"},
			}},
			actionerror.InvalidRouteError{Route: "I R ROUTE"}),

		Entry("InvalidRoute",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:   "some-name-1",
				Path:   RealPath,
				Routes: []string{"potato"},
			}},
			actionerror.InvalidRouteError{Route: "potato"}),

		// NonexistentAppPathError found in
		// merge_and_validate_settings_and_manifest_unix_test.go and
		// merge_and_validate_settings_and_manifest_windows_test.go

		Entry("MissingNameError", CommandLineSettings{}, nil, actionerror.MissingNameError{}),
		Entry("MissingNameError", CommandLineSettings{}, []manifest.Application{{}}, actionerror.MissingNameError{}),

		Entry("TriggerLegacyPushError",
			CommandLineSettings{},
			[]manifest.Application{{DeprecatedDomain: true}},
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"domain"}}),

		Entry("TriggerLegacyPushError",
			CommandLineSettings{},
			[]manifest.Application{{DeprecatedDomains: true}},
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"domains"}}),

		Entry("TriggerLegacyPushError",
			CommandLineSettings{},
			[]manifest.Application{{DeprecatedHost: true}},
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"host"}}),

		Entry("TriggerLegacyPushError",
			CommandLineSettings{},
			[]manifest.Application{{DeprecatedHosts: true}},
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"hosts"}}),

		Entry("TriggerLegacyPushError",
			CommandLineSettings{},
			[]manifest.Application{{DeprecatedNoHostname: true}},
			actionerror.TriggerLegacyPushError{DomainHostRelated: []string{"no-hostname"}}),

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
		// These represent manifest/commandline conflict errors
		Entry("CommmandLineOptionsAndManifestConflictError",
			CommandLineSettings{
				DefaultRouteDomain: "some-domain",
			},
			[]manifest.Application{{
				Routes: []string{"some-route-1", "some-route-2"},
			}},
			actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			},
		),
		Entry("CommmandLineOptionsAndManifestConflictError",
			CommandLineSettings{
				DefaultRouteHostname: "some-hostname",
			},
			[]manifest.Application{{
				Routes: []string{"some-route-1", "some-route-2"},
			}},
			actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			},
		),
		Entry("CommmandLineOptionsAndManifestConflictError",
			CommandLineSettings{
				NoHostname: true,
			},
			[]manifest.Application{{
				Routes: []string{"some-route-1", "some-route-2"},
			}},
			actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			},
		),
		Entry("CommmandLineOptionsAndManifestConflictError",
			CommandLineSettings{
				RoutePath: "some-route",
			},
			[]manifest.Application{{
				Routes: []string{"some-route-1", "some-route-2"},
			}},
			actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			},
		),

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
