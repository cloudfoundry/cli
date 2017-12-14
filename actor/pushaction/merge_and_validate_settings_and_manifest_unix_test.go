// +build !windows

package pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var (
		actor *Actor
	)

	BeforeEach(func() {
		actor = NewActor(nil, nil)
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
	)
})
