package translatableerror

import "time"

type StagingTimeoutError struct {
	AppName string
	Timeout time.Duration
}

func (_ StagingTimeoutError) Error() string {
	return "{{.AppName}} failed to stage within {{.Timeout}} minutes"
}

func (e StagingTimeoutError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.AppName,
		"Timeout": e.Timeout.Minutes(),
	})
}
