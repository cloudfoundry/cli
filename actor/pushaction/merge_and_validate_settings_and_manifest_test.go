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
		var (
			cmdSettings CommandLineSettings
			pwd         string
		)

		BeforeEach(func() {
			var err error
			pwd, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			cmdSettings = CommandLineSettings{
				Name:             "some-app",
				CurrentDirectory: pwd,
			}
		})

		It("creates a manifest with only the command line settings", func() {
			manifests, err := actor.MergeAndValidateSettingsAndManifests(cmdSettings, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(manifests).To(Equal([]manifest.Application{{
				Name: "some-app",
				Path: pwd,
			}}))
		})
	})

	Context("when passed command line settings and manifests", func() {
		// fill in here
	})
})
