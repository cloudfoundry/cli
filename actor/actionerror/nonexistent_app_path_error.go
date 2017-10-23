package actionerror

import "fmt"

type NonexistentAppPathError struct {
	Path string
}

func (e NonexistentAppPathError) Error() string {
	return fmt.Sprint("app path not found:", e.Path)
}
