package translatableerror

import "strings"

type HealthCheckTypeUnsupportedError struct {
	SupportedTypes []string
}

func (HealthCheckTypeUnsupportedError) Error() string {
	return "Your target CF API version only supports health check type values {{.SupportedTypes}} and {{.LastSupportedType}}."
}

func (e HealthCheckTypeUnsupportedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"SupportedTypes":    strings.Join(e.SupportedTypes[:len(e.SupportedTypes)-1], ", "),
		"LastSupportedType": e.SupportedTypes[len(e.SupportedTypes)-1],
	})
}
