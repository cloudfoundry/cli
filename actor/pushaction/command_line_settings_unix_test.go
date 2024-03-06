//go:build !windows
// +build !windows

package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineSettings with provided path", func() {
	const currentDirectory = "/some/current-directory"

	Describe("OverrideManifestSettings", func() {
		DescribeTable("Path", func(providedAppPath string, manifestPath string, expectedPath string) {
			settings := CommandLineSettings{
				CurrentDirectory: currentDirectory,
				ProvidedAppPath:  providedAppPath,
			}

			app := settings.OverrideManifestSettings(manifest.Application{
				Path: manifestPath,
			})

			Expect(app.Path).To(Equal(expectedPath))
		},

			Entry("path = current directory; provided and manifest paths are empty", "", "", currentDirectory),
			Entry("path = manfiest path; provided is empty and manifest path is not empty", "", "some-manifest-path", "some-manifest-path"),
			Entry("path = absolute provided path; provided relative path is not empty and manifest path is empty", "some-provided-path", "", "some-provided-path"),
			Entry("path = absolute provided path; provided relative path and manifest path are not empty", "some-provided-path", "some-manifest-path", "some-provided-path"),
			Entry("path = provided path; provided path is absolute", "/some-provided-path", "", "/some-provided-path"),
		)

		When("docker image is provided but path is not", func() {
			When("the image is provided via command line setting", func() {
				It("does not set path", func() {
					settings := CommandLineSettings{
						CurrentDirectory: currentDirectory,
						DockerImage:      "some-image",
					}

					app := settings.OverrideManifestSettings(manifest.Application{})

					Expect(app.Path).To(BeEmpty())
				})
			})

			When("the image is provided via manifest", func() {
				It("does not set path", func() {
					settings := CommandLineSettings{
						CurrentDirectory: currentDirectory,
					}

					app := settings.OverrideManifestSettings(manifest.Application{
						DockerImage: "some-image",
					})

					Expect(app.Path).To(BeEmpty())
				})
			})
		})
	})
})
