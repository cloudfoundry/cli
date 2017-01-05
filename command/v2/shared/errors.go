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

type CurrentUserError struct {
	Message string
}

func (e CurrentUserError) Error() string {
	return "Error retrieving current user:\n{{.Message}}"
}

func (e CurrentUserError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}

type OrgTargetError struct {
	Message string
}

func (e OrgTargetError) Error() string {
	return "Could not target org.\n{{.APIErr}}"
}

func (e OrgTargetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"APIErr": e.Message,
	})
}

type NoOrgTargetedError struct {
	Message string
}

func (e NoOrgTargetedError) Error() string {
	return "An org must be targeted before targeting a space"
}

func (e NoOrgTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{})
}

type GetOrgSpacesError struct {
	Message string
}

func (e GetOrgSpacesError) Error() string {
	return "Error getting spaces in organization.\n{{.APIErr}}"
}

func (e GetOrgSpacesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"APIErr": e.Message,
	})
}

func (e SpaceTargetError) Error() string {
	return "Unable to access space {{.SpaceName}}.\n{{.APIErr}}"
}

func (e SpaceTargetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"APIErr":    e.Message,
		"SpaceName": e.SpaceName,
	})
}

type SpaceTargetError struct {
	Message   string
	SpaceName string
}
