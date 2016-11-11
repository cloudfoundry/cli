package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
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
}

// NewTask runs a command in the Application environment associated with the
// provided Application GUID.
func (client *Client) NewTask(appGUID string, command string) (Task, Warnings, error) {
	bodyBytes, err := json.Marshal(NewTaskBody{Command: command})
	if err != nil {
		return Task{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		URL:    fmt.Sprintf("%s/v3/apps/%s/tasks", client.cloudControllerURL, appGUID),
		Method: http.MethodPost,
		Body:   bytes.NewBuffer(bodyBytes),
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
		URL:    fmt.Sprintf("%s/v3/apps/%s/tasks", client.cloudControllerURL, appGUID),
		Method: http.MethodGet,
		Query:  query,
	})
	if err != nil {
		return nil, nil, err
	}

	allTasks := []Task{}
	allWarnings := Warnings{}

	for {
		var tasks []Task
		wrapper := PaginatedWrapper{
			Resources: &tasks,
		}
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err = client.connection.Make(request, &response)
		allWarnings = append(allWarnings, response.Warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		allTasks = append(allTasks, tasks...)

		if wrapper.Pagination.Next.HREF == "" {
			break
		}

		request, err = client.newHTTPRequest(requestOptions{
			URL:    wrapper.Pagination.Next.HREF,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, allWarnings, err
		}
	}

	return allTasks, allWarnings, nil
}
