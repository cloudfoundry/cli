package net

import (
	"cf/terminal"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
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

	if TraceEnabled() {
		dumpRequest(req)
	}

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

	if TraceEnabled() {
		dumpRequest(request)
	}

	response, err = httpClient.Do(request)

	if err != nil {
		return
	}

	if TraceEnabled() {
		dumpedResponse, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println("Error dumping response")
		} else {
			fmt.Printf("\n%s\n%s\n", terminal.HeaderColor("RESPONSE:"), Sanitize(string(dumpedResponse)))
		}
	}

	return
}

func TraceEnabled() bool {
	traceEnv := strings.ToLower(os.Getenv("CF_TRACE"))
	return traceEnv == "true" || traceEnv == "yes"
}

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		fmt.Println("Error dumping request")
	} else {
		fmt.Printf("\n%s\n%s\n", terminal.HeaderColor("REQUEST:"), Sanitize(string(dumpedRequest)))
		if !shouldDisplayBody {
			fmt.Println("[MULTIPART/FORM-DATA CONTENT HIDDEN]")
		}
	}
}
