package fileutils

import (
	"io"
	"os"
)

func CopyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}
