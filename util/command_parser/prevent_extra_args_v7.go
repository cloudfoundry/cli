// +build V7

package command_parser

import (
	"strings"

	"code.cloudfoundry.org/cli/command/translatableerror"
)

func preventExtraArgs(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: strings.Join(args, " "),
		}
	}
	return nil
}
