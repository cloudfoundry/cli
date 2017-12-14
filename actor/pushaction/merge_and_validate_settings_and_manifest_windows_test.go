// +build windows

package pushaction_test

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {

	var (
		actor       *Actor
		cmdSettings CommandLineSettings

		currentDirectory string
	)

	BeforeEach(func() {
		actor = NewActor(nil, nil)
		currentDirectory = getCurrentDir()
	})

	Describe("sanitizing values", func() {
		var (
			apps       []manifest.Application
			mergedApps []manifest.Application
			executeErr error
		)

		BeforeEach(func() {
			cmdSettings = CommandLineSettings{
				CurrentDirectory: currentDirectory,
			}

			apps = []manifest.Application{
				{Name: "app-1"},
			}
		})

		JustBeforeEach(func() {
			mergedApps, executeErr = actor.MergeAndValidateSettingsAndManifests(cmdSettings, apps)
		})

		Context("when app path '\\' is set from the command line", func() {
			BeforeEach(func() {
				cmdSettings.ProvidedAppPath = `\`
			})

			It("sets the app path to the provided path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				abs, err := filepath.Abs(`\`)
				Expect(err).ToNot(HaveOccurred())
				Expect(mergedApps[0].Path).To(Equal(abs))
			})
		})

		Context("when app path '\\' is set from the manifest", func() {
			BeforeEach(func() {
				apps[0].Path = `\`
			})

			It("sets the app path to the provided path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				abs, err := filepath.Abs(`\`)
				Expect(err).ToNot(HaveOccurred())
				Expect(mergedApps[0].Path).To(Equal(abs))
			})
		})
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
				ProvidedAppPath: "C:\\does-not-exist",
			}, nil,
			actionerror.NonexistentAppPathError{Path: "C:\\does-not-exist"}),

		Entry("NonexistentAppPathError",
			CommandLineSettings{},
			[]manifest.Application{{
				Name: "some-name",
				Path: "C:\\does-not-exist",
			}},
			actionerror.NonexistentAppPathError{Path: "C:\\does-not-exist"}),
	)

})
