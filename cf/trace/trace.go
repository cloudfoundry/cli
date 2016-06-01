package trace

import (
	"fmt"
	"regexp"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

var LoggingToStdout bool

func Sanitize(input string) string {
	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized := re.ReplaceAllString(input, "Authorization: "+PrivateDataPlaceholder())

	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PrivateDataPlaceholder()+"&")

	sanitized = sanitizeJSON("access_token", sanitized)
	sanitized = sanitizeJSON("refresh_token", sanitized)
	sanitized = sanitizeJSON("token", sanitized)
	sanitized = sanitizeJSON("password", sanitized)
	sanitized = sanitizeJSON("oldPassword", sanitized)

	return sanitized
}

func sanitizeJSON(propertyName string, json string) string {
	regex := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"[^\,]*"`, propertyName))
	return regex.ReplaceAllString(json, fmt.Sprintf(`"%s":"%s"`, propertyName, PrivateDataPlaceholder()))
}

func PrivateDataPlaceholder() string {
	return T("[PRIVATE DATA HIDDEN]")
}
