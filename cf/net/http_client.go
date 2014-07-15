package net

import (
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/i18n"
)

var (
	t                        = i18n.Init()
	PRIVATE_DATA_PLACEHOLDER = t("[PRIVATE DATA HIDDEN]")
)

func newHttpClient(trustedCerts []tls.Certificate, disableSSL bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: NewTLSConfig(trustedCerts, disableSSL),
		Proxy:           http.ProxyFromEnvironment,
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: PrepareRedirect,
	}
}

func PrepareRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 1 {
		return errors.New(T("stopped after 1 redirect"))
	}

	prevReq := via[len(via)-1]
	req.Header.Set("Authorization", prevReq.Header.Get("Authorization"))
	dumpRequest(req)

	return nil
}

func Sanitize(input string) (sanitized string) {
	var sanitizeJson = func(propertyName string, json string) string {
		regex := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"[^"]*"`, propertyName))
		return regex.ReplaceAllString(json, fmt.Sprintf(`"%s":"%s"`, propertyName, PRIVATE_DATA_PLACEHOLDER))
	}

	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized = re.ReplaceAllString(input, "Authorization: "+PRIVATE_DATA_PLACEHOLDER)
	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PRIVATE_DATA_PLACEHOLDER+"&")

	sanitized = sanitizeJson("access_token", sanitized)
	sanitized = sanitizeJson("refresh_token", sanitized)
	sanitized = sanitizeJson("token", sanitized)
	sanitized = sanitizeJson("password", sanitized)
	sanitized = sanitizeJson("oldPassword", sanitized)

	return
}

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf(T("Error dumping request\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("REQUEST:")), time.Now().Format(time.RFC3339), Sanitize(string(dumpedRequest)))
		if !shouldDisplayBody {
			trace.Logger.Println(T("[MULTIPART/FORM-DATA CONTENT HIDDEN]"))
		}
	}
}

func dumpResponse(res *http.Response) {
	dumpedResponse, err := httputil.DumpResponse(res, true)
	if err != nil {
		trace.Logger.Printf(T("Error dumping response\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("RESPONSE:")), time.Now().Format(time.RFC3339), Sanitize(string(dumpedResponse)))
	}
}

func WrapNetworkErrors(host string, err error) error {
	var innerErr error
	switch typedErr := err.(type) {
	case *url.Error:
		innerErr = typedErr.Err
	case *websocket.DialError:
		innerErr = typedErr.Err
	}

	if innerErr != nil {
		switch typedErr := innerErr.(type) {
		case x509.UnknownAuthorityError:
			return errors.NewInvalidSSLCert(host, T("unknown authority"))
		case x509.HostnameError:
			return errors.NewInvalidSSLCert(host, T("not valid for the requested host"))
		case x509.CertificateInvalidError:
			return errors.NewInvalidSSLCert(host, "")
		case *net.OpError:
			return typedErr.Err
		}
	}

	return errors.NewWithError(T("Error performing request"), err)

}
