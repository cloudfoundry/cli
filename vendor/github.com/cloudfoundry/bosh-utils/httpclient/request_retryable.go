package httpclient

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"io"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type RequestRetryable struct {
	request   *http.Request
	requestID string
	delegate  Client
	attempt   int

	bodyBytes []byte // buffer request body to memory for retries
	response  *http.Response

	uuidGenerator         boshuuid.Generator
	seekableRequestBody   io.ReadCloser
	logger                boshlog.Logger
	logTag                string
	isResponseAttemptable func(*http.Response, error) (bool, error)
}

func NewRequestRetryable(
	request *http.Request,
	delegate Client,
	logger boshlog.Logger,
	isResponseAttemptable func(*http.Response, error) (bool, error),
) *RequestRetryable {
	if isResponseAttemptable == nil {
		isResponseAttemptable = defaultIsAttemptable
	}

	return &RequestRetryable{
		request:               request,
		delegate:              delegate,
		attempt:               0,
		uuidGenerator:         boshuuid.NewGenerator(),
		logger:                logger,
		logTag:                "clientRetryable",
		isResponseAttemptable: isResponseAttemptable,
	}
}

func (r *RequestRetryable) Attempt() (bool, error) {
	var err error

	if r.requestID == "" {
		r.requestID, err = r.uuidGenerator.Generate()
		if err != nil {
			return false, bosherr.WrapError(err, "Generating request uuid")
		}
	}

	_, implementsSeekable := r.request.Body.(io.ReadSeeker)
	if r.seekableRequestBody != nil || implementsSeekable {
		if r.seekableRequestBody == nil {
			r.seekableRequestBody = r.request.Body
		}

		seekable, ok := r.seekableRequestBody.(io.ReadSeeker)
		if !ok {
			return false, errors.New("Should never happen")
		}
		_, err = seekable.Seek(0, 0)
		r.request.Body = ioutil.NopCloser(seekable)

		if err != nil {
			return false, bosherr.WrapErrorf(err, "Seeking to begining of seekable request body during attempt %d", r.attempt)
		}
	} else {
		if r.request.Body != nil && r.bodyBytes == nil {
			defer r.request.Body.Close()
			r.bodyBytes, err = ioutil.ReadAll(r.request.Body)
			if err != nil {
				return false, bosherr.WrapError(err, "Buffering request body")
			}
		}

		// reset request body, because readers cannot be re-read
		if r.bodyBytes != nil {
			r.request.Body = ioutil.NopCloser(bytes.NewReader(r.bodyBytes))
		}
	}

	// close previous attempt's response body to prevent HTTP client resource leaks
	if r.response != nil {
		// net/http response body early closing does not block until the body is
		// properly cleaned up, which would lead to a 'request canceled' error.
		// Yielding the CPU should allow the scheduler to run the cleanup tasks
		// before continuing. But we found that that behavior is not deterministic,
		// we instead avoid the problem altogether by reading the entire body and
		// forcing an EOF.
		// This should not be necessary when the following CL gets accepted:
		// https://go-review.googlesource.com/c/go/+/62891
		io.Copy(ioutil.Discard, r.response.Body)

		r.response.Body.Close()
	}

	r.attempt++

	r.logger.Debug(r.logTag, "[requestID=%s] Requesting (attempt=%d): %s", r.requestID, r.attempt, formatRequest(r.request))
	r.response, err = r.delegate.Do(r.request)

	attemptable, err := r.isResponseAttemptable(r.response, err)
	if !attemptable && r.seekableRequestBody != nil {
		r.seekableRequestBody.Close()
	}

	return attemptable, err
}

func (r *RequestRetryable) Response() *http.Response {
	return r.response
}

func (r *RequestRetryable) wasSuccessful(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func defaultIsAttemptable(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false, nil
	}
	return true, bosherr.Errorf("Request failed, response: %s", formatResponse(resp))
}

func formatRequest(req *http.Request) string {
	if req == nil {
		return "Request(nil)"
	}

	return fmt.Sprintf("Request{ Method: '%s', URL: '%s' }", req.Method, req.URL)
}

func formatResponse(resp *http.Response) string {
	if resp == nil {
		return "Response(nil)"
	}

	return fmt.Sprintf("Response{ StatusCode: %d, Status: '%s' }", resp.StatusCode, resp.Status)
}
