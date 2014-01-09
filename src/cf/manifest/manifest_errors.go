package manifest

import (
	"fmt"
)

type ManifestErrors []error

func (errs ManifestErrors) Empty() bool {
	return len(errs) == 0
}

func (errs ManifestErrors) Error() (errorMessage string) {
	for _, err := range errs {
		errorMessage = fmt.Sprintf("%s%s\n", errorMessage, err)
	}
	return
}

func (errs ManifestErrors) String() string {
	return errs.Error()
}
