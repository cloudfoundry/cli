package flag_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("path types", func() {
	type completable interface {
		Complete(string) []flags.Completion
	}

	var (
		currentDir string
		tempDir    string
	)

	testComplete := func(subject completable) {
		Describe("Complete", func() {
			When("the prefix is empty", func() {
				It("returns all files and directories", func() {
					matches := subject.Complete("")
					Expect(matches).To(ConsistOf(
						flags.Completion{Item: "abc"},
						flags.Completion{Item: "abd"},
						flags.Completion{Item: fmt.Sprintf("~add%c", os.PathSeparator)},
						flags.Completion{Item: "~abd"},
						flags.Completion{Item: fmt.Sprintf("add%c", os.PathSeparator)},
						flags.Completion{Item: fmt.Sprintf("aee%c", os.PathSeparator)},
						flags.Completion{Item: "tfg"},
						flags.Completion{Item: "ABCD"},
					))
				})
			})

			When("the prefix is not empty", func() {
				When("there are matching paths", func() {
					It("returns the matching paths", func() {
						matches := subject.Complete("a")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "abc"},
							flags.Completion{Item: "abd"},
							flags.Completion{Item: fmt.Sprintf("add%c", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("aee%c", os.PathSeparator)},
						))
					})

					It("is case sensitive", func() {
						matches := subject.Complete("A")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "ABCD"},
						))
					})

					It("finds files starting with '~'", func() {
						matches := subject.Complete("~")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "~abd"},
							flags.Completion{Item: fmt.Sprintf("~add%c", os.PathSeparator)},
						))
					})
				})

				When("there are no matching paths", func() {
					It("returns no matches", func() {
						Expect(subject.Complete("z")).To(BeEmpty())
					})
				})
			})

			When("the prefix is ~/", func() {
				var prevHome string

				BeforeEach(func() {
					prevHome = os.Getenv("HOME")
				})

				AfterEach(func() {
					os.Setenv("HOME", prevHome)
				})

				When("$HOME is set", func() {
					var (
						tempDir string
						err     error
					)

					BeforeEach(func() {
						tempDir, err = ioutil.TempDir("", "")
						Expect(err).ToNot(HaveOccurred())
						os.Setenv("HOME", tempDir)

						for _, filename := range []string{"abc", "def"} {
							err = ioutil.WriteFile(filepath.Join(tempDir, filename), []byte{}, 0400)
							Expect(err).ToNot(HaveOccurred())
						}

						for _, dir := range []string{"adir", "bdir"} {
							err = os.Mkdir(filepath.Join(tempDir, dir), 0700)
							Expect(err).ToNot(HaveOccurred())
						}
					})

					AfterEach(func() {
						err = os.RemoveAll(tempDir)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns matching paths in $HOME", func() {
						matches := subject.Complete(fmt.Sprintf("~%c", os.PathSeparator))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: fmt.Sprintf("~%cabc", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("~%cdef", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("~%cadir%c", os.PathSeparator, os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("~%cbdir%c", os.PathSeparator, os.PathSeparator)},
						))
					})
				})
			})

			When("the prefix starts with ~/", func() {
				var prevHome string

				BeforeEach(func() {
					prevHome = os.Getenv("HOME")
				})

				AfterEach(func() {
					os.Setenv("HOME", prevHome)
				})

				When("$HOME is set", func() {
					var (
						tempDir string
						err     error
					)

					BeforeEach(func() {
						tempDir, err = ioutil.TempDir("", "")
						Expect(err).ToNot(HaveOccurred())
						os.Setenv("HOME", tempDir)

						for _, filename := range []string{"abc", "def"} {
							err = ioutil.WriteFile(filepath.Join(tempDir, filename), []byte{}, 0400)
							Expect(err).ToNot(HaveOccurred())
						}

						for _, dir := range []string{"adir", "bdir"} {
							err = os.Mkdir(filepath.Join(tempDir, dir), 0700)
							Expect(err).ToNot(HaveOccurred())
						}
					})

					AfterEach(func() {
						err = os.RemoveAll(tempDir)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns matching paths in $HOME", func() {
						matches := subject.Complete(fmt.Sprintf("~%ca", os.PathSeparator))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: fmt.Sprintf("~%cabc", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("~%cadir%c", os.PathSeparator, os.PathSeparator)},
						))
					})
				})
			})
		})
	}

	BeforeEach(func() {
		var err error
		currentDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(tempDir)
		Expect(err).ToNot(HaveOccurred())

		for _, filename := range []string{"abc", "abd", "~abd", "tfg", "ABCD"} {
			err = ioutil.WriteFile(filename, []byte{}, 0400)
			Expect(err).ToNot(HaveOccurred())
		}

		for _, dir := range []string{"~add", "add", "aee"} {
			err := os.Mkdir(dir, 0700)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		err := os.Chdir(currentDir)
		Expect(err).ToNot(HaveOccurred())
		err = os.RemoveAll(tempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Path", func() {
		var path Path
		testComplete(path)
	})

	Describe("PathWithExistenceCheck", func() {
		var pathWithExistenceCheck PathWithExistenceCheck

		BeforeEach(func() {
			pathWithExistenceCheck = PathWithExistenceCheck("")
		})

		testComplete(pathWithExistenceCheck)

		Describe("UnmarshalFlag", func() {
			When("the path does not exist", func() {
				It("returns a path does not exist error", func() {
					err := pathWithExistenceCheck.UnmarshalFlag("./some-dir/some-file")
					Expect(err).To(MatchError(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "The specified path './some-dir/some-file' does not exist.",
					}))
				})
			})

			When("the path exists", func() {
				When("the path is relative", func() {
					It("expands the path", func() {
						fullPath := filepath.Join(tempDir, "abc")
						err := pathWithExistenceCheck.UnmarshalFlag("abc")
						Expect(err).ToNot(HaveOccurred())
						Expect(string(pathWithExistenceCheck)).To(matchers.MatchPath(fullPath))
					})
				})

				When("the path is absolute", func() {
					It("preserves the path", func() {
						err := pathWithExistenceCheck.UnmarshalFlag(filepath.Join(tempDir, "abc"))
						Expect(err).ToNot(HaveOccurred())
						Expect(pathWithExistenceCheck).To(BeEquivalentTo(filepath.Join(tempDir, "abc")))
					})
				})
			})
		})
	})

	Describe("ManifestPathWithExistenceCheck", func() {
		var manifestPathWithExistenceCheck ManifestPathWithExistenceCheck

		BeforeEach(func() {
			manifestPathWithExistenceCheck = ManifestPathWithExistenceCheck("")
		})

		testComplete(manifestPathWithExistenceCheck)

		Describe("UnmarshalFlag", func() {
			When("the path does not exist", func() {
				It("returns a path does not exist error", func() {
					err := manifestPathWithExistenceCheck.UnmarshalFlag("./some-dir/some-file")
					Expect(err).To(MatchError(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "The specified path './some-dir/some-file' does not exist.",
					}))
				})
			})

			When("the path is a directory, and exists, but does not contain a manifest.{yml,yaml} file", func() {
				It("returns a path does not exist error", func() {
					err := manifestPathWithExistenceCheck.UnmarshalFlag(tempDir)
					Expect(err).To(MatchError(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "The specified directory '" + tempDir + "' does not contain a file named 'manifest.yml'.",
					}))
				})
			})

			When("the path exists", func() {
				It("sets the path", func() {
					err := manifestPathWithExistenceCheck.UnmarshalFlag("abc")
					Expect(err).ToNot(HaveOccurred())
					Expect(manifestPathWithExistenceCheck).To(BeEquivalentTo("abc"))
				})
			})
		})
	})

	Describe("JSONOrFileWithValidation", func() {
		var jsonOrFile JSONOrFileWithValidation

		BeforeEach(func() {
			jsonOrFile = JSONOrFileWithValidation{}
		})

		testComplete(jsonOrFile)

		Describe("UnmarshalFlag", func() {
			When("the file exists", func() {
				var tempPath string

				When("the file has valid JSON", func() {
					BeforeEach(func() {
						tempPath = tempFile(`{"this is":"valid JSON"}`)
					})

					It("reads and unmarshals the JSON from the file", func() {
						err := jsonOrFile.UnmarshalFlag(tempPath)
						Expect(err).ToNot(HaveOccurred())
						Expect(jsonOrFile).To(Equal(JSONOrFileWithValidation{
							IsSet: true,
							Value: map[string]interface{}{
								"this is": "valid JSON",
							},
						}))
					})
				})

				When("the file has invalid JSON", func() {
					BeforeEach(func() {
						tempPath = tempFile(`{"this is":"invalid JSON"`)
					})

					It("errors with the invalid configuration error", func() {
						err := jsonOrFile.UnmarshalFlag(tempPath)
						Expect(err).To(Equal(&flags.Error{
							Type:    flags.ErrRequired,
							Message: "Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object.",
						}))
					})
				})
			})

			When("the JSON is invalid", func() {
				It("errors with the invalid configuration error", func() {
					err := jsonOrFile.UnmarshalFlag(`{"this is":"invalid JSON"`)
					Expect(err).To(Equal(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object.",
					}))
				})
			})

			When("the JSON is valid", func() {
				It("reads and unmarshals the JSON", func() {
					err := jsonOrFile.UnmarshalFlag(`{"this is":"valid JSON"}`)
					Expect(err).ToNot(HaveOccurred())
					Expect(jsonOrFile).To(Equal(JSONOrFileWithValidation{
						IsSet: true,
						Value: map[string]interface{}{
							"this is": "valid JSON",
						},
					}))
				})
			})
		})
	})

	Describe("PathWithExistenceCheckOrURL", func() {
		var pathWithExistenceCheckOrURL PathWithExistenceCheckOrURL

		BeforeEach(func() {
			pathWithExistenceCheckOrURL = PathWithExistenceCheckOrURL("")
		})

		testComplete(pathWithExistenceCheckOrURL)

		Describe("UnmarshalFlag", func() {
			When("the path is a URL", func() {
				It("sets the path if it starts with 'http://'", func() {
					err := pathWithExistenceCheckOrURL.UnmarshalFlag("http://example.com/payload.tgz")
					Expect(err).ToNot(HaveOccurred())
					Expect(pathWithExistenceCheckOrURL).To(BeEquivalentTo("http://example.com/payload.tgz"))
				})

				It("sets the path if it starts with 'https://'", func() {
					err := pathWithExistenceCheckOrURL.UnmarshalFlag("https://example.com/payload.tgz")
					Expect(err).ToNot(HaveOccurred())
					Expect(pathWithExistenceCheckOrURL).To(BeEquivalentTo("https://example.com/payload.tgz"))
				})
			})

			When("the path does not exist", func() {
				It("returns a path does not exist error", func() {
					err := pathWithExistenceCheckOrURL.UnmarshalFlag("./some-dir/some-file")
					Expect(err).To(MatchError(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "The specified path './some-dir/some-file' does not exist.",
					}))
				})
			})

			When("the path exists", func() {
				It("sets the path", func() {
					err := pathWithExistenceCheckOrURL.UnmarshalFlag("abc")
					Expect(err).ToNot(HaveOccurred())
					Expect(pathWithExistenceCheckOrURL).To(BeEquivalentTo("abc"))
				})
			})
		})
	})

	Describe("PathWithAt", func() {
		var pathWithAt PathWithAt

		Describe("Complete", func() {
			When("the prefix is empty", func() {
				It("returns no matches", func() {
					Expect(pathWithAt.Complete("")).To(BeEmpty())
				})
			})

			When("the prefix doesn't start with @", func() {
				It("returns no matches", func() {
					Expect(pathWithAt.Complete("a@b")).To(BeEmpty())
				})
			})

			When("the prefix starts with @", func() {
				When("there are no characters after the @", func() {
					It("returns all files and directories", func() {
						matches := pathWithAt.Complete("@")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "@abc"},
							flags.Completion{Item: "@abd"},
							flags.Completion{Item: fmt.Sprintf("@~add%c", os.PathSeparator)},
							flags.Completion{Item: "@~abd"},
							flags.Completion{Item: fmt.Sprintf("@add%c", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("@aee%c", os.PathSeparator)},
							flags.Completion{Item: "@tfg"},
							flags.Completion{Item: "@ABCD"},
						))
					})
				})

				When("there are characters after the @", func() {
					When("there are matching paths", func() {
						It("returns the matching paths", func() {
							matches := pathWithAt.Complete("@a")
							Expect(matches).To(ConsistOf(
								flags.Completion{Item: "@abc"},
								flags.Completion{Item: "@abd"},
								flags.Completion{Item: fmt.Sprintf("@add%c", os.PathSeparator)},
								flags.Completion{Item: fmt.Sprintf("@aee%c", os.PathSeparator)},
							))
						})

						It("is case sensitive", func() {
							matches := pathWithAt.Complete("@A")
							Expect(matches).To(ConsistOf(
								flags.Completion{Item: "@ABCD"},
							))
						})
					})

					When("there are no matching paths", func() {
						It("returns no matches", func() {
							Expect(pathWithAt.Complete("@z")).To(BeEmpty())
						})
					})
				})
			})

			When("the prefix is @~/", func() {
				var prevHome string

				BeforeEach(func() {
					prevHome = os.Getenv("HOME")
				})

				AfterEach(func() {
					os.Setenv("HOME", prevHome)
				})

				When("$HOME is set", func() {
					var (
						tempDir string
						err     error
					)

					BeforeEach(func() {
						tempDir, err = ioutil.TempDir("", "")
						Expect(err).ToNot(HaveOccurred())
						os.Setenv("HOME", tempDir)

						for _, filename := range []string{"abc", "def"} {
							err = ioutil.WriteFile(filepath.Join(tempDir, filename), []byte{}, 0400)
							Expect(err).ToNot(HaveOccurred())
						}

						for _, dir := range []string{"adir", "bdir"} {
							err = os.Mkdir(filepath.Join(tempDir, dir), 0700)
							Expect(err).ToNot(HaveOccurred())
						}
					})

					AfterEach(func() {
						err = os.RemoveAll(tempDir)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns matching paths in $HOME", func() {
						matches := pathWithAt.Complete(fmt.Sprintf("@~%c", os.PathSeparator))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: fmt.Sprintf("@~%cabc", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("@~%cdef", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("@~%cadir%c", os.PathSeparator, os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("@~%cbdir%c", os.PathSeparator, os.PathSeparator)},
						))
					})
				})
			})

			When("the prefix starts with @~/", func() {
				var prevHome string

				BeforeEach(func() {
					prevHome = os.Getenv("HOME")
				})

				AfterEach(func() {
					os.Setenv("HOME", prevHome)
				})

				When("$HOME is set", func() {
					var (
						tempDir string
						err     error
					)

					BeforeEach(func() {
						tempDir, err = ioutil.TempDir("", "")
						Expect(err).ToNot(HaveOccurred())
						os.Setenv("HOME", tempDir)

						for _, filename := range []string{"abc", "def"} {
							err = ioutil.WriteFile(filepath.Join(tempDir, filename), []byte{}, 0400)
							Expect(err).ToNot(HaveOccurred())
						}

						for _, dir := range []string{"adir", "bdir"} {
							err = os.Mkdir(filepath.Join(tempDir, dir), 0700)
							Expect(err).ToNot(HaveOccurred())
						}
					})

					AfterEach(func() {
						err = os.RemoveAll(tempDir)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns matching paths in $HOME", func() {
						matches := pathWithAt.Complete(fmt.Sprintf("@~%ca", os.PathSeparator))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: fmt.Sprintf("@~%cabc", os.PathSeparator)},
							flags.Completion{Item: fmt.Sprintf("@~%cadir%c", os.PathSeparator, os.PathSeparator)},
						))
					})
				})
			})
		})
	})

	Describe("PathWithBool", func() {
		var pathWithBool PathWithBool

		Describe("Complete", func() {
			When("the prefix is empty", func() {
				It("returns bool choices and all files and directories", func() {
					matches := pathWithBool.Complete("")
					Expect(matches).To(ConsistOf(
						flags.Completion{Item: "true"},
						flags.Completion{Item: "false"},
						flags.Completion{Item: "abc"},
						flags.Completion{Item: "abd"},
						flags.Completion{Item: fmt.Sprintf("add%c", os.PathSeparator)},
						flags.Completion{Item: "~abd"},
						flags.Completion{Item: fmt.Sprintf("~add%c", os.PathSeparator)},
						flags.Completion{Item: fmt.Sprintf("aee%c", os.PathSeparator)},
						flags.Completion{Item: "tfg"},
						flags.Completion{Item: "ABCD"},
					))
				})
			})

			When("the prefix is not empty", func() {
				When("there are matching bool/paths", func() {
					It("returns the matching bool/paths", func() {
						matches := pathWithBool.Complete("t")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "true"},
							flags.Completion{Item: "tfg"},
						))
					})

					It("paths are case sensitive", func() {
						matches := pathWithBool.Complete("A")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "ABCD"},
						))
					})

					It("bools are not case sensitive", func() {
						matches := pathWithBool.Complete("Tr")
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "true"},
						))
					})
				})

				When("there are no matching bool/paths", func() {
					It("returns no matches", func() {
						Expect(pathWithBool.Complete("z")).To(BeEmpty())
					})
				})
			})
		})
	})
})
