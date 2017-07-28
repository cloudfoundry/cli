package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineSettings", func() {
	var (
		settings CommandLineSettings
	)

	BeforeEach(func() {
		settings = CommandLineSettings{}
	})

	Describe("ApplicationPath", func() {
		// more tests under command_line_settings_*OS*_test.go

		Context("when ProvidedAppPath is *not* set", func() {
			BeforeEach(func() {
				settings.CurrentDirectory = "current-dir"
			})

			It("returns the CurrentDirectory", func() {
				Expect(settings.ApplicationPath()).To(Equal("current-dir"))
			})
		})
	})

	DescribeTable("OverrideManifestSettings",
		func(settings CommandLineSettings, input manifest.Application, output manifest.Application) {
			Expect(settings.OverrideManifestSettings(input)).To(Equal(output))
		},
		Entry("overrides buildpack name",
			CommandLineSettings{BuildpackName: "not-sixpack"},
			manifest.Application{BuildpackName: "sixpack"},
			manifest.Application{BuildpackName: "not-sixpack"},
		),
		Entry("passes through buildpack name",
			CommandLineSettings{},
			manifest.Application{BuildpackName: "sixpack"},
			manifest.Application{BuildpackName: "sixpack"},
		),
		Entry("overrides command",
			CommandLineSettings{Command: "not-steve"},
			manifest.Application{Command: "steve"},
			manifest.Application{Command: "not-steve"},
		),
		Entry("passes through command",
			CommandLineSettings{},
			manifest.Application{Command: "steve"},
			manifest.Application{Command: "steve"},
		),
		Entry("overrides disk quota",
			CommandLineSettings{DiskQuota: 1024},
			manifest.Application{DiskQuota: 512},
			manifest.Application{DiskQuota: 1024},
		),
		Entry("passes through disk quota",
			CommandLineSettings{},
			manifest.Application{DiskQuota: 1024},
			manifest.Application{DiskQuota: 1024},
		),
		Entry("overrides docker image",
			CommandLineSettings{DockerImage: "not-steve"},
			manifest.Application{DockerImage: "steve"},
			manifest.Application{DockerImage: "not-steve"},
		),
		Entry("passes through docker image",
			CommandLineSettings{},
			manifest.Application{DockerImage: "steve"},
			manifest.Application{DockerImage: "steve"},
		),
		Entry("overrides health check endpoint with '/' when the health check type is http",
			CommandLineSettings{HealthCheckType: "http"},
			manifest.Application{HealthCheckHTTPEndpoint: "/foo"},
			manifest.Application{HealthCheckHTTPEndpoint: "/",
				HealthCheckType: "http"},
		),
		Entry("passes through health check endpoint when the health check type is not http",
			CommandLineSettings{HealthCheckType: "port"},
			manifest.Application{HealthCheckHTTPEndpoint: "/foo"},
			manifest.Application{HealthCheckHTTPEndpoint: "/foo",
				HealthCheckType: "port"},
		),
		Entry("overrides health check timeout",
			CommandLineSettings{HealthCheckTimeout: 1024},
			manifest.Application{HealthCheckTimeout: 512},
			manifest.Application{HealthCheckTimeout: 1024},
		),
		Entry("passes through health check timeout",
			CommandLineSettings{},
			manifest.Application{HealthCheckTimeout: 1024},
			manifest.Application{HealthCheckTimeout: 1024},
		),
		Entry("overrides health check type",
			CommandLineSettings{HealthCheckType: "port"},
			manifest.Application{HealthCheckType: "http"},
			manifest.Application{HealthCheckType: "port"},
		),
		Entry("passes through health check type",
			CommandLineSettings{},
			manifest.Application{HealthCheckType: "http"},
			manifest.Application{HealthCheckType: "http"},
		),
		Entry("overrides instances",
			CommandLineSettings{Instances: 1024},
			manifest.Application{Instances: 512},
			manifest.Application{Instances: 1024},
		),
		Entry("passes through instances",
			CommandLineSettings{},
			manifest.Application{Instances: 1024},
			manifest.Application{Instances: 1024},
		),
		Entry("overrides memory",
			CommandLineSettings{Memory: 1024},
			manifest.Application{Memory: 512},
			manifest.Application{Memory: 1024},
		),
		Entry("passes through memory",
			CommandLineSettings{},
			manifest.Application{Memory: 1024},
			manifest.Application{Memory: 1024},
		),
		Entry("overrides name",
			CommandLineSettings{Name: "not-steve"},
			manifest.Application{Name: "steve"},
			manifest.Application{Name: "not-steve"},
		),
		Entry("passes through name",
			CommandLineSettings{},
			manifest.Application{Name: "steve"},
			manifest.Application{Name: "steve"},
		),
		Entry("overrides stack name",
			CommandLineSettings{StackName: "not-steve"},
			manifest.Application{StackName: "steve"},
			manifest.Application{StackName: "not-steve"},
		),
		Entry("passes through stack name",
			CommandLineSettings{},
			manifest.Application{StackName: "steve"},
			manifest.Application{StackName: "steve"},
		),
	)

	Describe("OverrideManifestSettings", func() {
		// more tests under command_line_settings_*OS*_test.go

		var input, output manifest.Application

		BeforeEach(func() {
			input.Name = "steve"
		})

		JustBeforeEach(func() {
			output = settings.OverrideManifestSettings(input)
		})

		Describe("name", func() {
			Context("when the command line settings provides a name", func() {
				BeforeEach(func() {
					settings.Name = "not-steve"
				})

				It("overrides the name", func() {
					Expect(output.Name).To(Equal("not-steve"))
				})
			})

			Context("when the command line settings name is blank", func() {
				It("passes the manifest name through", func() {
					Expect(output.Name).To(Equal("steve"))
				})
			})
		})
	})
})
