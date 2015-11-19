package net

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/onsi/ginkgo"
)

type TestRequest struct {
	Method   string
	Path     string
	Header   http.Header
	Matcher  RequestMatcher
	Response TestResponse
}

type RequestMatcher func(*http.Request)

type TestResponse struct {
	Body   string
	Status int
	Header http.Header
}

type TestHandler struct {
	Requests  []TestRequest
	CallCount int
}

func (h *TestHandler) AllRequestsCalled() bool {
	return h.CallCount == len(h.Requests)
}

func urlQueryContains(container, containee url.Values) bool {
	//Cloud Controller often uses "q" as a container for search queries, which may be semantically
	//equivalent to CC but be actually different strings.

	//Example: "foo:bar;baz:qux" is semantically the same as "baz:qux;foo:bar". CC doesn't care about order.

	//Therefore, we crack apart "q" params on their seperator (a colon) and compare the resulting
	//substrings.  No other params seem to use semicolon separators AND are order-dependent, so we just
	//run all params through the same process.
	for key := range containee {

		containerValues := strings.Split(container.Get(key), ";")
		containeeValues := strings.Split(containee.Get(key), ";")

		allValuesFound := make([]bool, len(containeeValues))

		for index, expected := range containeeValues {
			for _, actual := range containerValues {
				if expected == actual {
					allValuesFound[index] = true
					break
				}
			}
		}
		for _, ok := range allValuesFound {
			if !ok {
				return false
			}
		}
	}

	return true
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer ginkgo.GinkgoRecover()

	if len(h.Requests) <= h.CallCount {
		h.logError("Index out of range! Test server called too many times. Final Request:", r.Method, r.RequestURI)
		return
	}

	tester := h.Requests[h.CallCount]
	h.CallCount++

	// match method
	if tester.Method != r.Method {
		h.logError("Method does not match.\nExpected: %s\nActual:   %s", tester.Method, r.Method)
	}

	// match path
	paths := strings.Split(tester.Path, "?")
	if paths[0] != r.URL.Path {
		h.logError("Path does not match.\nExpected: %s\nActual:   %s", paths[0], r.URL.Path)
	}
	// match query string
	if len(paths) > 1 {
		actualValues, _ := url.ParseQuery(r.URL.RawQuery)
		expectedValues, _ := url.ParseQuery(paths[1])

		if !urlQueryContains(actualValues, expectedValues) {
			h.logError("Query string does not match.\nExpected: %s\nActual:   %s", paths[1], r.URL.RawQuery)
		}
	}

	for key, values := range tester.Header {
		key = http.CanonicalHeaderKey(key)
		actualValues := strings.Join(r.Header[key], ";")
		expectedValues := strings.Join(values, ";")

		if key == "Authorization" && !strings.Contains(actualValues, expectedValues) {
			h.logError("%s header is not contained in actual value.\nExpected: %s\nActual:   %s", key, expectedValues, actualValues)
		}
		if key != "Authorization" && actualValues != expectedValues {
			h.logError("%s header did not match.\nExpected: %s\nActual:   %s", key, expectedValues, actualValues)
		}
	}

	// match custom request matcher
	if tester.Matcher != nil {
		tester.Matcher(r)
	}

	// set response headers
	header := w.Header()
	for name, values := range tester.Response.Header {
		if len(values) < 1 {
			continue
		}
		header.Set(name, values[0])
	}

	// write response
	w.WriteHeader(tester.Response.Status)
	fmt.Fprintln(w, tester.Response.Body)
}

func NewTLSServer(requests []TestRequest) (*httptest.Server, *TestHandler) {
	handler := &TestHandler{Requests: requests}
	return httptest.NewTLSServer(handler), handler
}

func NewServer(requests []TestRequest) (*httptest.Server, *TestHandler) {
	handler := &TestHandler{Requests: requests}
	return httptest.NewServer(handler), handler
}

func (h *TestHandler) logError(msg string, args ...interface{}) {
	println(fmt.Sprintf(msg, args...))
	ginkgo.Fail("failed")
}
