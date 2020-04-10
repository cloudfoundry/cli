package ccerror

type InvalidStartError struct {
}

func (e InvalidStartError) Error() string {
	return "App cannot start without a package to stage or a droplet to run."
}
