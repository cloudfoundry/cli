package actionerror

import "fmt"

// ProcessInstanceNotFoundError is returned when the proccess type or process instance cannot be found
type ProcessInstanceNotFoundError struct {
	ProcessType   string
	InstanceIndex uint
}

func (e ProcessInstanceNotFoundError) Error() string {
	return fmt.Sprintf("Instance %d of process %s not found", e.InstanceIndex, e.ProcessType)
}
