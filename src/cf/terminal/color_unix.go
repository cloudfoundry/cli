// +build darwin freebsd linux netbsd openbsd

package terminal

import (
	"code.google.com/p/go.crypto/ssh/terminal"
)

func isTerminal() bool {
	return terminal.IsTerminal(1)
}
