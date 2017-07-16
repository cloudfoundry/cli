package flag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (Path) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
}

type PathWithExistenceCheck string

func (PathWithExistenceCheck) Complete(prefix string) []flags.Completion {
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

type JSONOrFileWithValidation map[string]interface{}

func (JSONOrFileWithValidation) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
}

func (p *JSONOrFileWithValidation) UnmarshalFlag(pathOrJSON string) error {
	var jsonBytes []byte

	errorToReturn := &flags.Error{
		Type:    flags.ErrRequired,
		Message: "Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object.",
	}

	_, err := os.Stat(pathOrJSON)
	if err == nil {
		jsonBytes, err = ioutil.ReadFile(pathOrJSON)
		if err != nil {
			return errorToReturn
		}
	} else {
		jsonBytes = []byte(pathOrJSON)
	}

	var jsonMap map[string]interface{}

	if jsonIsInvalid := json.Unmarshal(jsonBytes, &jsonMap); jsonIsInvalid != nil {
		return errorToReturn
	}

	*p = JSONOrFileWithValidation(jsonMap)
	return nil
}

type PathWithExistenceCheckOrURL string

func (PathWithExistenceCheckOrURL) Complete(prefix string) []flags.Completion {
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

func (PathWithAt) Complete(prefix string) []flags.Completion {
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

func (PathWithBool) Complete(prefix string) []flags.Completion {
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
