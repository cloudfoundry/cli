package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
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
}

type TaskTemplate struct {
	Process TaskProcessTemplate `json:"process,omitempty"`
}

type TaskProcessTemplate struct {
	Guid string `json:"guid,omitempty"`
}

// CreateApplicationTask runs a command in the Application environment
// associated with the provided Application GUID.
func (client *Client) CreateApplicationTask(appGUID string, task Task) (Task, Warnings, error) {
	var responseBody Task

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationTasksRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		RequestBody:  task,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetApplicationTasks returns a list of tasks associated with the provided
// application GUID. Results can be filtered by providing URL queries.
func (client *Client) GetApplicationTasks(appGUID string, query ...Query) ([]Task, Warnings, error) {
	var resources []Task

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetApplicationTasksRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		Query:        query,
		ResponseBody: Task{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Task))
			return nil
		},
	})

	return resources, warnings, err
}

// UpdateTaskCancel cancels a task.
func (client *Client) UpdateTaskCancel(taskGUID string) (Task, Warnings, error) {
	var responseBody Task

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PutTaskCancelRequest,
		URIParams: internal.Params{
			"task_guid": taskGUID,
		},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
