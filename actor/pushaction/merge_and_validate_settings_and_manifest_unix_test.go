//go:build !windows
// +build !windows

package pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var (
		actor *Actor
	)

	BeforeEach(func() {
		actor, _, _, _ = getTestPushActor()
	})

	DescribeTable("validation errors",
		func(settings CommandLineSettings, apps []manifest.Application, expectedErr error) {
			_, err := actor.MergeAndValidateSettingsAndManifests(settings, apps)
			if expectedErr == nil {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(expectedErr))
			}
		},

		Entry("NonexistentAppPathError",
			CommandLineSettings{
				Name:            "some-name",
				ProvidedAppPath: "/does-not-exist",
			}, nil,
			actionerror.NonexistentAppPathError{Path: "/does-not-exist"}),

		Entry("NonexistentAppPathError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name: "some-name",
				Path: "/does-not-exist",
			}},
			actionerror.NonexistentAppPathError{Path: "/does-not-exist"}),

		Entry("no NonexistentAppPathError if docker image provided via command line",
			CommandLineSettings{
				Name:        "some-name",
				DockerImage: "some-docker-image",
			}, nil, nil),

		Entry("no NonexistentAppPathError if docker image provided via manifest",
			CommandLineSettings{},
			[]manifest.Application{{
				Name:        "some-name",
				DockerImage: "some-docker-image",
			}}, nil),

		Entry("no NonexistentAppPathError if droplet path provided via command line",
			CommandLineSettings{
				Name:        "some-name",
				DropletPath: "some-droplet-path",
			}, nil, nil),
	)
})
