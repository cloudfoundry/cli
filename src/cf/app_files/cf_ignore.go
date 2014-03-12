package app_files

import (
	"glob"
	"strings"
	"path"
)


type CfIgnore interface {
	FileShouldBeIgnored(path string) bool
}

func NewCfIgnore(text string) (CfIgnore) {
	patterns := []ignorePattern{}
	inclusions := []glob.Glob{}
	exclusions := []glob.Glob{}

	lines := strings.Split(text, "\n")
	lines = append(DefaultIgnoreFiles, lines...)

	for _, pattern := range lines {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		if strings.HasPrefix(pattern, "!") {
			pattern := pattern[1:]
			pattern = path.Clean(pattern)
			inclusions = append(inclusions, globsForPattern(pattern)...)
		} else {
			pattern = path.Clean(pattern)
			exclusions = append(exclusions, globsForPattern(pattern)...)
		}
	}


	for _, glob := range exclusions {
		patterns = append(patterns, ignorePattern{true, glob})
	}
	for _, glob := range inclusions {
		patterns = append(patterns, ignorePattern{false, glob})
	}

	return cfIgnore(patterns)
}

func (ignore cfIgnore) FileShouldBeIgnored(path string) bool {
	result := false

	for _, pattern := range ignore {
		if strings.HasPrefix(pattern.glob.String(), "/") && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if pattern.glob.Match(path) {
			result = pattern.exclude
		}
	}

	return result
}

func globsForPattern(pattern string) (globs []glob.Glob) {
	globs = append(globs, glob.MustCompileGlob(pattern))
	globs = append(globs, glob.MustCompileGlob(path.Join(pattern, "*")))
	globs = append(globs, glob.MustCompileGlob(path.Join(pattern, "**", "*")))

	if !strings.HasPrefix(pattern, "/") {
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern)))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern, "*")))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern, "**", "*")))
	}

	return
}

type ignorePattern struct {
	exclude bool
	glob glob.Glob
}

type cfIgnore []ignorePattern
