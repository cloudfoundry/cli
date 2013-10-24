package cf

import (
	"os"
)

type FakeZipper struct {
	ZippedDir string
	ZippedFile *os.File
}

func (zipper *FakeZipper) Zip(dir string) (zipFile *os.File, err error) {
	zipper.ZippedDir = dir
	return zipper.ZippedFile, nil
}
