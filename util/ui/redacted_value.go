package ui

import (
	"net/http"
	"regexp"
)

// RedactedValue is the text that is displayed for redacted content. (eg
// authorization tokens, passwords, etc.)
const RedactedValue = "[PRIVATE DATA HIDDEN]"

func RedactHeaders(header http.Header) http.Header {
	redactedHeaders := make(http.Header)
	re := regexp.MustCompile(`([&?]code)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
	for key, value := range header {
		if key == "Authorization" || key == "Set-Cookie" {
			redactedHeaders[key] = []string{RedactedValue}
		} else {
			for index, v := range value {
				value[index] = re.ReplaceAllString(v, "$1="+RedactedValue)
			}
			redactedHeaders[key] = value
		}
	}
	return redactedHeaders
}
