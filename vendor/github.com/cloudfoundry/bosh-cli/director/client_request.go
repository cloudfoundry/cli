package director

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ClientRequest struct {
	endpoint     string
	contextId    string
	httpClient   *httpclient.HTTPClient
	fileReporter FileReporter
	logger       boshlog.Logger
}

func NewClientRequest(
	endpoint string,
	httpClient *httpclient.HTTPClient,
	fileReporter FileReporter,
	logger boshlog.Logger,
) ClientRequest {
	return ClientRequest{
		endpoint:     endpoint,
		httpClient:   httpClient,
		fileReporter: fileReporter,
		logger:       logger,
	}
}
func (r ClientRequest) WithContext(contextId string) ClientRequest {
	// returns a copy of the ClientRequest
	r.contextId = contextId
	return r
}

func (r ClientRequest) Get(path string, response interface{}) error {
	respBody, _, err := r.RawGet(path, nil, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return nil
}

func (r ClientRequest) Post(path string, payload []byte, f func(*http.Request), response interface{}) error {
	respBody, _, err := r.RawPost(path, payload, f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return nil
}

func (r ClientRequest) Put(path string, payload []byte, f func(*http.Request), response interface{}) error {
	respBody, _, err := r.RawPut(path, payload, f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return nil
}

func (r ClientRequest) Delete(path string, response interface{}) error {
	respBody, _, err := r.RawDelete(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return nil
}

func (r ClientRequest) RawGet(path string, out io.Writer, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(f)

	resp, err := r.httpClient.GetCustomized(url, wrapperFunc)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Performing request GET '%s'", url)
	}

	return r.readResponse(resp, out)
}

// RawPost follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawPost(path string, payload []byte, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := func(req *http.Request) {
		if f != nil {
			f(req)
		}

		isArchive := req.Header.Get("content-type") == "application/x-compressed"

		if isArchive && req.ContentLength > 0 && req.Body != nil {
			req.Body = r.fileReporter.TrackUpload(req.ContentLength, req.Body)
		}
	}

	wrapperFunc = r.setContextIDHeader(wrapperFunc)

	resp, err := r.httpClient.PostCustomized(url, payload, wrapperFunc)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Performing request POST '%s'", url)
	}

	return r.optionallyFollowResponse(url, resp)
}

// RawPut follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawPut(path string, payload []byte, f func(*http.Request)) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(f)

	resp, err := r.httpClient.PutCustomized(url, payload, wrapperFunc)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Performing request PUT '%s'", url)
	}

	return r.optionallyFollowResponse(url, resp)
}

// RawDelete follows redirects via GET unlike generic HTTP clients
func (r ClientRequest) RawDelete(path string) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	wrapperFunc := r.setContextIDHeader(nil)

	resp, err := r.httpClient.DeleteCustomized(url, wrapperFunc)
	if err != nil {
		return nil, nil, bosherr.WrapErrorf(err, "Performing request DELETE '%s'", url)
	}

	return r.optionallyFollowResponse(url, resp)
}

func (r ClientRequest) setContextIDHeader(f func(*http.Request)) func(*http.Request) {
	return func(req *http.Request) {
		if f != nil {
			f(req)
		}
		if r.contextId != "" {
			req.Header.Set("X-Bosh-Context-Id", r.contextId)
		}
	}
}

func (r ClientRequest) optionallyFollowResponse(url string, resp *http.Response) ([]byte, *http.Response, error) {
	body, resp, err := r.readResponse(resp, nil)
	if err != nil {
		return body, resp, err
	}

	// Follow redirect via GET
	if resp != nil && resp.StatusCode == http.StatusFound {
		redirectURL, err := resp.Location()
		if err != nil || redirectURL == nil {
			return body, resp, bosherr.WrapErrorf(
				err, "Getting Location header from POST '%s'", url)
		}

		return r.RawGet(redirectURL.Path, nil, nil)
	}

	return body, resp, nil
}

type ShouldTrackDownload interface {
	ShouldTrackDownload() bool
}

func (r ClientRequest) readResponse(resp *http.Response, out io.Writer) ([]byte, *http.Response, error) {
	defer resp.Body.Close()

	logTag := "director.clientRequest"

	var respBody []byte

	if out == nil {
		if resp.Request != nil {
			sanitizer := RequestSanitizer{Request: (*resp.Request)}
			sanitizedRequest, _ := sanitizer.SanitizeRequest()
			b, err := httputil.DumpRequest(&sanitizedRequest, true)
			if err == nil {
				r.logger.Debug(logTag, "Dumping Director client request:\n%s", string(b))
			}
		}

		b, err := httputil.DumpResponse(resp, true)
		if err == nil {
			r.logger.Debug(logTag, "Dumping Director client response:\n%s", string(b))
		}

		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, bosherr.WrapError(err, "Reading Director response")
		}
	}

	not200 := resp.StatusCode != http.StatusOK
	not201 := resp.StatusCode != http.StatusCreated
	not204 := resp.StatusCode != http.StatusNoContent
	not206 := resp.StatusCode != http.StatusPartialContent
	not302 := resp.StatusCode != http.StatusFound

	if not200 && not201 && not204 && not206 && not302 {
		msg := "Director responded with non-successful status code '%d' response '%s'"
		return nil, resp, bosherr.Errorf(msg, resp.StatusCode, respBody)
	}

	if out != nil {
		showProgress := true

		if typedOut, ok := out.(ShouldTrackDownload); ok {
			showProgress = typedOut.ShouldTrackDownload()
		}

		if showProgress {
			out = r.fileReporter.TrackDownload(resp.ContentLength, out)
		}

		_, err := io.Copy(out, resp.Body)
		if err != nil {
			return nil, nil, bosherr.WrapError(err, "Copying Director response")
		}
	}

	return respBody, resp, nil
}
