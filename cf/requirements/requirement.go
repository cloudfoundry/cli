package requirements

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Requirement

type Requirement interface {
	Execute() error
}
