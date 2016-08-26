// this is a workaround while the i18n4go tool isn't fixed to handle our
// indirect use case we should delete this once that is fixed
// https://github.com/maximilien/i18n4go/issues/45
package cf

import (
	. "code.cloudfoundry.org/cli/cf/i18n"
)

func _() {
	T("SEE ALSO:")
	T("APPS:")
	T("ENVIRONMENT VARIABLE GROUPS:")
	T("ADVANCED:")
	T("SPACE ADMIN:")
	T("ROUTES:")
	T("INSTALLED PLUGIN COMMANDS:")
	T("ADD/REMOVE PLUGIN:")
	T("SERVICES:")
	T("BUILDPACKS:")
	T("SERVICE ADMIN:")
	T("GETTING STARTED:")
	T("SPACES:")
	T("DOMAINS:")
	T("ADD/REMOVE PLUGIN REPOSITORY:")
	T("FEATURE FLAGS:")
	T("USER ADMIN:")
	T("ORGS:")
	T("SECURITY GROUP:")
	T("ORG ADMIN:")
}
