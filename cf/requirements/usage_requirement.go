package requirements

import (
	"errors"
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RequirementFunction func() error

func (f RequirementFunction) Execute() error {
	return f()
}

func NewUsageRequirement(cmd usable, errorMessage string, pred func() bool) Requirement {
	return RequirementFunction(func() error {
		if pred() {
			m := fmt.Sprintf("%s. %s\n\n%s", T("Incorrect Usage"), errorMessage, cmd.Usage())

			return errors.New(m)
		}

		return nil
	})
}

type usable interface {
	Usage() string
}
