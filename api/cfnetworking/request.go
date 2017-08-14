package cfnetworking

import (
	"io"
	"net/http"
)

//go:generate counterfeiter . ReadSeeker

type ReadSeeker interface {
	io.ReadSeeker
}

// Request represents the request of the cloud controller.
type Request struct {
	*http.Request

	body io.ReadSeeker
}

func (r *Request) ResetBody() error {
	if r.body == nil {
		return nil
	}

	_, err := r.body.Seek(0, 0)
	return err
}

func NewRequest(request *http.Request, body io.ReadSeeker) *Request {
	return &Request{
		Request: request,
		body:    body,
	}
}
