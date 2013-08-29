package testhelpers

import "bytes"

type FakeZipper struct {
	ZippedDir string
	ZippedBuffer *bytes.Buffer
}

func (zipper *FakeZipper) Zip(dir string) (zipBuffer *bytes.Buffer, err error) {
	zipper.ZippedDir = dir
	return zipper.ZippedBuffer, nil
}
