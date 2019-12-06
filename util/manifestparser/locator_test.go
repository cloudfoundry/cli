package manifestparser_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/cli/util/manifestparser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Locator", func() {
	var locator *Locator

	BeforeEach(func() {
		locator = NewLocator()
	})

	Describe("Path", func() {
		var (
			originalFilepathOrDirectory string
			filepathOrDirectory         string

			expectedPath string
			exists       bool
			executeErr   error

			workingDir string
		)

		BeforeEach(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "manifest-locator-working-dir")
			Expect(err).ToNot(HaveOccurred())
			workingDir, err = filepath.EvalSymlinks(workingDir)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(workingDir)).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			expectedPath, exists, executeErr = locator.Path(filepathOrDirectory)
		})

		When("given a file path", func() {
			BeforeEach(func() {
				filepathOrDirectory = filepath.Join(workingDir, "some-manifest.yml")
				Expect(ioutil.WriteFile(filepathOrDirectory, nil, 0600)).To(Succeed())
			})

			It("returns the path and true", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPath).To(Equal(filepathOrDirectory))
				Expect(exists).To(BeTrue())
			})
		})

		When("given a directory", func() {
			BeforeEach(func() {
				filepathOrDirectory = workingDir
			})

			When("a manifest.yml exists in the directory", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "manifest.yml"), nil, 0600)).To(Succeed())
				})

				It("returns the path and true", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPath).To(Equal(filepath.Join(workingDir, "manifest.yml")))
					Expect(exists).To(BeTrue())
				})
			})

			When("a manifest.yaml exists in the directory", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(workingDir, "manifest.yaml"), nil, 0600)).To(Succeed())
				})

				It("returns the path and true", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPath).To(Equal(filepath.Join(workingDir, "manifest.yaml")))
					Expect(exists).To(BeTrue())
				})
			})

			When("no manifest exists in the directory", func() {
				It("returns empty and false", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(expectedPath).To(BeEmpty())
					Expect(exists).To(BeFalse())
				})
			})

			When("there's an error detecting the manifest", func() {
				BeforeEach(func() {
					locator.FilesToCheckFor = []string{strings.Repeat("a", 10000)}
				})

				It("returns the error", func() {
					Expect(executeErr).To(HaveOccurred())
				})
			})
		})

		When("the provided filepathOrDirectory does not exist", func() {
			BeforeEach(func() {
				filepathOrDirectory = "/does/not/exist"
			})

			It("returns empty and false", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPath).To(BeEmpty())
				Expect(exists).To(BeFalse())
			})
		})

		When("there's an error detecting the manifest", func() {
			BeforeEach(func() {
				filepathOrDirectory = filepath.Join(workingDir, strings.Repeat("a", 10000))
			})

			It("returns the error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(expectedPath).To(BeEmpty())
				Expect(exists).To(BeFalse())
			})
		})

		When("the path to the manifest is a symbolic link", func() {
			BeforeEach(func() {
				originalFilepathOrDirectory = filepath.Join(workingDir, "some-manifest.yml")
				Expect(ioutil.WriteFile(originalFilepathOrDirectory, nil, 0600)).To(Succeed())
				filepathOrDirectory = filepath.Join(workingDir, "link-to-some-manifest.yml")
				err := os.Symlink(originalFilepathOrDirectory, filepathOrDirectory)
				Expect(err).To(BeNil())
			})

			It("returns the absolute path, not the symbolic link", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(expectedPath).To(Equal(originalFilepathOrDirectory))
				Expect(exists).To(BeTrue())
			})
		})
	})
})
