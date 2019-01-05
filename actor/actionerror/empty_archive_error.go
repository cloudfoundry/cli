package actionerror

import "fmt"

type EmptyArchiveError struct {
	Path string
}

func (e EmptyArchiveError) Error() string {
	return fmt.Sprint(e.Path, "is an empty archive")
}
