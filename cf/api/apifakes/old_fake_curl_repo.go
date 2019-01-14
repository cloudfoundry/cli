package apifakes

type OldFakeCurlRepository struct {
	Method          string
	Path            string
	Header          string
	Body            string
	ResponseHeader  string
	ResponseBody    string
	FailOnHTTPError bool
	Error           error
}

func (repo *OldFakeCurlRepository) Request(method, path, header, body string, failOnHTTPError bool) (resHeaders, resBody string, apiErr error) {
	repo.Method = method
	repo.Path = path
	repo.Header = header
	repo.Body = body
	repo.FailOnHTTPError = failOnHTTPError

	resHeaders = repo.ResponseHeader
	resBody = repo.ResponseBody
	apiErr = repo.Error
	return
}
