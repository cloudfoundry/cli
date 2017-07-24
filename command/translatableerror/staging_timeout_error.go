package translatableerror

import "time"

type StagingTimeoutError struct {
	AppName string
	Timeout time.Duration
}

func (StagingTimeoutError) Error() string {
	return `Error staging application {{.AppName}}: timed out after {{.Timeout}} {{if eq .Timeout 1.0}}minute{{else}}minutes{{end}}`
}

func (e StagingTimeoutError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.AppName,
		"Timeout": e.Timeout.Minutes(),
	})
}
