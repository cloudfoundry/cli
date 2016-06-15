package requirements

import (
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type NumberArguments struct {
	passedArgs   []string
	expectedArgs []string
	commandUsage string
}

func NewNumberArguments(passedArgs []string, expectedArgs []string, commandUsage string) Requirement {
	return NumberArguments{
		passedArgs:   passedArgs,
		expectedArgs: expectedArgs,
		commandUsage: commandUsage,
	}
}

func (r NumberArguments) Execute() error {
	if len(r.passedArgs) != len(r.expectedArgs) {
		return NumberArgumentsError{ExpectedArgs: r.expectedArgs, CommandUsage: r.commandUsage}
	}

	return nil
}

type NumberArgumentsError struct {
	ExpectedArgs []string
	CommandUsage string
}

func (e NumberArgumentsError) Error() string {
	return T("Incorrect Usage. Requires {{.Arguments}} as arguments\n\n{{.CommandUsage}}", map[string]string{
		"Arguments": strings.Join(e.ExpectedArgs, ", "),
	})
}
