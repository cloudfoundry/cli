package translatableerror

type ManualClientCredentialsError struct{}

func (e ManualClientCredentialsError) Error() string {
	return "Error: Support for manually writing your client credentials to config.json has been removed. For similar functionality please use `cf auth --client-credentials`."
}

func (e ManualClientCredentialsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
