package manifest

import (
	"fmt"
	"strings"
)

type GlobalFieldsError struct {
	Fields []string
}

func (e GlobalFieldsError) Error() string {
	return fmt.Sprintf("unsupported global fields: %s", strings.Join(e.Fields, ", "))
}
