package requirements

import (
	"errors"
	"fmt"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type RequirementFunction func() error

func (f RequirementFunction) Execute() error {
	return f()
}

func NewUsageRequirement(cmd Usable, errorMessage string, pred func() bool) Requirement {
	return RequirementFunction(func() error {
		if pred() {
			m := fmt.Sprintf("%s. %s\n\n%s", T("Incorrect Usage"), errorMessage, cmd.Usage())

			return errors.New(m)
		}

		return nil
	})
}

type Usable interface {
	Usage() string
}
