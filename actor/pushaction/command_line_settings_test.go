package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"

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

	DescribeTable("OverrideManifestSettings",
		func(settings CommandLineSettings, input manifest.Application, output manifest.Application) {
			Expect(settings.OverrideManifestSettings(input)).To(Equal(output))
		},
		Entry("overrides buildpack name",
			CommandLineSettings{Buildpack: types.FilteredString{IsSet: true, Value: "sixpack"}},
			manifest.Application{Buildpack: types.FilteredString{IsSet: true, Value: "not-sixpack"}},
			manifest.Application{Buildpack: types.FilteredString{IsSet: true, Value: "sixpack"}},
		),
		Entry("passes through buildpack name",
			CommandLineSettings{Buildpack: types.FilteredString{IsSet: false, Value: ""}},
			manifest.Application{Buildpack: types.FilteredString{IsSet: true, Value: "not-sixpack"}},
			manifest.Application{Buildpack: types.FilteredString{IsSet: true, Value: "not-sixpack"}},
		),
		Entry("overrides command",
			CommandLineSettings{Command: types.FilteredString{IsSet: true, Value: "not-steve"}},
			manifest.Application{Command: types.FilteredString{IsSet: true, Value: "steve"}},
			manifest.Application{Command: types.FilteredString{IsSet: true, Value: "not-steve"}},
		),
		Entry("passes through command",
			CommandLineSettings{},
			manifest.Application{Command: types.FilteredString{IsSet: true, Value: "steve"}},
			manifest.Application{Command: types.FilteredString{IsSet: true, Value: "steve"}},
		),
		Entry("overrides domain",
			CommandLineSettings{DefaultRouteDomain: "not-steve"},
			manifest.Application{Domain: "steve"},
			manifest.Application{Domain: "not-steve"},
		),
		Entry("passes through domain",
			CommandLineSettings{},
			manifest.Application{Domain: "steve"},
			manifest.Application{Domain: "steve"},
		),
		Entry("overrides hostname",
			CommandLineSettings{DefaultRouteHostname: "not-steve"},
			manifest.Application{Hostname: "steve"},
			manifest.Application{Hostname: "not-steve"},
		),
		Entry("passes through hostname",
			CommandLineSettings{},
			manifest.Application{Hostname: "steve"},
			manifest.Application{Hostname: "steve"},
		),
		Entry("overrides disk quota",
			CommandLineSettings{DiskQuota: 1024},
			manifest.Application{DiskQuota: types.NullByteSizeInMb{Value: 512, IsSet: true}},
			manifest.Application{DiskQuota: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
		),
		Entry("passes through disk quota",
			CommandLineSettings{},
			manifest.Application{DiskQuota: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
			manifest.Application{DiskQuota: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
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
		Entry("overrides docker username",
			CommandLineSettings{DockerUsername: "not-steve"},
			manifest.Application{DockerUsername: "steve"},
			manifest.Application{DockerUsername: "not-steve"},
		),
		Entry("passes through docker username",
			CommandLineSettings{},
			manifest.Application{DockerUsername: "steve"},
			manifest.Application{DockerUsername: "steve"},
		),
		Entry("overrides docker password",
			CommandLineSettings{DockerPassword: "not-steve"},
			manifest.Application{DockerPassword: "steve"},
			manifest.Application{DockerPassword: "not-steve"},
		),
		Entry("passes through docker password",
			CommandLineSettings{},
			manifest.Application{DockerPassword: "steve"},
			manifest.Application{DockerPassword: "steve"},
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
			CommandLineSettings{Instances: types.NullInt{Value: 1024, IsSet: true}},
			manifest.Application{Instances: types.NullInt{Value: 512, IsSet: true}},
			manifest.Application{Instances: types.NullInt{Value: 1024, IsSet: true}},
		),
		Entry("passes through instances",
			CommandLineSettings{},
			manifest.Application{Instances: types.NullInt{Value: 1024, IsSet: true}},
			manifest.Application{Instances: types.NullInt{Value: 1024, IsSet: true}},
		),
		Entry("overrides memory",
			CommandLineSettings{Memory: 1024},
			manifest.Application{Memory: types.NullByteSizeInMb{Value: 512, IsSet: true}},
			manifest.Application{Memory: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
		),
		Entry("passes through memory",
			CommandLineSettings{},
			manifest.Application{Memory: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
			manifest.Application{Memory: types.NullByteSizeInMb{Value: 1024, IsSet: true}},
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
		Entry("overrides no hostname",
			CommandLineSettings{NoHostname: true},
			manifest.Application{},
			manifest.Application{NoHostname: true},
		),
		Entry("passes through no hostname",
			CommandLineSettings{},
			manifest.Application{NoHostname: true},
			manifest.Application{NoHostname: true},
		),
		Entry("overrides no route",
			CommandLineSettings{NoRoute: true},
			manifest.Application{},
			manifest.Application{NoRoute: true},
		),
		Entry("passes through no route",
			CommandLineSettings{},
			manifest.Application{NoRoute: true},
			manifest.Application{NoRoute: true},
		),
		Entry("overrides random route",
			CommandLineSettings{RandomRoute: true},
			manifest.Application{},
			manifest.Application{RandomRoute: true},
		),
		Entry("passes through random route",
			CommandLineSettings{},
			manifest.Application{RandomRoute: true},
			manifest.Application{RandomRoute: true},
		),
		Entry("overrides route path",
			CommandLineSettings{RoutePath: "/some-route-path"},
			manifest.Application{},
			manifest.Application{RoutePath: "/some-route-path"},
		),
		Entry("passes through route path",
			CommandLineSettings{},
			manifest.Application{RoutePath: "/some-route-path"},
			manifest.Application{RoutePath: "/some-route-path"},
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
