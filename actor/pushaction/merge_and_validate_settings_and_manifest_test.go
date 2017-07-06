package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var (
		actor       *Actor
		cmdSettings CommandLineSettings
	)

	BeforeEach(func() {
		actor = NewActor(nil)
	})

	Context("when only passed command line settings", func() {
		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: "/current/directory",
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
				Path:        "/current/directory",
			}}))
		})
	})

	Context("when passed command line settings and manifests", func() {
		var apps []manifest.Application

		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: "/current/directory",
			}

			apps = []manifest.Application{
				{Name: "app-1"},
				{Name: "app-2"},
			}
		})

		It("merges command line settings and manifest apps", func() {
			mergedApps, err := actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
			Expect(err).ToNot(HaveOccurred())

			Expect(mergedApps).To(HaveLen(2))
			Expect(mergedApps).To(ConsistOf(
				manifest.Application{
					Name: "app-1",
					Path: "/current/directory",
				},
				manifest.Application{
					Name: "app-2",
					Path: "/current/directory",
				},
			))
		})
	})

	DescribeTable("validation errors",
		func(settings CommandLineSettings, apps []manifest.Application, expectedErr error) {
			_, err := actor.MergeAndValidateSettingsAndManifests(settings, apps)
			Expect(err).To(MatchError(expectedErr))
		},

		Entry("MissingNameError", CommandLineSettings{}, nil, MissingNameError{}),
		Entry("MissingNameError", CommandLineSettings{}, []manifest.Application{{}}, MissingNameError{}),
	)
})
