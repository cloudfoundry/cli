package testutils

import (
	"fmt"
)

type TarVerifier struct {
	BlobPath string
}

func (b TarVerifier) FileExists(fileName string) (bool, error) {
	session, err := RunCommand("tar", "-tf", b.BlobPath, fmt.Sprintf("./%s", fileName))
	if err != nil {
		return false, err
	}
	return session.ExitCode() == 0, nil
}

func (b TarVerifier) FileContents(fileName string) ([]byte, error) {
	session, err := RunCommand("tar", "--to-stdout", "-xf", b.BlobPath, fmt.Sprintf("./%s", fileName))
	if err != nil {
		return []byte{}, err
	}

	return session.Out.Contents(), nil
}

func (b TarVerifier) Listing() (string, error) {
	session, err := RunCommand("tar", "-tf", b.BlobPath)
	if err != nil {
		return "", err
	}

	return string(session.Out.Contents()), nil
}
