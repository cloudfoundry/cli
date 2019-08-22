// +build !V7

package main

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"strings"
)

func preventExtraArgs(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: strings.Join(args, " "),
		}
	}
	return nil
}
