package pushaction_test

import (
	"os"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var actor *Actor

	BeforeEach(func() {
		actor = NewActor(nil)
	})

	Context("when only passed command line settings", func() {
		var cmdSettings CommandLineSettings

		Context("when pushing the current directory", func() {
			var pwd string

			BeforeEach(func() {
				var err error
				pwd, err = os.Getwd()
				Expect(err).ToNot(HaveOccurred())

				cmdSettings = CommandLineSettings{
					Name:             "some-app",
					CurrentDirectory: pwd,
					DockerImage:      "some-image",
				}
			})

			It("creates a manifest with only the command line settings", func() {
				manifests, err := actor.MergeAndValidateSettingsAndManifests(cmdSettings, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(manifests).To(Equal([]manifest.Application{{
					Name:        "some-app",
					Path:        pwd,
					DockerImage: "some-image",
				}}))
			})
		})

		Context("when the command line settings directory path is not empty", func() {
			BeforeEach(func() {
				cmdSettings = CommandLineSettings{
					CurrentDirectory: "some-current-directory",
					AppPath:          "some-directory-path",
				}
			})

			It("uses the setting directory in the application manifest", func() {
				manifests, err := actor.MergeAndValidateSettingsAndManifests(cmdSettings, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(manifests).To(HaveLen(1))
				Expect(manifests[0].Path).To(Equal("some-directory-path"))
			})
		})
	})

	Context("when passed command line settings and manifests", func() {
		// fill in here
	})
})
