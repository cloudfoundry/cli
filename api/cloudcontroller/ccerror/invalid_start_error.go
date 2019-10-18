package ccerror

type InvalidStartError struct {
}

func (e InvalidStartError) Error() string {
	return "App can not start with out a package to stage or a droplet to run."
}
