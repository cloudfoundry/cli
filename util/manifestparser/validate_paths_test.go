package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	"fmt"
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
			It("should successfully handle absolute paths", func() {
				var err error
				var yamlContents []byte
				var yamlString string

				pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
				firstPath := filepath.Join(tempDir, "path-first-app")
				err = os.Mkdir(firstPath, 0755)
				Expect(err).ToNot(HaveOccurred())
				secondPath := filepath.Join(tempDir, "path-second-app")
				err = os.Mkdir(secondPath, 0755)
				Expect(err).ToNot(HaveOccurred())

				yamlString = `---
applications:
    - name: first-app
      path: %s
    - name: second-app
      path: %s
`
				interpolatedManifest := fmt.Sprintf(yamlString, firstPath, secondPath)
				yamlContents = []byte(interpolatedManifest)

				err = ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				parser.PathToManifest = pathToYAMLFile
				err = parser.Parse(pathToYAMLFile)
				Expect(err).ToNot(HaveOccurred())

				executeErr = parser.Validate()

				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("should successfully handle relative paths", func() {
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
