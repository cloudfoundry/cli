// +build V7

package rpc

import "code.cloudfoundry.org/cli/util/ui"

//go:generate counterfeiter . CommandParser

type CommandParser interface {
	ParseCommandFromArgs(ui *ui.UI, args []string) (int, error)
}
