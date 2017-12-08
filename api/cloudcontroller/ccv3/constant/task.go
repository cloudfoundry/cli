package constant

// TaskState represents the state of the task
type TaskState string

const (
	// TaskPending is when the task is pending.
	TaskPending TaskState = "PENDING"
	// TaskRunning is when the task is running.
	TaskRunning TaskState = "RUNNING"
	// TaskSucceeded is when the task succeeded.
	TaskSucceeded TaskState = "SUCCEEDED"
	// TaskCanceling is when the task is canceling.
	TaskCanceling TaskState = "CANCELING"
	// TaskFailed is when the task Failed.
	TaskFailed TaskState = "FAILED"
)
