package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Task represents a Cloud Controller V3 Task.
type Task struct {
	// Command represents the command that will be executed. May be excluded
	// based on the user's role.
	Command string `json:"command"`
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
}

// CreateApplicationTask runs a command in the Application environment
// associated with the provided Application GUID.
func (client *Client) CreateApplicationTask(appGUID string, task Task) (Task, Warnings, error) {
	bodyBytes, err := json.Marshal(task)
	if err != nil {
		return Task{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationTasksRequest,
		URIParams: internal.Params{
			"app_guid": appGUID,
		},
		Body: bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Task{}, nil, err
	}

	var responseTask Task
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseTask,
	}

	err = client.connection.Make(request, &response)
	return responseTask, response.Warnings, err
}

// GetApplicationTasks returns a list of tasks associated with the provided
// application GUID. Results can be filtered by providing URL queries.
func (client *Client) GetApplicationTasks(appGUID string, query ...Query) ([]Task, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationTasksRequest,
		URIParams: internal.Params{
			"app_guid": appGUID,
		},
		Query: query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullTasksList []Task
	warnings, err := client.paginate(request, Task{}, func(item interface{}) error {
		if task, ok := item.(Task); ok {
			fullTasksList = append(fullTasksList, task)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Task{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullTasksList, warnings, err
}

// UpdateTaskCancel cancels a task.
func (client *Client) UpdateTaskCancel(taskGUID string) (Task, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutTaskCancelRequest,
		URIParams: internal.Params{
			"task_guid": taskGUID,
		},
	})
	if err != nil {
		return Task{}, nil, err
	}

	var task Task
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &task,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Task{}, response.Warnings, err
	}

	return task, response.Warnings, nil
}
