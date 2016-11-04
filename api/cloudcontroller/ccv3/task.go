package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// Task represents a Cloud Controller V3 Task.
type Task struct {
	SequenceID int `json:"sequence_id"`
}

// RunTaskBody represents the body of the request to create a Task.
type RunTaskBody struct {
	Command string `json:"command"`
}

// RunTask runs a command in the Application environment associated with the
// provided Application GUID.
func (client *Client) RunTask(appGUID string, command string) (Task, Warnings, error) {
	bodyBytes, err := json.Marshal(RunTaskBody{Command: command})
	if err != nil {
		return Task{}, nil, err
	}

	body := bytes.NewBuffer(bodyBytes)
	request, err := client.newHTTPRequest(requestOptions{
		URL:    fmt.Sprintf("%s/v3/apps/%s/tasks", client.cloudControllerURL, appGUID),
		Method: http.MethodPost,
		Body:   body,
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
