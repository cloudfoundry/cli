package actionerror

import "fmt"

type FileChangedError struct {
	Filename string
}

func (e FileChangedError) Error() string {
	return fmt.Sprint("SHA1 mismatch for:", e.Filename)
}
