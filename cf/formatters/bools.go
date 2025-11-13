package formatters

import (
	. "code.cloudfoundry.org/cli/v9/cf/i18n"
)

func Allowed(allowed bool) string {
	if allowed {
		return T("allowed")
	}
	return T("disallowed")
}
