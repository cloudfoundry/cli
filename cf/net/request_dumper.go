package net

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	. "code.cloudfoundry.org/cli/v8/cf/i18n"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
	"code.cloudfoundry.org/cli/v8/cf/trace"
	"code.cloudfoundry.org/cli/v8/util/ui"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RequestDumperInterface

type RequestDumperInterface interface {
	DumpRequest(*http.Request)
	DumpResponse(*http.Response)
}

type RequestDumper struct {
	printer trace.Printer
}

func NewRequestDumper(printer trace.Printer) RequestDumper {
	return RequestDumper{printer: printer}
}

func (p RequestDumper) DumpRequest(req *http.Request) {
	p.printer.Printf("\n%s [%s]\n", terminal.HeaderColor(T("REQUEST:")), time.Now().Format(time.RFC3339))

	re := regexp.MustCompile(`([&?]code)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
	redactedURI := re.ReplaceAllString(req.URL.RequestURI(), "$1="+ui.RedactedValue)

	p.printer.Printf("%s %s %s\n", req.Method, redactedURI, req.Proto)

	p.printer.Printf("Host: %s", req.URL.Host)

	headers := ui.RedactHeaders(req.Header)
	p.displaySortedHeaders(headers)

	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		p.printer.Println(T("[MULTIPART/FORM-DATA CONTENT HIDDEN]"))
	}

	if req.Body == nil {
		return
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	if len(requestBody) == 0 {
		return
	}

	if strings.Contains(contentType, "application/json") {
		if err != nil {
			p.printer.Println("Unable to read Request Body:", err)
			return
		}

		sanitizedJSON, err := ui.SanitizeJSON(requestBody)

		if err != nil {
			p.printer.Println("Failed to sanitize json body:", err)
			return
		}

		p.printer.Printf("%s\n", sanitizedJSON)
	}
	if strings.Contains(contentType, "x-www-form-urlencoded") {

		formData, err := url.ParseQuery(string(requestBody))
		if err != nil {
			p.printer.Println("Failed to parse form:", err)
			return
		}

		redactedData := p.redactFormData(formData)
		p.displayFormData(redactedData)
	}
}

func (p RequestDumper) displaySortedHeaders(headers http.Header) {
	sortedHeaders := []string{}
	for key, _ := range headers {
		sortedHeaders = append(sortedHeaders, key)
	}
	sort.Strings(sortedHeaders)

	for _, header := range sortedHeaders {
		for _, value := range headers[header] {
			p.printer.Printf("%s: %s\n", T(header), value)
		}
	}
	p.printer.Printf("\n")
}

func (p RequestDumper) redactFormData(formData url.Values) url.Values {
	for key := range formData {
		if key == "password" || key == "Authorization" || strings.Contains(key, "token") {
			formData.Set(key, ui.RedactedValue)
		}
	}
	return formData
}

func (p RequestDumper) displayFormData(formData url.Values) {
	var buf strings.Builder
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := formData[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	p.printer.Printf("%s\n", buf.String())
}

func (p RequestDumper) DumpResponse(res *http.Response) {
	dumpedResponse, err := httputil.DumpResponse(res, true)
	if err != nil {
		p.printer.Printf(T("Error dumping response\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		p.printer.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("RESPONSE:")), time.Now().Format(time.RFC3339), trace.Sanitize(string(dumpedResponse)))
	}
}
