package director

import (
	"fmt"
	gourl "net/url"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TaskImpl struct {
	client Client

	id             int
	startedAt      time.Time
	lastActivityAt time.Time

	state          string
	user           string
	deploymentName string

	description string
	result      string
	contextId   string
}

func (t TaskImpl) ID() int                   { return t.id }
func (t TaskImpl) ContextID() string         { return t.contextId }
func (t TaskImpl) StartedAt() time.Time      { return t.startedAt }
func (t TaskImpl) LastActivityAt() time.Time { return t.lastActivityAt }

func (t TaskImpl) State() string { return t.state }

func (t TaskImpl) IsError() bool {
	return t.state == "error" || t.state == "timeout" || t.state == "cancelled"
}

func (t TaskImpl) User() string           { return t.user }
func (t TaskImpl) DeploymentName() string { return t.deploymentName }

func (t TaskImpl) Description() string { return t.description }
func (t TaskImpl) Result() string      { return t.result }

func (t TaskImpl) Cancel() error { return t.client.CancelTask(t.id) }

type TaskResp struct {
	ID int // 165

	StartedAt      int64 `json:"started_at"` // 1440318199
	LastActivityAt int64 `json:"timestamp"`  // 1440318199

	State      string // e.g. "queued", "processing", "done", "error", "cancelled"
	User       string // e.g. "admin"
	Deployment string

	Description string // e.g. "create release"
	Result      string // e.g. "Created release `bosh-ui/0+dev.17'"
	ContextId   string `json:"context_id"`
}

func NewTaskFromResp(client Client, r TaskResp) TaskImpl {
	return TaskImpl{
		client: client,

		id: r.ID,

		startedAt:      time.Unix(r.StartedAt, 0).UTC(),
		lastActivityAt: time.Unix(r.LastActivityAt, 0).UTC(),

		state:          r.State,
		user:           r.User,
		deploymentName: r.Deployment,

		description: r.Description,
		result:      r.Result,
		contextId:   r.ContextId,
	}
}

func (d DirectorImpl) CurrentTasks(filter TasksFilter) ([]Task, error) {
	tasks := []Task{}

	taskResps, err := d.client.CurrentTasks(filter)
	if err != nil {
		return tasks, err
	}

	for _, r := range taskResps {
		tasks = append(tasks, NewTaskFromResp(d.client, r))
	}

	return tasks, nil
}

func (d DirectorImpl) RecentTasks(limit int, filter TasksFilter) ([]Task, error) {
	tasks := []Task{}

	taskResps, err := d.client.RecentTasks(limit, filter)
	if err != nil {
		return tasks, err
	}

	for _, r := range taskResps {
		tasks = append(tasks, NewTaskFromResp(d.client, r))
	}

	return tasks, nil
}

func (d DirectorImpl) FindTask(id int) (Task, error) {
	taskResp, err := d.client.Task(id)
	if err != nil {
		return TaskImpl{}, err
	}

	return NewTaskFromResp(d.client, taskResp), nil
}

func (d DirectorImpl) FindTasksByContextId(contextId string) ([]Task, error) {
	tasks := []Task{}

	taskResps, err := d.client.FindTasksByContextId(contextId)
	if err != nil {
		return tasks, err
	}

	for _, r := range taskResps {
		tasks = append(tasks, NewTaskFromResp(d.client, r))
	}

	return tasks, nil
}

func (t TaskImpl) EventOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "event", taskReporter)
}

func (t TaskImpl) CPIOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "cpi", taskReporter)
}

func (t TaskImpl) DebugOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "debug", taskReporter)
}

func (t TaskImpl) ResultOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "result", taskReporter)
}

func (c Client) CurrentTasks(filter TasksFilter) ([]TaskResp, error) {
	var tasks []TaskResp

	query := gourl.Values{}

	query.Add("state", "processing,cancelling,queued")
	query.Add("verbose", c.taskVerbosity(filter.All))

	if len(filter.Deployment) > 0 {
		query.Add("deployment", filter.Deployment)
	}

	path := fmt.Sprintf("/tasks?%s", query.Encode())

	err := c.clientRequest.Get(path, &tasks)
	if err != nil {
		return tasks, bosherr.WrapErrorf(err, "Finding current tasks")
	}

	return tasks, nil
}

func (c Client) RecentTasks(limit int, filter TasksFilter) ([]TaskResp, error) {
	var tasks []TaskResp

	query := gourl.Values{}

	query.Add("limit", fmt.Sprintf("%d", limit))
	query.Add("verbose", c.taskVerbosity(filter.All))

	if len(filter.Deployment) > 0 {
		query.Add("deployment", filter.Deployment)
	}

	path := fmt.Sprintf("/tasks?%s", query.Encode())

	err := c.clientRequest.Get(path, &tasks)
	if err != nil {
		return tasks, bosherr.WrapErrorf(err, "Finding recent tasks")
	}

	return tasks, nil
}

func (c Client) FindTasksByContextId(contextId string) ([]TaskResp, error) {
	var tasks []TaskResp

	query := gourl.Values{}

	query.Add("context_id", contextId)

	path := fmt.Sprintf("/tasks?%s", query.Encode())

	err := c.clientRequest.Get(path, &tasks)
	if err != nil {
		return tasks, bosherr.WrapErrorf(err, "Finding tasks by context_id:'%s'", contextId)
	}

	return tasks, nil
}

func (c Client) taskVerbosity(includeAll bool) string {
	if includeAll {
		return "2"
	}
	return "1"
}

func (c Client) Task(id int) (TaskResp, error) {
	var task TaskResp

	err := c.clientRequest.Get(fmt.Sprintf("/tasks/%d", id), &task)
	if err != nil {
		return task, bosherr.WrapErrorf(err, "Finding task '%d'", id)
	}

	return task, nil
}

func (c Client) TaskOutput(id int, type_ string, taskReporter TaskReporter) error {
	err := c.taskClientRequest.WaitForCompletion(id, type_, taskReporter)
	if err != nil {
		return bosherr.WrapErrorf(err, "Capturing task '%d' output", id)
	}

	return nil
}

func (c Client) CancelTask(id int) error {
	path := fmt.Sprintf("/task/%d", id)

	_, _, err := c.clientRequest.RawDelete(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Cancelling task '%d'", id)
	}

	return nil
}
