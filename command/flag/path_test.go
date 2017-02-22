package flag_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("path flag types", func() {
	var (
		currentDir string
		tempDir    string
	)

	BeforeEach(func() {
		var err error
		currentDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(tempDir)
		Expect(err).ToNot(HaveOccurred())

		for _, filename := range []string{"abc", "abd", "efg"} {
			err = ioutil.WriteFile(filename, []byte{}, 0400)
			Expect(err).ToNot(HaveOccurred())
		}

		for _, dir := range []string{"add", "aee"} {
			err := os.Mkdir(dir, os.ModeDir)
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

		BeforeEach(func() {
			path = Path("")
		})

		Describe("Complete", func() {
			Context("when the prefix is empty", func() {
				It("returns all files and directories", func() {
					Expect(path.Complete("")).To(ConsistOf(
						flags.Completion{Item: "abc"},
						flags.Completion{Item: "abd"},
						flags.Completion{Item: "add/"},
						flags.Completion{Item: "aee/"},
						flags.Completion{Item: "efg"},
					))
				})
			})

			Context("when the prefix is not empty", func() {
				Context("when there are matching paths", func() {
					It("returns the matching paths", func() {
						Expect(path.Complete("a")).To(ConsistOf(
							flags.Completion{Item: "abc"},
							flags.Completion{Item: "abd"},
							flags.Completion{Item: "add/"},
							flags.Completion{Item: "aee/"},
						))
					})
				})

				Context("when there are no matching paths", func() {
					It("returns no matches", func() {
						Expect(path.Complete("z")).To(BeEmpty())
					})
				})
			})
		})
	})

	Describe("PathWithExistenceCheck", func() {
		var pathWithExistenceCheck PathWithExistenceCheck

		BeforeEach(func() {
			pathWithExistenceCheck = PathWithExistenceCheck("")
		})

		// The Complete method is not tested because it shares the same code as
		// Path.Complete().

		Describe("UnmarshalFlag", func() {
			Context("when the path does not exist", func() {
				It("returns a path does not exist error", func() {
					err := pathWithExistenceCheck.UnmarshalFlag("./some-dir/some-file")
					Expect(err).To(MatchError(&flags.Error{
						Type:    flags.ErrRequired,
						Message: "The specified path './some-dir/some-file' does not exist.",
					}))
				})
			})

			Context("when the path exists", func() {
				It("sets the path", func() {
					err := pathWithExistenceCheck.UnmarshalFlag("abc")
					Expect(err).ToNot(HaveOccurred())
					Expect(pathWithExistenceCheck).To(BeEquivalentTo("abc"))
				})
			})
		})
	})

	Describe("PathWithAt", func() {
		var pathWithAt PathWithAt

		BeforeEach(func() {
			pathWithAt = PathWithAt("")
		})

		Describe("Complete", func() {
			Context("when the prefix is empty", func() {
				It("returns no completions", func() {
					Expect(pathWithAt.Complete("")).To(BeEmpty())
				})
			})

			Context("when the prefix doesn't start with @", func() {
				It("returns no completions", func() {
					Expect(pathWithAt.Complete("a@b")).To(BeEmpty())
				})
			})

			Context("when the prefix starts with @", func() {
				Context("when there are no characters after the @", func() {
					It("returns all files and directories", func() {
						Expect(pathWithAt.Complete("@")).To(ConsistOf(
							flags.Completion{Item: "@abc"},
							flags.Completion{Item: "@abd"},
							flags.Completion{Item: "@add/"},
							flags.Completion{Item: "@aee/"},
							flags.Completion{Item: "@efg"},
						))
					})
				})

				Context("when there are characters after the @", func() {
					Context("when there are matching paths", func() {
						It("returns the matching paths", func() {
							Expect(pathWithAt.Complete("@a")).To(ConsistOf(
								flags.Completion{Item: "@abc"},
								flags.Completion{Item: "@abd"},
								flags.Completion{Item: "@add/"},
								flags.Completion{Item: "@aee/"},
							))
						})
					})

					Context("when there are no matching paths", func() {
						It("returns no matches", func() {
							Expect(pathWithAt.Complete("@z")).To(BeEmpty())
						})
					})
				})
			})
		})
	})
})
