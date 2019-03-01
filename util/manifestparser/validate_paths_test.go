package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)


var _ = Describe("ValidatePaths", func() {
	var executeErr error

	Describe("paths must exist", func() {
		var tempDir string
		var parser Parser
		var pathToYAMLFile string

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "manifest-push-unit")
			Expect(err).ToNot(HaveOccurred())

			parser = *NewParser()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
		})

		When("all application paths in the manifest do exist", func() {
			It("should exit successfully", func() {
				var err error
				var yamlContents []byte

				pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
				err = os.Mkdir(filepath.Join(tempDir, "path-first-app"), 0755)
				Expect(err).ToNot(HaveOccurred())
				err = os.Mkdir(filepath.Join(tempDir, "path-second-app"), 0755)
				Expect(err).ToNot(HaveOccurred())

				yamlContents = []byte(`---
applications:
    - name: first-app
      path: ./path-first-app/
    - name: second-app
      path: ./path-second-app/
`)
				err = ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				parser.PathToManifest = pathToYAMLFile
				err = parser.Parse(pathToYAMLFile)
				Expect(err).ToNot(HaveOccurred())

				executeErr = parser.Validate()

				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("should not end up permanently changing the working directory", func() {
				var err error
				var yamlContents []byte

				pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
				err = os.Mkdir(filepath.Join(tempDir, "path-first-app"), 0755)
				Expect(err).ToNot(HaveOccurred())
				err = os.Mkdir(filepath.Join(tempDir, "path-second-app"), 0755)
				Expect(err).ToNot(HaveOccurred())

				yamlContents = []byte(`---
applications:
    - name: first-app
      path: ./path-first-app/
    - name: second-app
      path: ./path-second-app/
`)
				err = ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				parser.PathToManifest = pathToYAMLFile
				err = parser.Parse(pathToYAMLFile)
				Expect(err).ToNot(HaveOccurred())

				initialWorkingDir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())

				executeErr = parser.Validate()

				finalWorkingDir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				Expect(initialWorkingDir).To(Equal(finalWorkingDir))
			})
		})

		When("an application path in the manifest does not actually exist", func() {
			var yamlContents []byte

			BeforeEach(func() {
				yamlContents = []byte(`---
applications:
    - name: first-app
      path: /does/not/exist
`)
				pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
				err := ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
				Expect(err).ToNot(HaveOccurred())
			})

			It("exits with an error", func() {
				parser.PathToManifest = pathToYAMLFile
				err := parser.Parse(pathToYAMLFile)
				Expect(err).ToNot(HaveOccurred())

				executeErr = parser.Validate()

				Expect(executeErr).To(HaveOccurred())
			})
		})
	})
})
