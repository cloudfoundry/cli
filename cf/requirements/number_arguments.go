package requirements

import (
	"strings"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type NumberArguments struct {
	passedArgs   []string
	expectedArgs []string
}

func NewNumberArguments(passedArgs []string, expectedArgs []string) Requirement {
	return NumberArguments{
		passedArgs:   passedArgs,
		expectedArgs: expectedArgs,
	}
}

func (r NumberArguments) Execute() error {
	if len(r.passedArgs) != len(r.expectedArgs) {
		return NumberArgumentsError{ExpectedArgs: r.expectedArgs}
	}

	return nil
}

type NumberArgumentsError struct {
	ExpectedArgs []string
}

func (e NumberArgumentsError) Error() string {
	return T("Incorrect Usage. Requires {{.Arguments}}", map[string]string{
		"Arguments": strings.Join(e.ExpectedArgs, ", "),
	})
}
