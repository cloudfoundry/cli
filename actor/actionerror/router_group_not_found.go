package actionerror

import "fmt"

type RouterGroupNotFoundError struct {
	Name string
}

func (err RouterGroupNotFoundError) Error() string {
	return fmt.Sprintf("Router group %s not found", err.Name)
}
