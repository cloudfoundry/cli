package download

import "net/http"

//go:generate counterfeiter . HTTPClient

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}
