package api

import "cf/net"

type FakeCurlRepository struct {
	Method         string
	Path           string
	Header         string
	Body           string
	ResponseHeader string
	ResponseBody   string
	ApiResponse    net.ApiResponse
}

func (repo *FakeCurlRepository) Request(method, path, header, body string) (resHeaders, resBody string, apiResponse net.ApiResponse) {
	repo.Method = method
	repo.Path = path
	repo.Header = header
	repo.Body = body

	resHeaders = repo.ResponseHeader
	resBody = repo.ResponseBody
	apiResponse = repo.ApiResponse
	return
}
