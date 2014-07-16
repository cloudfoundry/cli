package formatters

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
)

func Allowed(allowed bool) string {
	if allowed {
		return T("allowed")
	} else {
		return T("disallowed")
	}
}
