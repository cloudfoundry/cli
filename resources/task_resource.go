package resources

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
)

// Task represents a Cloud Controller V3 Task.
type Task struct {
	// Command represents the command that will be executed. May be excluded
	// based on the user's role.
	Command string `json:"command,omitempty"`
	// CreatedAt represents the time with zone when the object was created.
	CreatedAt string `json:"created_at,omitempty"`
	// DiskInMB represents the disk in MB allocated for the task.
	DiskInMB uint64 `json:"disk_in_mb,omitempty"`
	// GUID represents the unique task identifier.
	GUID string `json:"guid,omitempty"`
	// LogRateLimitInBPS represents the log rate limit in bytes allocated for the task.
	LogRateLimitInBPS int `json:"log_rate_limit_in_bytes_per_second,omitempty"`
	// MemoryInMB represents the memory in MB allocated for the task.
	MemoryInMB uint64 `json:"memory_in_mb,omitempty"`
	// Name represents the name of the task.
	Name string `json:"name,omitempty"`
	// SequenceID represents the user-facing id of the task. This number is
	// unique for every task associated with a given app.
	SequenceID int64 `json:"sequence_id,omitempty"`
	// State represents the task state.
	State constant.TaskState `json:"state,omitempty"`
	// Tasks can use a process as a template to fill in
	// command, memory, disk values
	//
	// Using a pointer so that it can be set to nil to prevent
	// json serialization when no template is used
	Template *TaskTemplate `json:"template,omitempty"`

	// Result contains the task result
	Result *TaskResult `json:"result,omitempty"`
}

type TaskTemplate struct {
	Process TaskProcessTemplate `json:"process,omitempty"`
}

type TaskProcessTemplate struct {
	Guid string `json:"guid,omitempty"`
}

type TaskResult struct {
	FailureReason string `json:"failure_reason,omitempty"`
}
