package test_helpers

import (
	"net/http"
	"sync"
)

type FakeHandler struct {
	GenerateHandler func(chan []byte) http.Handler
	InputChan       chan []byte
	Fail            bool
	ContentLen      string
	called          bool
	lastURL         string
	authHeader      string
	sync.RWMutex
}

func (fh *FakeHandler) GetAuthHeader() string {
	fh.RLock()
	defer fh.RUnlock()
	return fh.authHeader
}

func (fh *FakeHandler) SetAuthHeader(authHeader string) {
	fh.Lock()
	defer fh.Unlock()
	fh.authHeader = authHeader
}

func (fh *FakeHandler) GetLastURL() string {
	fh.RLock()
	defer fh.RUnlock()
	return fh.lastURL
}

func (fh *FakeHandler) SetLastURL(url string) {
	fh.Lock()
	defer fh.Unlock()
	fh.lastURL = url
}

func (fh *FakeHandler) Call() {
	fh.Lock()
	defer fh.Unlock()
	fh.called = true
}

func (fh *FakeHandler) WasCalled() bool {
	fh.RLock()
	defer fh.RUnlock()
	return fh.called
}

func (fh *FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fh.SetLastURL(r.URL.String())
	fh.SetAuthHeader(r.Header.Get("Authorization"))
	fh.Call()
	if len(fh.ContentLen) > 0 {
		rw.Header().Set("Content-Length", fh.ContentLen)
	}

	if fh.Fail {
		return
	}

	fh.RLock()
	defer fh.RUnlock()
	handler := fh.GenerateHandler(fh.InputChan)
	handler.ServeHTTP(rw, r)
}

func (fh *FakeHandler) Close() {
	close(fh.InputChan)
}

func (fh *FakeHandler) Reset() {
	fh.Lock()
	defer fh.Unlock()

	fh.InputChan = make(chan []byte)
	fh.called = false
}
