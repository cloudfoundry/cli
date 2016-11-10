package v3actions

import (
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Task represents a V3 actor Task.
type Task ccv3.Task

// RunTaskError represents a generic error encountered when creating a task.
type RunTaskError struct {
	Message string
}

func (e RunTaskError) Error() string {
	splitMessage := strings.Split(e.Message, ": ")
	if len(splitMessage) > 1 {
		return splitMessage[1]
	}
	return e.Message
}

// TaskWorkersUnavailableError is returned when there are no workers to run a
// given task.
type TaskWorkersUnavailableError struct {
	Message string
}

func (e TaskWorkersUnavailableError) Error() string {
	return e.Message
}

// RunTask runs the provided command in the application environment associated
// with the provided application GUID.
func (actor Actor) RunTask(appGUID string, command string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.RunTask(appGUID, command)
	switch err.(type) {
	case cloudcontroller.UnprocessableEntityError:
		return Task{}, Warnings(warnings), RunTaskError{Message: err.Error()}
	case cloudcontroller.TaskWorkersUnavailableError:
		return Task{}, Warnings(warnings), TaskWorkersUnavailableError{Message: err.Error()}
	}

	return Task(task), Warnings(warnings), err
}

// GetApplicationTasks returns a list of tasks associated with the provided
// appplication GUID.
func (actor Actor) GetApplicationTasks(appGUID string, sortOrder SortOrder) ([]Task, Warnings, error) {
	query := url.Values{}
	if sortOrder == Descending {
		query.Add("order_by", "-created_at")
	}

	tasks, warnings, err := actor.CloudControllerClient.GetApplicationTasks(appGUID, query)
	actorWarnings := Warnings(warnings)
	if err != nil {
		return nil, actorWarnings, err
	}

	allTasks := []Task{}
	for _, task := range tasks {
		allTasks = append(allTasks, Task(task))
	}

	return allTasks, actorWarnings, nil
}
