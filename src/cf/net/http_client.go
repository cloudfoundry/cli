package net

import (
	"cf/terminal"
	"cf/trace"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

const (
	PRIVATE_DATA_PLACEHOLDER = "[PRIVATE DATA HIDDEN]"
)

func newHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}
	return &http.Client{
		Transport:     tr,
		CheckRedirect: PrepareRedirect,
	}
}

func PrepareRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 1 {
		return errors.New("stopped after 1 redirect")
	}

	prevReq := via[len(via)-1]

	req.Header.Set("Authorization", prevReq.Header.Get("Authorization"))

	dumpRequest(req)

	return nil
}

func Sanitize(input string) (sanitized string) {
	var sanitizeJson = func(propertyName string, json string) string {
		re := regexp.MustCompile(fmt.Sprintf(`"%s":"[^"]*"`, propertyName))
		return re.ReplaceAllString(json, fmt.Sprintf(`"%s":"`+PRIVATE_DATA_PLACEHOLDER+`"`, propertyName))
	}

	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized = re.ReplaceAllString(input, "Authorization: "+PRIVATE_DATA_PLACEHOLDER)
	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PRIVATE_DATA_PLACEHOLDER+"&")

	sanitized = sanitizeJson("access_token", sanitized)
	sanitized = sanitizeJson("refresh_token", sanitized)
	sanitized = sanitizeJson("token", sanitized)

	return
}

func doRequest(request *http.Request) (response *http.Response, err error) {
	httpClient := newHttpClient()

	dumpRequest(request)

	response, err = httpClient.Do(request)
	if err != nil {
		return
	}

	dumpResponse(response)
	return
}

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf("Error dumping request\n%s\n", err)
	} else {
		trace.Logger.Printf("\n%s\n%s\n", terminal.HeaderColor("REQUEST:"), Sanitize(string(dumpedRequest)))
		if !shouldDisplayBody {
			trace.Logger.Println("[MULTIPART/FORM-DATA CONTENT HIDDEN]")
		}
	}
}

func dumpResponse(res *http.Response) {
	dumpedResponse, err := httputil.DumpResponse(res, true)
	if err != nil {
		trace.Logger.Printf("Error dumping response\n%s\n", err)
	} else {
		trace.Logger.Printf("\n%s\n%s\n", terminal.HeaderColor("RESPONSE:"), Sanitize(string(dumpedResponse)))
	}
}
