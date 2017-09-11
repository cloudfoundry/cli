// +build !windows

package manifest_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest with paths", func() {
	var (
		pathToManifest string
		manifest       string
	)

	JustBeforeEach(func() {
		tempFile, err := ioutil.TempFile("", "manifest-test-")
		Expect(err).ToNot(HaveOccurred())
		Expect(tempFile.Close()).ToNot(HaveOccurred())
		pathToManifest = tempFile.Name()

		err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
	})

	Describe("ReadAndMergeManifests", func() {
		var (
			apps       []Application
			executeErr error
		)

		JustBeforeEach(func() {
			apps, executeErr = ReadAndMergeManifests(pathToManifest)
		})

		BeforeEach(func() {
			manifest = `---
applications:
- name: "app-1"
  path: /foo
- name: "app-2"
  path: bar
- name: "app-3"
  path: ../baz
`
		})

		It("reads the manifest file", func() {
			tempDir := filepath.Dir(pathToManifest)
			parentTempDir := filepath.Dir(tempDir)
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(apps).To(ConsistOf(
				Application{Name: "app-1", Path: "/foo"},
				Application{Name: "app-2", Path: filepath.Join(tempDir, "bar")},
				Application{Name: "app-3", Path: filepath.Join(parentTempDir, "baz")},
			))
		})
	})
})
