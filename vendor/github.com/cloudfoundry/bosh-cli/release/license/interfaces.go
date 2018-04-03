package license

//go:generate counterfeiter . DirReader

type DirReader interface {
	Read(string) (*License, error)
}
