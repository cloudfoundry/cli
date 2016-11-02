package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

type Task struct {
	SequenceID int `json:"sequence_id"`
}

type RunTaskBody struct {
	Command string `json:"command"`
}

func (client *CloudControllerClient) RunTask(appGUID string, command string) (Task, Warnings, error) {

	bodyBytes, err := json.Marshal(RunTaskBody{Command: command})
	if err != nil {
		return Task{}, nil, err
	}

	body := bytes.NewBuffer(bodyBytes)
	request, err := client.newHTTPRequest(requestOptions{
		URI:    fmt.Sprintf("/v3/apps/%s/tasks", appGUID),
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
