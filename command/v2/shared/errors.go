package shared

import (
	"fmt"
	"time"
)

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

type NoOrganizationTargetedError struct{}

func (e NoOrganizationTargetedError) Error() string {
	return "An org must be targeted before targeting a space"
}

func (e NoOrganizationTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

type OrganizationNotFoundError struct {
	Name string
}

func (e OrganizationNotFoundError) Error() string {
	return "Organization '{{.Name}}' not found."
}

func (e OrganizationNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}

type SpaceNotFoundError struct {
	Name string
}

func (e SpaceNotFoundError) Error() string {
	return "Space '{{.Name}}' not found."
}

func (e SpaceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}

type HTTPHealthCheckInvalidError struct {
}

func (e HTTPHealthCheckInvalidError) Error() string {
	return "Health check type must be 'http' to set a health check HTTP endpoint."
}

func (e HTTPHealthCheckInvalidError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

type InvalidRefreshTokenError struct {
}

func (e InvalidRefreshTokenError) Error() string {
	return "The token expired, was revoked, or the token ID is incorrect. Please log back in to re-authenticate."
}

func (e InvalidRefreshTokenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

type StagingFailedError struct {
	Message    string
	BinaryName string
}

func (e StagingFailedError) Error() string {
	return `{{.Message}}\n\nTIP: Use '{{.BuildpackCommand}}' to see a list of supported buildpacks.`
}

func (e StagingFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message":          e.Message,
		"BuildpackCommand": fmt.Sprintf("%s buildpacks", e.BinaryName),
	})
}
