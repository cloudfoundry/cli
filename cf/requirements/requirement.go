package requirements

//go:generate counterfeiter . Requirement

type Requirement interface {
	Execute() error
}
