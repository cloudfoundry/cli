package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Task represents a Cloud Controller V3 Task.
type Task struct {
	GUID       string `json:"guid"`
	SequenceID int    `json:"sequence_id"`
	Name       string `json:"name"`
	Command    string `json:"command"`
	State      string `json:"state"`
	CreatedAt  string `json:"created_at"`
}

// NewTaskBody represents the body of the request to create a Task.
type NewTaskBody struct {
	Command string `json:"command"`
	Name    string `json:"name,omitempty"`
}

// NewTask runs a command in the Application environment associated with the
// provided Application GUID.
func (client *Client) NewTask(appGUID string, command string, name string) (Task, Warnings, error) {
	bodyBytes, err := json.Marshal(NewTaskBody{
		Command: command,
		Name:    name,
	})
	if err != nil {
		return Task{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.NewAppTaskRequest,
		URIParams: internal.Params{
			"guid": appGUID,
		},
		Body: bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return Task{}, nil, err
	}

	var task Task
	response := cloudcontroller.Response{
		Result: &task,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Task{}, response.Warnings, err
	}

	return task, response.Warnings, nil
}

// GetApplicationTasks returns a list of tasks associated with the provided
// application GUID. Results can be filtered by providing URL queries.
func (client *Client) GetApplicationTasks(appGUID string, query url.Values) ([]Task, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppTasksRequest,
		URIParams: internal.Params{
			"guid": appGUID,
		},
		Query: query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullTasksList []Task
	warnings, err := client.paginate(request, Task{}, func(item interface{}) error {
		if app, ok := item.(Task); ok {
			fullTasksList = append(fullTasksList, app)
		} else {
			return cloudcontroller.UnknownObjectInListError{
				Expected:   Task{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullTasksList, warnings, err
}

// UpdateTask cancels a task.
func (client *Client) UpdateTask(taskGUID string) (Task, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URL:    fmt.Sprintf("%s/v3/tasks/%s/cancel", client.cloudControllerURL, taskGUID),
		Method: http.MethodPut,
	})
	if err != nil {
		return Task{}, nil, err
	}

	var task Task
	response := cloudcontroller.Response{
		Result: &task,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Task{}, response.Warnings, err
	}

	return task, response.Warnings, nil
}
