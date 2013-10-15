package net

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"fmt"
)

type TestRequest struct {
	Method   string
	Path     string
	Matcher  RequestMatcher
	Response TestResponse
}

type RequestMatcher func (*http.Request) bool

type TestResponse struct {
	Body   string
	Status int
	Header http.Header
}

type TestHandler struct {
	Requests  []TestRequest
	CallCount int
	T         *testing.T
}

func (h *TestHandler) AllRequestsCalled() bool {
	return h.CallCount == len(h.Requests)
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (len(h.Requests) <= h.CallCount) {
		h.logError("Index out of range! Test server called too many times. Final Request:",r.Method, r.RequestURI)
		return
	}

	tester := h.Requests[h.CallCount]
	h.CallCount++

	// match method
	if tester.Method != r.Method {
		h.logError("Method does not match.\nExpected: %s\nActual:   %s",tester.Method,r.Method)
	}

	// match path
	paths := strings.Split(tester.Path, "?")
	if paths[0] != r.URL.Path {
		h.logError("Path does not match.\nExpected: %s\nActual:   %s",paths[0],r.URL.Path)
	}
	// match query string
	if len(paths) > 1 {
		if !strings.Contains(r.RequestURI, paths[1]) {
			h.logError("Query string does not match.\nExpected: %s\nActual:   %s",paths[1],r.URL.Path)
		}
	}

	// match required headers
	if r.Header.Get("accept") != "application/json" {
		h.logError("Accept header did not match.\nExpected: application/json\nActual:   ",r.Header.Get("accept"))
	}
	if !strings.HasPrefix(r.Header.Get("authorization"), "BEARER my_access_token") {
		h.logError("Authorization header did not match.\nExpected: BEARER my_access_token\nActual:   ",r.Header.Get("authorization"))
	}

	// match custom request matcher
	if tester.Matcher != nil && !tester.Matcher(r){
		h.logError("Custom request matcher did not match")
	}

	// set response headers
	header := w.Header()
	for name, values := range tester.Response.Header {
		if (len(values) < 1) {
			continue
		}
		header.Set(name, values[0])
	}

	// write response
	w.WriteHeader(tester.Response.Status)
	fmt.Fprintln(w, tester.Response.Body)
}

func NewServer(t *testing.T, requests []TestRequest) ( s *httptest.Server, h *TestHandler) {
	h = &TestHandler{
		Requests:requests,
		T: t,
	}
	s = httptest.NewTLSServer(h)
	return
}

func (h *TestHandler) logError(msg string, args ...interface{}) {
	h.T.Logf(msg, args...)
	h.T.Fail()
}
