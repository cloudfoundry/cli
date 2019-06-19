package matchers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

type CommandInCategoryMatcher struct {
	command     string
	category    string
	description string

	actualDescription string
	actualCategory    string
}

func HaveCommandInCategoryWithDescription(command, category, description string) types.GomegaMatcher {
	return &CommandInCategoryMatcher{
		command:     command,
		category:    category,
		description: description,
	}
}

func getDescriptionAndCategory(out, command string) (string, string) {
	lines := strings.Split(out, "\n")
	description := ""
	category := ""
	index := -1

	for i, line := range lines {
		r := regexp.MustCompile(fmt.Sprintf(`^\s+%s\s+(.*)$`, command))
		matches := r.FindAllStringSubmatch(line, 1)
		if matches != nil {
			description = matches[0][1]
			index = i
			break
		}
	}

	for i := index - 1; i >= 0; i-- {
		r := regexp.MustCompile(`^(.+):$`)
		matches := r.FindAllStringSubmatch(lines[i], 1)
		if matches != nil {
			category = matches[0][1]
			break
		}
	}

	return description, category
}

func (matcher *CommandInCategoryMatcher) Match(actual interface{}) (success bool, err error) {
	var session *gexec.Session
	var out string
	var ok bool
	if session, ok = actual.(*gexec.Session); !ok {
		return false, errors.New("HaveCommandInCategory: Actual value must be a *gexec.Session.")
	}
	out = string(session.Out.Contents())

	matcher.actualDescription, matcher.actualCategory = getDescriptionAndCategory(out, matcher.command)
	if matcher.actualDescription == "" && matcher.actualCategory == "" {
		return false, fmt.Errorf("HaveCommandInCategory: output:\n%s\ndoes not contain command `%s`\n", out, matcher.command)
	}

	return matcher.actualDescription == matcher.description && matcher.actualCategory == matcher.category, nil
}

func (matcher *CommandInCategoryMatcher) FailureMessage(actual interface{}) string {
	if matcher.actualDescription != matcher.description {
		return fmt.Sprintf("Expected command `%s` to have description:\n %s\nbut found:\n %s\n", matcher.command, matcher.description, matcher.actualDescription)
	}

	return fmt.Sprintf("Expected command `%s` to be in category `%s` but found it in `%s`\n", matcher.command, matcher.category, matcher.actualCategory)
}

func (matcher *CommandInCategoryMatcher) NegatedFailureMessage(actual interface{}) string {
	panic("Not implemented. Are you sure you want to negate this test?")
	return "Not implemented. Are you sure you want to negate this test?"
}
