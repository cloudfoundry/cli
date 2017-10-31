package actionerror

type PackageProcessingExpiredError struct{}

func (PackageProcessingExpiredError) Error() string {
	return "Package expired after upload"
}
