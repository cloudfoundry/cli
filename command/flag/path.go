package flag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (_ Path) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
}

type PathWithExistenceCheck string

func (_ PathWithExistenceCheck) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
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

type PathWithExistenceCheckOrURL string

func (_ PathWithExistenceCheckOrURL) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
}

func (p *PathWithExistenceCheckOrURL) UnmarshalFlag(path string) error {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
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
	}

	*p = PathWithExistenceCheckOrURL(path)
	return nil
}

type PathWithAt string

func (_ PathWithAt) Complete(prefix string) []flags.Completion {
	if prefix == "" || prefix[0] != '@' {
		return nil
	}

	prefix = prefix[1:]

	var homeDir string
	if strings.HasPrefix(prefix, fmt.Sprintf("~%c", os.PathSeparator)) {
		// when $HOME is empty this will complete on /, however this is not tested
		homeDir = os.Getenv("HOME")
		prefix = fmt.Sprintf("%s%s", homeDir, prefix[1:])
	}

	return findMatches(
		fmt.Sprintf("%s*", prefix),
		func(path string) string {
			if homeDir != "" {
				newPath, err := filepath.Rel(homeDir, path)
				if err == nil {
					path = filepath.Join("~", newPath)
				}
			}
			return fmt.Sprintf("@%s", path)
		})
}

type PathWithBool string

func (_ PathWithBool) Complete(prefix string) []flags.Completion {
	return append(
		completions([]string{"true", "false"}, prefix, false),
		completeWithTilde(prefix)...,
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
			formattedMatch = fmt.Sprintf("%s%c", formattedMatch, os.PathSeparator)
		}
		matches = append(matches, flags.Completion{Item: formattedMatch})
	}

	return matches
}

func completeWithTilde(prefix string) []flags.Completion {
	var homeDir string
	if strings.HasPrefix(prefix, fmt.Sprintf("~%c", os.PathSeparator)) {
		// when $HOME is empty this will complete on /, however this is not tested
		homeDir = os.Getenv("HOME")
		prefix = fmt.Sprintf("%s%s", homeDir, prefix[1:])
	}

	return findMatches(
		fmt.Sprintf("%s*", prefix),
		func(path string) string {
			if homeDir != "" {
				newPath, err := filepath.Rel(homeDir, path)
				if err == nil {
					path = filepath.Join("~", newPath)
				}
			}
			return path
		})
}
