package fmt

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const (
	Indent = "  "
	Bullet = "- "
)

func MultilineError(err error) string {
	return prefixingMultilineError(err, "", "")
}

func prefixingMultilineError(err error, prefix string, bullet string) string {
	currPrefix := prefix + bullet
	prefix = prefix + strings.Repeat(" ", len(bullet))

	switch specificErr := err.(type) {
	case bosherr.ComplexError:
		return currPrefix + specificErr.Err.Error() + ":\n" + prefixingMultilineError(specificErr.Cause, prefix+Indent, "")
	case bosherr.MultiError:
		lines := make([]string, len(specificErr.Errors), len(specificErr.Errors))
		for i, sibling := range specificErr.Errors {
			lines[i] = prefixingMultilineError(sibling, prefix, Bullet)
		}
		return strings.Join(lines, "\n")
	case boshsys.ExecError:
		lines := []string{
			"Error Executing Command:",
			prefixEachLine(specificErr.Command, Indent),
			"StdOut:",
			prefixEachLine(specificErr.StdOut, Indent),
			"StdErr:",
			prefixEachLine(specificErr.StdErr, Indent),
		}
		return prefixEachLine(strings.Join(lines, "\n"), prefix)
	default:
		return currPrefix + specificErr.Error()
	}
}

func prefixEachLine(str string, prefix string) string {
	return prefix + strings.Replace(str, "\n", "\n"+prefix, -1)
}
