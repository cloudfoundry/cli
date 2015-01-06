package net

import (
	_ "crypto/sha512"
	"crypto/x509"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type HttpClientInterface interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

var NewHttpClient = func(tr *http.Transport) HttpClientInterface {
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
	copyHeaders(prevReq, req)
	dumpRequest(req)

	return nil
}

func copyHeaders(from *http.Request, to *http.Request) {
	for key, values := range from.Header {
		// do not copy POST-specific headers
		if key != "Content-Type" && key != "Content-Length" {
			to.Header.Set(key, strings.Join(values, ","))
		}
	}
}

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequest(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf(T("Error dumping request\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("REQUEST:")), time.Now().Format(time.RFC3339), trace.Sanitize(string(dumpedRequest)))
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
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("RESPONSE:")), time.Now().Format(time.RFC3339), trace.Sanitize(string(dumpedResponse)))
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
		switch innerErr.(type) {
		case x509.UnknownAuthorityError:
			return errors.NewInvalidSSLCert(host, T("unknown authority"))
		case x509.HostnameError:
			return errors.NewInvalidSSLCert(host, T("not valid for the requested host"))
		case x509.CertificateInvalidError:
			return errors.NewInvalidSSLCert(host, "")
		}
	}

	return errors.NewWithError(T("Error performing request"), err)

}
