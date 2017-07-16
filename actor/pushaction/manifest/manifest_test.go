package manifest_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/pushaction/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
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
- name: "app-2"
`
		})

		It("reads the manifest file", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(apps).To(ConsistOf(
				Application{Name: "app-1"},
				Application{Name: "app-2"},
			))
		})
	})
})
