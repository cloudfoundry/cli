package director

import (
	"net/http"
)

type RequestSanitizer struct {
	Request http.Request
}

// This will destructively mutate rs.Request
func (rs RequestSanitizer) SanitizeRequest() (http.Request, error) {
	rs.sanitizeAuthorization()

	return rs.Request, nil
}

func (rs RequestSanitizer) sanitizeAuthorization() {
	if rs.Request.Header.Get("Authorization") != "" {
		rs.Request.Header.Del("Authorization")
		rs.Request.Header.Add("Authorization", "[removed]")
	}

	return
}
