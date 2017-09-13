// +build windows

package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineSettings with provided path", func() {
	const currentDirectory = "C:\\some\\current-directory"

	Describe("ApplicationPath with provided path", func() {
		DescribeTable("ApplicationPath", func(providedAppPath string, expectedPath string) {
			settings := CommandLineSettings{
				CurrentDirectory: currentDirectory,
				ProvidedAppPath:  providedAppPath,
			}

			Expect(settings.ApplicationPath()).To(Equal(expectedPath))
		},

			Entry("path = provided path; provided path is absolute", "C:\\some\\path", "C:\\some\\path"),
			Entry("path = full path to provided path; provided path is relative", ".\\some-path", "C:\\some\\current-directory\\some-path"),
		)
	})

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
			Entry("path = absolute provided path; provided relative path is not empty and manifest path is empty", "some-provided-path", "", "C:\\some\\current-directory\\some-provided-path"),
			Entry("path = absolute provided path; provided relative path and manifest path are not empty", "some-provided-path", "some-manifest-path", "C:\\some\\current-directory\\some-provided-path"),
			Entry("path = provided path; provided path is absolute", "C:\\some-provided-path", "", "C:\\some-provided-path"),
		)
	})
})
