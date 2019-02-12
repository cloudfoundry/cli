package ui

import "net/http"

// RedactedValue is the text that is displayed for redacted content. (eg
// authorization tokens, passwords, etc.)
const RedactedValue = "[PRIVATE DATA HIDDEN]"

func RedactHeaders(header http.Header) http.Header {
	redactedHeaders := make(http.Header)
	for key, value := range header {
		if key == "Authorization" {
			redactedHeaders[key] = []string{RedactedValue}
		} else {
			redactedHeaders[key] = value
		}
	}
	return redactedHeaders
}
