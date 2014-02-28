package net

import (
	"cf/terminal"
	"cf/trace"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

const (
	PRIVATE_DATA_PLACEHOLDER = "[PRIVATE DATA HIDDEN]"
)

func newHttpClient(trustedCerts []tls.Certificate) *http.Client {
	var certPool *x509.CertPool
	if len(trustedCerts) > 0 {
		certPool = x509.NewCertPool()
		for _, tlsCert := range trustedCerts {
			cert, _ := x509.ParseCertificate(tlsCert.Certificate[0])
			certPool.AddCert(cert)
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: certPool},
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

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf("Error dumping request\n%s\n", err)
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor("REQUEST:"), time.Now().Format(time.RFC3339), Sanitize(string(dumpedRequest)))
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
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor("RESPONSE:"), time.Now().Format(time.RFC3339), Sanitize(string(dumpedResponse)))
	}
}
