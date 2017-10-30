package actionerror

import "fmt"

// ProcessInstanceNotRunningError is returned when trying to perform an action
// on an instance that is not running
type ProcessInstanceNotRunningError struct {
	ProcessType   string
	InstanceIndex uint
}

func (e ProcessInstanceNotRunningError) Error() string {
	return fmt.Sprintf("Instance %d of process %s not running", e.InstanceIndex, e.ProcessType)
}
