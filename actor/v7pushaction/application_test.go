package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Applications", func() {
	Describe("Application", func() {
		DescribeTable("CalculatedBuildpacks",
			func(v2Buildpack string, v3Buildpacks []string, expected []string) {
				var buildpack types.FilteredString
				if len(v2Buildpack) > 0 {
					buildpack = types.FilteredString{
						Value: v2Buildpack,
						IsSet: true,
					}
				}
				Expect(Application{
					Application: v2action.Application{
						Buildpack: buildpack,
					},
					Buildpacks: v3Buildpacks,
				}.CalculatedBuildpacks()).To(Equal(expected))
			},

			Entry("returns buildpacks when it contains values",
				"some-buildpack", []string{"some-buildpack", "some-other-buildpack"},
				[]string{"some-buildpack", "some-other-buildpack"}),

			Entry("always returns buildpacks when it is set",
				"some-buildpack", []string{},
				[]string{}),

			Entry("returns v2 buildpack when buildpacks is not set",
				"some-buildpack", nil,
				[]string{"some-buildpack"}),

			Entry("returns empty when nothing is set", "", nil, nil),
		)
	})
})
