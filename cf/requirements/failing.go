package requirements

import "errors"

type Failing struct {
	Message string
}

func (r Failing) Execute() error {
	return errors.New(r.Message)
}
