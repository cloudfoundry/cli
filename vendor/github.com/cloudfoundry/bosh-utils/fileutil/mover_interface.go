package fileutil

type Mover interface {
	Move(string, string) error
}
