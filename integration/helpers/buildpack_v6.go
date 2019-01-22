// +build !V7

package helpers

import (
	"fmt"
	"regexp"
)

func BuildpacksOutputRegex(fields BuildpackFields) string {
	anyStringRegex := `\S+`
	optionalStringRegex := `\S*`
	anyBoolRegex := `(true|false)`
	anyIntRegex := `\d+`

	nameRegex := anyStringRegex
	if fields.Name != "" {
		nameRegex = regexp.QuoteMeta(fields.Name)
	}

	positionRegex := anyIntRegex
	if fields.Position != "" {
		positionRegex = regexp.QuoteMeta(fields.Position)
	}

	enabledRegex := anyBoolRegex
	if fields.Enabled != "" {
		enabledRegex = regexp.QuoteMeta(fields.Enabled)
	}

	lockedRegex := anyBoolRegex
	if fields.Locked != "" {
		lockedRegex = regexp.QuoteMeta(fields.Locked)
	}

	filenameRegex := anyStringRegex
	if fields.Filename != "" {
		filenameRegex = regexp.QuoteMeta(fields.Filename)
	}

	stackRegex := optionalStringRegex
	if fields.Stack != "" {
		stackRegex = regexp.QuoteMeta(fields.Stack)
	}

	return fmt.Sprintf(
		`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
		nameRegex,
		positionRegex,
		enabledRegex,
		lockedRegex,
		filenameRegex,
		stackRegex,
	)
}
