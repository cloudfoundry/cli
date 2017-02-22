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
	if prefix == "" || prefix[0] != '@' {
		return nil
	}

	return findMatches(
		fmt.Sprintf("%s*", prefix[1:]),
		func(path string) string {
			return fmt.Sprintf("@%s", path)
		})
}

type PathWithBool string

func (_ PathWithBool) Complete(prefix string) []flags.Completion {
	return append(
		completions([]string{"true", "false"}, prefix),
		completeWithNoFormatting(prefix)...,
	)
}

func findMatches(pattern string, formatMatch func(string) string) []flags.Completion {
	paths, _ := filepath.Glob(pattern)
	if paths == nil {
		return nil
	}

	matches := []flags.Completion{}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		formattedMatch := formatMatch(path)
		if info.IsDir() {
			matches = append(matches, flags.Completion{Item: fmt.Sprintf("%s/", formattedMatch)})
		} else {
			matches = append(matches, flags.Completion{Item: formattedMatch})
		}
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
