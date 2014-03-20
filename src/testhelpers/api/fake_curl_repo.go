package api

type FakeCurlRepository struct {
	Method         string
	Path           string
	Header         string
	Body           string
	ResponseHeader string
	ResponseBody   string
	Error          error
}

func (repo *FakeCurlRepository) Request(method, path, header, body string) (resHeaders, resBody string, apiErr error) {
	repo.Method = method
	repo.Path = path
	repo.Header = header
	repo.Body = body

	resHeaders = repo.ResponseHeader
	resBody = repo.ResponseBody
	apiErr = repo.Error
	return
}
