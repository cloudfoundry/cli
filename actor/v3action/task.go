package v3action

import (
	"fmt"
	"net/url"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Task represents a V3 actor Task.
type Task ccv3.Task

// TaskWorkersUnavailableError is returned when there are no workers to run a
// given task.
type TaskWorkersUnavailableError struct {
	Message string
}

func (e TaskWorkersUnavailableError) Error() string {
	return e.Message
}

// TaskNotFoundError is returned when no tasks matching the filters are found.
type TaskNotFoundError struct {
	SequenceID int
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("Task sequence ID %d not found.", e.SequenceID)
}

// RunTask runs the provided command in the application environment associated
// with the provided application GUID.
func (actor Actor) RunTask(appGUID string, command string, name string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.NewTask(appGUID, command, name)
	if err != nil {
		if e, ok := err.(cloudcontroller.TaskWorkersUnavailableError); ok {
			return Task{}, Warnings(warnings), TaskWorkersUnavailableError{Message: e.Error()}
		}
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

func (actor Actor) GetTaskBySequenceIDAndApplication(sequenceID int, appGUID string) (Task, Warnings, error) {
	query := url.Values{
		"sequence_ids": []string{strconv.Itoa(sequenceID)},
	}

	tasks, warnings, err := actor.CloudControllerClient.GetApplicationTasks(appGUID, query)
	if err != nil {
		return Task{}, Warnings(warnings), err
	}

	if len(tasks) == 0 {
		return Task{}, Warnings(warnings), TaskNotFoundError{SequenceID: sequenceID}
	}

	return Task(tasks[0]), Warnings(warnings), nil
}

func (actor Actor) TerminateTask(taskGUID string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.UpdateTask(taskGUID)
	return Task(task), Warnings(warnings), err
}
