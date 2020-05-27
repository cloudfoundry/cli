package trace

import (
	"fmt"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

var LoggingToStdout bool

func Sanitize(input string) string {
	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized := re.ReplaceAllString(input, "Authorization: "+PrivateDataPlaceholder())

	re = regexp.MustCompile(`(?m)^Set-Cookie: .*`)
	sanitized = re.ReplaceAllString(sanitized, "Set-Cookie: "+PrivateDataPlaceholder())

	// allow query parameter to contain all characters of the "query" character class, except for &
	// https://tools.ietf.org/html/rfc3986#appendix-A
	re = regexp.MustCompile(`([&?]password)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
	sanitized = re.ReplaceAllString(sanitized, "$1="+PrivateDataPlaceholder())

	re = regexp.MustCompile(`([&?]code)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
	sanitized = re.ReplaceAllString(sanitized, "$1="+PrivateDataPlaceholder())

	sanitized = sanitizeJSON("token", sanitized)
	sanitized = sanitizeJSON("password", sanitized)

	return sanitized
}

func sanitizeJSON(propertySubstring string, json string) string {
	regex := regexp.MustCompile(fmt.Sprintf(`(?i)"([^"]*%s[^"]*)":\s*"[^\,]*"`, propertySubstring))
	return regex.ReplaceAllString(json, fmt.Sprintf(`"$1":"%s"`, PrivateDataPlaceholder()))
}

func PrivateDataPlaceholder() string {
	return T("[PRIVATE DATA HIDDEN]")
}
