package trace

import (
	"fmt"
	"regexp"

	"io"
	"os"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

var stdout io.Writer = os.Stdout

var LoggingToStdout bool

func SetStdout(s io.Writer) {
	stdout = s
}

func Sanitize(input string) string {
	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized := re.ReplaceAllString(input, "Authorization: "+PRIVATE_DATA_PLACEHOLDER())

	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PRIVATE_DATA_PLACEHOLDER()+"&")

	sanitized = sanitizeJson("access_token", sanitized)
	sanitized = sanitizeJson("refresh_token", sanitized)
	sanitized = sanitizeJson("token", sanitized)
	sanitized = sanitizeJson("password", sanitized)
	sanitized = sanitizeJson("oldPassword", sanitized)

	return sanitized
}

func sanitizeJson(propertyName string, json string) string {
	regex := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"[^\,]*"`, propertyName))
	return regex.ReplaceAllString(json, fmt.Sprintf(`"%s":"%s"`, propertyName, PRIVATE_DATA_PLACEHOLDER()))
}

func PRIVATE_DATA_PLACEHOLDER() string {
	return T("[PRIVATE DATA HIDDEN]")
}
