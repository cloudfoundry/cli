package actionerror

type PackageProcessingFailedError struct{}

func (PackageProcessingFailedError) Error() string {
	return "Package failed to process correctly after upload"
}
