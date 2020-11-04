package download

import "net/http"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HTTPClient

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}
