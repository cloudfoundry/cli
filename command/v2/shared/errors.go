package shared

import "time"

type JobFailedError struct {
	JobGUID string
	Message string
}

func (e JobFailedError) Error() string {
	return "Job ({{.JobGUID}}) failed: {{.Message}}"
}

func (e JobFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
		"JobGUID": e.JobGUID,
	})
}

type JobTimeoutError struct {
	JobGUID string
	Timeout time.Duration
}

func (e JobTimeoutError) Error() string {
	return "Job ({{.JobGUID}}) polling timeout has been reached. The operation may still be running on the CF instance. Your CF operator may have more information."
}

func (e JobTimeoutError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"JobGUID": e.JobGUID,
	})
}
