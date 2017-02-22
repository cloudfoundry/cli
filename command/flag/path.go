package flag

import (
	"fmt"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
)

type Path string

func (_ Path) Complete(prefix string) []flags.Completion {
	return completeWithNoFormatting(prefix)
}

type PathWithExistenceCheck string

func (_ PathWithExistenceCheck) Complete(prefix string) []flags.Completion {
	return completeWithNoFormatting(prefix)
}

func (p *PathWithExistenceCheck) UnmarshalFlag(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &flags.Error{
				Type:    flags.ErrRequired,
				Message: fmt.Sprintf("The specified path '%s' does not exist.", path),
			}
		}
		return err
	}

	*p = PathWithExistenceCheck(path)
	return nil
}

type PathWithAt string

func (_ PathWithAt) Complete(prefix string) []flags.Completion {
	if len(prefix) > 0 && prefix[0] == '@' {
		return findMatches(
			fmt.Sprintf("%s*", prefix[1:]),
			func(path string) string {
				return fmt.Sprintf("@%s", path)
			})
	}

	return nil
}

func findMatches(pattern string, formatMatch func(string) string) []flags.Completion {
	paths, _ := filepath.Glob(pattern)
	if paths == nil {
		return nil
	}

	matches := make([]flags.Completion, len(paths))
	for i, path := range paths {
		matches[i].Item = formatMatch(path)
	}

	return matches
}

func completeWithNoFormatting(prefix string) []flags.Completion {
	return findMatches(
		fmt.Sprintf("%s*", prefix),
		func(path string) string {
			return path
		})
}
