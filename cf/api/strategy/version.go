package strategy

import (
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
)

type Version struct {
	Major int64
	Minor int64
	Patch int64
}

func ParseVersion(input string) (Version, error) {
	parts := strings.Split(input, ".")
	if len(parts) != 3 {
		return Version{}, errors.NewWithFmt(T("Could not parse version number: {{.Input}}",
			map[string]interface{}{"Input": input}))
	}

	major, err1 := strconv.ParseInt(parts[0], 10, 64)
	minor, err2 := strconv.ParseInt(parts[1], 10, 64)
	patch, err3 := strconv.ParseInt(parts[2], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return Version{}, errors.NewWithFmt(T("Could not parse version number: {{.Input}}",
			map[string]interface{}{"Input": input}))
	}

	return Version{major, minor, patch}, nil
}

func (version Version) LessThan(other Version) bool {
	if version.Major < other.Major {
		return true
	}

	if version.Major > other.Major {
		return false
	}

	if version.Minor < other.Minor {
		return true
	}

	if version.Minor > other.Minor {
		return false
	}

	if version.Patch < other.Patch {
		return true
	}

	return false
}

func (version Version) GreaterThanOrEqualTo(other Version) bool {
	return !version.LessThan(other)
}
