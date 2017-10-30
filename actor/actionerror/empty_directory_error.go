package actionerror

import "fmt"

type EmptyDirectoryError struct {
	Path string
}

func (e EmptyDirectoryError) Error() string {
	return fmt.Sprint(e.Path, "is empty")
}
