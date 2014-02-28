package api

import "cf/errors"

type FakeCurlRepository struct {
	Method         string
	Path           string
	Header         string
	Body           string
	ResponseHeader string
	ResponseBody   string
	ApiResponse    errors.Error
}

func (repo *FakeCurlRepository) Request(method, path, header, body string) (resHeaders, resBody string, apiResponse errors.Error) {
	repo.Method = method
	repo.Path = path
	repo.Header = header
	repo.Body = body

	resHeaders = repo.ResponseHeader
	resBody = repo.ResponseBody
	apiResponse = repo.ApiResponse
	return
}
