package manifest

type InheritanceFieldError struct{}

func (InheritanceFieldError) Error() string {
	return "unsupported inheritance"
}
