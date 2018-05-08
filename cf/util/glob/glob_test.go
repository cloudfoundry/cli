package glob

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var globs = [][]string{
	{"/", `^/$`},
	{"/a", `^/a$`},
	{"/a.b", `^/a\.b$`},
	{"/a-b", `^/a\-b$`},
	{"/a?", `^/a[^/]$`},
	{"/a/b", `^/a/b$`},
	{"/*", `^/[^/]*$`},
	{"/*/a", `^/[^/]*/a$`},
	{"/*a/b", `^/[^/]*a/b$`},
	{"/a*/b", `^/a[^/]*/b$`},
	{"/a*a/b", `^/a[^/]*a/b$`},
	{"/*a*/b", `^/[^/]*a[^/]*/b$`},
	{"/**", `^/.*$`},
	{"/**/a", `^/.*/a$`},
}

var matches = [][]string{
	{"/a/b", "/a/b"},
	{"/a?", "/ab", "/ac"},
	{"/a*", "/a", "/ab", "/abc"},
	{"/a**", "/a", "/ab", "/abc", "/a/", "/a/b", "/ab/c"},
	{`c:\a\b\.d`, `c:\a\b\.d`},
	{`c:\**\.d`, `c:\a\b\.d`},
}

var nonMatches = [][]string{
	{"/a/b", "/a/c", "/a/", "/a/b/", "/a/bc"},
	{"/a?", "/", "/abc", "/a", "/a/"},
	{"/a*", "/", "/a/", "/ba"},
	{"/a**", "/", "/ba"},
}

var _ = Describe("Glob", func() {
	It("translates globs to regexes", func() {
		for _, parts := range globs {
			pat, exp := parts[0], parts[1]
			got, err := translateGlob(pat)

			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(Equal(got), "expected %q, but got %q from %q", exp, got, pat)
		}
	})

	It("creates regexes that match correct file paths", func() {
		for _, parts := range matches {
			pat, paths := parts[0], parts[1:]
			glob, err := CompileGlob(pat)

			Expect(err).NotTo(HaveOccurred())
			for _, path := range paths {
				Expect(glob.Match(path)).To(BeTrue(), "path %q should match %q", pat, path)
			}
		}
	})

	It("creates regexes that do not match file incorrect file paths", func() {
		for _, parts := range nonMatches {
			pat, paths := parts[0], parts[1:]
			glob, err := CompileGlob(pat)

			Expect(err).NotTo(HaveOccurred())
			for _, path := range paths {
				Expect(glob.Match(path)).To(BeFalse(), "path %q should match %q", pat, path)
			}
		}
	})
})

func TestGlobSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Glob Suite")
}
