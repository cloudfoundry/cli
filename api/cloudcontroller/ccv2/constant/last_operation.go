package constant

type LastOperationState string

const (
	LastOperationInProgress LastOperationState = "in progress"
	LastOperationSucceeded  LastOperationState = "succeeded"
	LastOperationFailed     LastOperationState = "failed"
)
