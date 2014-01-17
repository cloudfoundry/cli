package glob

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestGlobTranslateOk(t *testing.T) {
	for _, parts := range globs {
		pat, exp := parts[0], parts[1]
		got, err := translateGlob(pat)

		assert.NoError(t, err)
		assert.Equal(t, exp, got, "expected %q, but got %q from %q", exp, got, pat)
	}
}

func TestGlobMatches(t *testing.T) {
	for _, parts := range matches {
		pat, paths := parts[0], parts[1:]
		glob, err := CompileGlob(pat)

		assert.NoError(t, err)
		for _, path := range paths {
			assert.True(t, glob.Match(path), "path %q should match %q", pat, path)
		}
	}
}

func TestGlobNonMatches(t *testing.T) {
	for _, parts := range nonMatches {
		pat, paths := parts[0], parts[1:]
		glob, err := CompileGlob(pat)

		assert.NoError(t, err)
		for _, path := range paths {
			assert.False(t, glob.Match(path), "path %q should match %q", pat, path)
		}
	}
}
